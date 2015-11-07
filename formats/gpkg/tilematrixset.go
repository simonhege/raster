// Copyright 2015 Simon HEGE. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gpkg

import (
	"database/sql"
)

//TileMatrixSet defines the spatial reference system and the maximum bounding box
//for all possible tiles in a tile pyramid user data table.
type TileMatrixSet struct {
	TableName string  `json:"table_name"`
	SrsID     int64   `json:"srs_id"`
	MinX      float64 `json:"min_x"`
	MinY      float64 `json:"min_y"`
	MaxX      float64 `json:"max_x"`
	MaxY      float64 `json:"max_y"`
}

//ListTileMatrixSet retrieves the list of all TileMatrixSet registered in the GeoPackage
func (h *Handle) ListTileMatrixSet() ([]*TileMatrixSet, error) {

	rows, err := queryTileMatrixSet(h.db, "")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*TileMatrixSet

	for rows.Next() {
		item := &TileMatrixSet{}

		err := rows.Scan(item)
		if err != nil {
			return nil, err
		}

		result = append(result, item)
	}

	return result, rows.Err()
}

//GetTileMatrixSet retrieves the single TileMatrixSet having the given TableName in the GeoPackage. Returns sql.ErrNoRows if not found.
func (h *Handle) GetTileMatrixSet(tableName string) (*TileMatrixSet, error) {
	return getTileMatrixSet(h.db, tableName)
}

type tileMatrixSetRows struct {
	rows *sql.Rows
}

func (rs *tileMatrixSetRows) Close() error {
	return rs.rows.Close()
}
func (rs *tileMatrixSetRows) Err() error {
	return rs.rows.Err()
}
func (rs *tileMatrixSetRows) Next() bool {
	return rs.rows.Next()
}
func (rs *tileMatrixSetRows) Scan(dest *TileMatrixSet) error {
	return rs.rows.Scan(&dest.TableName, &dest.SrsID, &dest.MinX, &dest.MinY, &dest.MaxX, &dest.MaxY)
}

func queryTileMatrixSet(db sqlQueryer, additionalClause string, args ...interface{}) (*tileMatrixSetRows, error) {

	rows, err := db.Query(`SELECT 
		table_name,
		srs_id, 
		min_x,
		min_y,
		max_x,
		max_y
		FROM gpkg_tile_matrix_set `+additionalClause, args...)

	if err != nil {
		return nil, err
	}

	return &tileMatrixSetRows{rows: rows}, nil
}
func querySingleTileMatrixSet(db sqlQueryer, additionalClause string, args ...interface{}) (*TileMatrixSet, error) {

	var dest TileMatrixSet

	err := db.QueryRow(`SELECT 
		table_name, 
		srs_id,
		min_x,
		min_y,
		max_x,
		max_y
		FROM gpkg_tile_matrix_set `+additionalClause, args...).Scan(&dest.TableName, &dest.SrsID, &dest.MinX, &dest.MinY, &dest.MaxX, &dest.MaxY)

	return &dest, err
}
func getTileMatrixSet(db sqlQueryer, tableName string) (*TileMatrixSet, error) {

	var dest TileMatrixSet

	err := db.QueryRow(`SELECT 
		table_name, 
		srs_id,
		min_x,
		min_y,
		max_x,
		max_y
		FROM gpkg_tile_matrix_set WHERE table_name=?`, tableName).Scan(&dest.TableName, &dest.SrsID, &dest.MinX, &dest.MinY, &dest.MaxX, &dest.MaxY)

	return &dest, err
}
