// Copyright 2015 Simon HEGE. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gpkg

import (
	"database/sql"
	"time"
)

//Contents provides identifying and descriptive information about available data
type Contents struct {
	TableName   string    `json:"table_name"`
	DataType    string    `json:"data_type"` //features or tiles
	Identifier  *string   `json:"identifier,omitempty"`
	Description *string   `json:"description,omitempty"`
	LastChange  time.Time `json:"last_change"`
	MinX        *float64  `json:"min_x,omitempty"`
	MinY        *float64  `json:"min_y,omitempty"`
	MaxX        *float64  `json:"max_x,omitempty"`
	MaxY        *float64  `json:"max_y,omitempty"`
	SrsID       *int64    `json:"srs_id,omitempty"`
}

//ListContents retrieves the list of all Contents registered in the GeoPackage
func (h *Handle) ListContents() ([]*Contents, error) {

	rows, err := queryContents(h.db, "")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*Contents

	for rows.Next() {
		item := &Contents{}

		err := rows.Scan(item)
		if err != nil {
			return nil, err
		}

		result = append(result, item)
	}

	return result, rows.Err()

}

//GetContents retrieves the single Contents having the given TableName in the GeoPackage. Returns sql.ErrNoRows if not found.
func (h *Handle) GetContents(tableName string) (*Contents, error) {
	return getContents(h.db, tableName)
}

type contentsRows struct {
	rows *sql.Rows
}

func (rs *contentsRows) Close() error {
	return rs.rows.Close()
}
func (rs *contentsRows) Err() error {
	return rs.rows.Err()
}
func (rs *contentsRows) Next() bool {
	return rs.rows.Next()
}
func (rs *contentsRows) Scan(dest *Contents) error {
	return rs.rows.Scan(&dest.TableName, &dest.DataType, &dest.Identifier, &dest.Description, &dest.LastChange, &dest.MinX, &dest.MinY, &dest.MaxX, &dest.MaxY, &dest.SrsID)
}

func queryContents(db sqlQueryer, additionalClause string, args ...interface{}) (*contentsRows, error) {

	rows, err := db.Query(`SELECT 
		table_name, 
		data_type,
		identifier,
		description,
		last_change,
		min_x,
		min_y,
		max_x,
		max_y,
		srs_id
		FROM gpkg_contents `+additionalClause, args...)

	if err != nil {
		return nil, err
	}

	return &contentsRows{rows: rows}, nil
}
func querySingleContents(db sqlQueryer, additionalClause string, args ...interface{}) (*Contents, error) {

	var dest Contents

	err := db.QueryRow(`SELECT 
		table_name, 
		data_type,
		identifier,
		description,
		last_change,
		min_x,
		min_y,
		max_x,
		max_y,
		srs_id
		FROM gpkg_contents `+additionalClause, args...).Scan(&dest.TableName, &dest.DataType, &dest.Identifier, &dest.Description, &dest.LastChange, &dest.MinX, &dest.MinY, &dest.MaxX, &dest.MaxY, &dest.SrsID)

	return &dest, err
}
func getContents(db sqlQueryer, tableName string) (*Contents, error) {

	var dest Contents

	err := db.QueryRow(`SELECT 
		table_name, 
		data_type,
		identifier,
		description,
		last_change,
		min_x,
		min_y,
		max_x,
		max_y,
		srs_id
		FROM gpkg_contents WHERE table_name=?`, tableName).Scan(&dest.TableName, &dest.DataType, &dest.Identifier, &dest.Description, &dest.LastChange, &dest.MinX, &dest.MinY, &dest.MaxX, &dest.MaxY, &dest.SrsID)

	return &dest, err
}
