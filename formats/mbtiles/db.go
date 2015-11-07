// Copyright 2014-2015 Simon HEGE. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package mbtiles

import (
	"database/sql"
	"image"

	"github.com/xeonx/raster"
)

//Metadata contains all the metadata stored into a MBTiles database.
type Metadata struct {
	Name        string
	Type        string //overlay or baselayer
	Version     int
	Description string
	Format      string //png or jpg
	//TODO Bounds
	Attribution string
	//TODO UTFGrid keys
}

//DB is a database handle to a MBTiles sqlite database. It's safe for concurrent use by multiple goroutines.
type DB struct {
	db       *sql.DB
	metadata Metadata
}

func hasTable(db *sql.DB, tableName string) bool {
	var count int
	err := db.QueryRow("SELECT count(*) FROM sqlite_master WHERE name=?", tableName).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

//Open opens a MBTiles sqlite database at the given location.
func Open(filepath string) (*DB, error) {
	//Open database
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}
	mbtilesDb := &DB{db: db}

	//Check if metadata and tiles table exists
	if !hasTable(db, "metadata") {
		_, err = db.Exec(`
			CREATE TABLE metadata (name text, value text);
		`)
		if err != nil {
			mbtilesDb.Close()
			return nil, err
		}
	}
	if !hasTable(db, "tiles") {
		_, err = db.Exec(`
			CREATE TABLE tiles (zoom_level integer, tile_column integer, tile_row integer, tile_data blob);
			CREATE INDEX tiles_idx ON tiles (zoom_level, tile_column, tile_row);
		`)
		if err != nil {
			mbtilesDb.Close()
			return nil, err
		}
	}

	//Read metadata
	m := map[string]interface{}{
		"name":        &mbtilesDb.metadata.Name,
		"type":        &mbtilesDb.metadata.Type,
		"version":     &mbtilesDb.metadata.Version,
		"description": &mbtilesDb.metadata.Description,
		"format":      &mbtilesDb.metadata.Format,
		"attribution": &mbtilesDb.metadata.Attribution,
	}
	for k, v := range m {
		err = mbtilesDb.readMetadata(k, v)
		if err != nil {
			mbtilesDb.Close()
			return nil, err
		}
	}

	//Set default values on empty items
	if mbtilesDb.metadata.Type == "" {
		mbtilesDb.metadata.Type = "baselayer"
	}
	if mbtilesDb.metadata.Format == "" {
		mbtilesDb.metadata.Format = "png"
	}

	return mbtilesDb, nil
}

//Create creates a MBTiles sqlite database at the given location.
func Create(filepath string, metadata Metadata) (*DB, error) {

	//Create database
	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, err
	}
	mbtilesDb := &DB{db: db}

	//Init schema
	_, err = db.Exec(`
		CREATE TABLE metadata (name text, value text);
		CREATE TABLE tiles (zoom_level integer, tile_column integer, tile_row integer, tile_data blob);
		CREATE INDEX tiles_idx ON tiles (zoom_level, tile_column, tile_row);
	`)
	if err != nil {
		mbtilesDb.Close()
		return nil, err
	}

	//Store metadata
	m := map[string]interface{}{
		"name":        metadata.Name,
		"type":        metadata.Type,
		"version":     metadata.Version,
		"description": metadata.Description,
		"format":      metadata.Format,
		"attribution": metadata.Attribution,
	}
	for k, v := range m {
		err = mbtilesDb.saveMetadata(k, v)
		if err != nil {
			mbtilesDb.Close()
			return nil, err
		}
	}

	mbtilesDb.metadata = metadata

	return mbtilesDb, nil
}

//TileFormat exposes the image format of the source (png or jpg)
func (m *DB) TileFormat() string {
	return m.metadata.Format
}

//Close closes the underlying sqlite connexion
func (m *DB) Close() error {
	if m.db != nil {
		if err := m.db.Close(); err != nil {
			return err
		}
	}
	return nil
}

//Metadata retrieves the metadata stored into a MBTiles database.
func (m *DB) Metadata() Metadata {
	return m.metadata
}

//saveMetadata saves a metadata item
func (m *DB) saveMetadata(key string, value interface{}) error {
	_, err := m.db.Exec("INSERT OR REPLACE INTO metadata VALUES ( ? , ? )", key, value)
	return err
}

//readMetadata retrieves a metadata item
func (m *DB) readMetadata(key string, value interface{}) error {
	err := m.db.QueryRow("SELECT value FROM metadata WHERE name = ?", key).Scan(value)
	if err == sql.ErrNoRows {
		return nil //No change to the value. Reset should have been done by the caller.
	}
	return err
}

//GetRaw retrieves the tile for a given level/x/y. No check is performed on the image format.
func (m *DB) GetRaw(level int, x, y int) ([]byte, error) {
	var tileData []byte
	err := m.db.QueryRow("SELECT tile_data FROM tiles WHERE zoom_level = ? AND tile_column = ? AND tile_row = ?", level, x, y).Scan(&tileData)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return tileData, nil
}

//Get retrieves and decodes the tile for a given level/x/y
func (m *DB) Get(level int, x, y int) (image.Image, error) {

	tileData, err := m.GetRaw(level, x, y)
	if err != nil {
		return nil, err
	}

	if tileData == nil {
		return nil, nil
	}

	return raster.Decode(tileData, m.metadata.Format)
}

//Contains returns true if the database already contains the tile for a given level/x/y
func (m *DB) Contains(level int, x, y int) (bool, error) {
	var res int
	err := m.db.QueryRow("SELECT zoom_level FROM tiles WHERE zoom_level = ? AND tile_column = ? AND tile_row = ?", level, x, y).Scan(&res)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

//SetRaw stores the tile for a given level/x/y. No check is performed on the image format.
func (m *DB) SetRaw(level int, x, y int, img []byte) error {

	_, err := m.db.Exec("INSERT INTO tiles (zoom_level, tile_column, tile_row, tile_data) VALUES ( ? , ? , ? , ? )", level, x, y, img)

	return err
}

//Set encodes and stores accordingly to the specified format the tile for a given level/x/y
func (m *DB) Set(level int, x, y int, img image.Image) error {

	b, err := raster.Encode(img, m.metadata.Format)
	if err != nil {
		return err
	}

	return m.SetRaw(level, x, y, b)
}

//Clear removes all tiles for a given layer
func (m *DB) Clear(level int) error {
	_, err := m.db.Exec("DELETE FROM tiles WHERE zoom_level = ?;", level)

	return err
}

//ListTileLayers list all available tile layers
//
//As mbtiles consist of a single layer, the returned array wil have a lenght of 1.
func (m *DB) ListTileLayers() ([]string, error) {
	var layers []string
	layers = append(layers, m.metadata.Name)
	return layers, nil
}

//OpenTileLayer opens the tile layer for reading
//
//As mbtiles consist of a single layer, the same layer will be returned or raster.ErrLayerNotFund will be raised.
func (m *DB) OpenTileLayer(name string) (raster.TileReader, error) {
	if name == m.metadata.Name {
		return m, nil
	}
	return nil, raster.ErrLayerNotFund
}

//CreateTileLayer creates a layer in the data source, or opens it for writing
//
//As mbtiles consist of a single layer, no new layer can be created.
func (m *DB) CreateTileLayer(name string) (raster.TileReadWriter, error) {
	if name == m.metadata.Name || m.metadata.Name == "" {
		return m, nil
	}
	return nil, raster.ErrLayerNotFund
}
