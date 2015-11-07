// Copyright 2015 Simon HEGE. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/*
Package gpkg provides a GeoPackage compliant database manipulation helpers.

An sqlite3 driver should be imported in main package, for example "github.com/mattn/go-sqlite3"

GeoPackage is an open, standards-based, platform-independent, portable, self-describing, compact format for transferring geospatial information.
It is defined in OGC 12-128r11. Additional resources can be found at http://www.geopackage.org/
*/
package gpkg

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg" //Register image decoder
	_ "image/png"  //Register image decoder

	"github.com/xeonx/geom"
	"github.com/xeonx/geom/encoding/geojson"
	"github.com/xeonx/geom/encoding/wkb"
	"github.com/xeonx/raster"
)

//Handle is a database handle to a GeoPackage. It's safe for concurrent use by multiple goroutines.
type Handle struct {
	db *sql.DB
}

//Open opens a GeoPackage (sqlite database) at the given location.
func Open(filepath string) (*Handle, error) {
	//Open database
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}
	h := &Handle{db: db}

	return h, nil
}

//Close closes the underlying sqlite connexion
func (h *Handle) Close() error {
	if h.db != nil {
		if err := h.db.Close(); err != nil {
			return err
		}
	}
	return nil
}

type feature struct {
	Type       string                 `json:"type"`
	ID         interface{}            `json:"id,omitempty"`
	Geometry   interface{}            `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
	BBOX       []float64              `json:"bbox,omitempty"`
}

func decodeGeometry(b []byte) (geom.Geometry, []float64, error) {

	buffer := bytes.NewBuffer(b)
	var magic uint16
	var version uint8
	var flags uint8
	var srsID uint32
	var byteOrder binary.ByteOrder = binary.BigEndian
	binary.Read(buffer, byteOrder, &magic)
	binary.Read(buffer, byteOrder, &version)
	binary.Read(buffer, byteOrder, &flags)
	if (flags & 0x01) == 0x01 {
		byteOrder = binary.LittleEndian
	}
	binary.Read(buffer, byteOrder, &srsID)

	//emptyGeometryFlag := ((flags >> 4) & 0x01) == 0x01
	envelopeFlag := ((flags >> 1) & 0x07)

	var bbox []float64
	if envelopeFlag > 0 {
		var xmin, xmax, ymin, ymax float64
		binary.Read(buffer, byteOrder, &xmin)
		binary.Read(buffer, byteOrder, &xmax)
		binary.Read(buffer, byteOrder, &ymin)
		binary.Read(buffer, byteOrder, &ymax)
		switch envelopeFlag {
		case 1:
			bbox = []float64{xmin, ymin, xmax, ymax}
		case 2:
			var zmin, zmax float64
			binary.Read(buffer, byteOrder, &zmin)
			binary.Read(buffer, byteOrder, &zmax)
			bbox = []float64{xmin, ymin, zmin, xmax, ymax, zmax}
		case 3:
			var mmin, mmax float64
			binary.Read(buffer, byteOrder, &mmin)
			binary.Read(buffer, byteOrder, &mmax)
			bbox = []float64{xmin, ymin, mmin, xmax, ymax, mmax}
		case 4:
			var zmin, zmax, mmin, mmax float64
			binary.Read(buffer, byteOrder, &zmin)
			binary.Read(buffer, byteOrder, &zmax)
			binary.Read(buffer, byteOrder, &mmin)
			binary.Read(buffer, byteOrder, &mmax)
			bbox = []float64{xmin, ymin, zmin, mmin, xmax, ymax, zmax, mmax}
		}
	}

	geom, err := wkb.Read(buffer)
	if err != nil {
		return nil, nil, err
	}

	return geom, bbox, nil
}

//ListFeatures retrieves the list of all items in the given feature GeoPackage table
func (h *Handle) ListFeatures(tableName string) (interface{}, error) {

	gColumns, err := h.ListGeometryColumnsForTable(tableName)
	if err != nil {
		return nil, err
	}
	geomColumns := make(map[string]*GeometryColumns)
	for _, gc := range gColumns {
		geomColumns[gc.ColumnName] = gc
	}

	rows, err := h.db.Query(fmt.Sprintf("SELECT * FROM %s", tableName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result []feature

	valuesPtr := make([]interface{}, len(columns))
	for i := range columns {
		var temp interface{}
		valuesPtr[i] = &temp
	}

	for rows.Next() {

		err := rows.Scan(valuesPtr...)
		if err != nil {
			return nil, err
		}

		var f feature
		f.Properties = make(map[string]interface{})
		f.Type = "Feature"
		for i, c := range columns {

			value, ok := valuesPtr[i].(*interface{})
			if !ok {
				return nil, errors.New("Invalid database type")
			}

			if c == "id" || c == "fid" {
				f.ID = *value
				continue
			}

			if _, isGeom := geomColumns[c]; isGeom {

				if *value != nil {

					b, ok := (*value).([]byte)
					if !ok {
						return nil, errors.New("Invalid database type")
					}

					var geom geom.Geometry
					geom, f.BBOX, err = decodeGeometry(b)
					if err != nil {
						return nil, err
					}
					f.Geometry, err = geojson.ToGeoJSON(geom)
					if err != nil {
						return nil, err
					}
				}
				continue

			}

			f.Properties[c] = *value
		}

		result = append(result, f)
	}

	return struct {
		Type     string    `json:"type"`
		Features []feature `json:"features"`
	}{
		Type:     "FeatureCollection",
		Features: result,
	}, nil //, rows.Err()
}

//GetTile retrieves a single tile in a GeoPackage tiles table
func (h *Handle) GetTile(tableName string, level, x, y int64) (image.Image, error) {

	var data []byte
	err := h.db.QueryRow(fmt.Sprintf("SELECT tile_data FROM %s WHERE zoom_level=? AND tile_column=? AND tile_row=?", tableName), level, x, y).Scan(&data)
	if err != nil {
		return nil, err
	}
	buffer := bytes.NewBuffer(data)

	image, _, err := image.Decode(buffer)
	if err != nil {
		return nil, err
	}

	return image, nil
}

func (h *Handle) hasTable(tableName string) bool {
	var count int
	err := h.db.QueryRow("SELECT count(*) FROM sqlite_master WHERE name=?", tableName).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

//HasExtensionsTable tests if the GeoPackage contains an extension table
func (h *Handle) HasExtensionsTable() bool {
	return h.hasTable("gpkg_extensions")
}

//HasFeatures tests if the GeoPackage contains features tables
func (h *Handle) HasFeatures() bool {
	var count int
	err := h.db.QueryRow("SELECT count(*) FROM gpkg_contents WHERE data_type='features'").Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

//HasTiles tests if the GeoPackage contains tiles tables
func (h *Handle) HasTiles() bool {
	var count int
	err := h.db.QueryRow("SELECT count(*) FROM gpkg_contents WHERE data_type='tiles'").Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

type sqlExecer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type sqlQueryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

type tileContent struct {
	h    *Handle
	name string
}

//TileFormat exposes the image format of the source (png or jpg)
func (t tileContent) TileFormat() string {
	//Look at the first tile
	var data []byte
	err := t.h.db.QueryRow(fmt.Sprintf("SELECT tile_data FROM %s LIMIT 1", t.name)).Scan(&data)
	if err == nil {

		buffer := bytes.NewBuffer(data)

		_, format, err := image.Decode(buffer)
		if err == nil {
			return format
		}

	}
	//Default is jpg
	return "jpg"
}

//GetRaw retrieves the tile for a given level/x/y.
func (t tileContent) GetRaw(level, x, y int) ([]byte, error) {

	var data []byte
	err := t.h.db.QueryRow(fmt.Sprintf("SELECT tile_data FROM %s WHERE zoom_level=? AND tile_column=? AND tile_row=?", t.name), level, x, y).Scan(&data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

//Contains returns true if the reader already contains the tile for a given level/x/y
func (t tileContent) Contains(level int, x, y int) (bool, error) {

	var count int
	err := t.h.db.QueryRow(fmt.Sprintf("SELECT count(*) FROM %s WHERE zoom_level=? AND tile_column=? AND tile_row=?", t.name), level, x, y).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return count == 1, nil
}

//ListTileLayers list all available tile layers
func (h *Handle) ListTileLayers() ([]string, error) {
	var layers []string

	sets, err := h.ListTileMatrixSet()
	if err != nil {
		return nil, err
	}

	for _, tms := range sets {
		layers = append(layers, tms.TableName)
	}

	return layers, nil
}

//OpenTileLayer opens the tile layer for reading
func (h *Handle) OpenTileLayer(name string) (raster.TileReader, error) {
	return tileContent{
		h:    h,
		name: name,
	}, nil
}
