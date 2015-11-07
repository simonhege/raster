// Copyright 2015 Simon HEGE. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gpkg

import (
	"database/sql"
)

//TileMatrix documents the structure of the tile matrix at each zoom level in each tiles table.
type TileMatrix struct {
	TableName    string  `json:"table_name"`
	ZoomLevel    int64   `json:"zoom_level"`
	MatrixWidth  int64   `json:"matrix_width"`
	MatrixHeight int64   `json:"matrix_height"`
	TileWidth    int64   `json:"tile_width"`
	TileHeight   int64   `json:"tile_height"`
	PixelXSize   float64 `json:"pixel_x_size"`
	PixelYSize   float64 `json:"pixel_y_size"`
}

//ListTileMatrix retrieves the list of all TileMatrix registered in the GeoPackage
func (h *Handle) ListTileMatrix() ([]*TileMatrix, error) {

	rows, err := queryTileMatrix(h.db, "")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*TileMatrix

	for rows.Next() {
		item := &TileMatrix{}

		err := rows.Scan(item)
		if err != nil {
			return nil, err
		}

		result = append(result, item)
	}

	return result, rows.Err()
}

//ListTileMatrixForTable retrieves the list of all TileMatrix registered in the GeoPackage
func (h *Handle) ListTileMatrixForTable(tableName string) ([]*TileMatrix, error) {

	rows, err := queryTileMatrix(h.db, "WHERE table_name=?", tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*TileMatrix

	for rows.Next() {
		item := &TileMatrix{}

		err := rows.Scan(item)
		if err != nil {
			return nil, err
		}

		result = append(result, item)
	}

	return result, rows.Err()
}

//GetTileMatrix retrieves the single TileMatrix having the given TableName and ZoomLevel in the GeoPackage. Returns sql.ErrNoRows if not found.
func (h *Handle) GetTileMatrix(tableName string, zoomLevel int64) (*TileMatrix, error) {
	return getTileMatrix(h.db, tableName, zoomLevel)
}

type tileMatrixRows struct {
	rows *sql.Rows
}

func (rs *tileMatrixRows) Close() error {
	return rs.rows.Close()
}
func (rs *tileMatrixRows) Err() error {
	return rs.rows.Err()
}
func (rs *tileMatrixRows) Next() bool {
	return rs.rows.Next()
}
func (rs *tileMatrixRows) Scan(dest *TileMatrix) error {
	return rs.rows.Scan(&dest.TableName, &dest.ZoomLevel, &dest.MatrixWidth, &dest.MatrixHeight, &dest.TileWidth, &dest.TileHeight, &dest.PixelXSize, &dest.PixelYSize)
}

func queryTileMatrix(db sqlQueryer, additionalClause string, args ...interface{}) (*tileMatrixRows, error) {

	rows, err := db.Query(`SELECT 
		table_name,
		zoom_level, 
		matrix_width,
		matrix_height,
		tile_width,
		tile_height,
		pixel_x_size,
		pixel_y_size
		FROM gpkg_tile_matrix `+additionalClause, args...)

	if err != nil {
		return nil, err
	}

	return &tileMatrixRows{rows: rows}, nil
}
func querySingleTileMatrix(db sqlQueryer, additionalClause string, args ...interface{}) (*TileMatrix, error) {

	var dest TileMatrix

	err := db.QueryRow(`SELECT 
		table_name,
		zoom_level, 
		matrix_width,
		matrix_height,
		tile_width,
		tile_height,
		pixel_x_size,
		pixel_y_size
		FROM gpkg_tile_matrix `+additionalClause, args...).Scan(&dest.TableName, &dest.ZoomLevel, &dest.MatrixWidth, &dest.MatrixHeight, &dest.TileWidth, &dest.TileHeight, &dest.PixelXSize, &dest.PixelYSize)

	return &dest, err
}
func getTileMatrix(db sqlQueryer, tableName string, zoomLevel int64) (*TileMatrix, error) {

	var dest TileMatrix

	err := db.QueryRow(`SELECT
		table_name,
		zoom_level, 
		matrix_width,
		matrix_height,
		tile_width,
		tile_height,
		pixel_x_size,
		pixel_y_size
		FROM gpkg_tile_matrix WHERE table_name=? AND zoom_level=?`, tableName, zoomLevel).Scan(&dest.TableName, &dest.ZoomLevel, &dest.MatrixWidth, &dest.MatrixHeight, &dest.TileWidth, &dest.TileHeight, &dest.PixelXSize, &dest.PixelYSize)

	return &dest, err
}
