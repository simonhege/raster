// Copyright 2015 Simon HEGE. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gpkg

import (
	"database/sql"
)

//GeometryColumns identifies geometry columns in tables
type GeometryColumns struct {
	TableName        string `json:"table_name"`
	ColumnName       string `json:"column_name"`
	GeometryTypeName string `json:"geometry_type_name"`
	SrsID            int64  `json:"srs_id"`
	Z                uint8  `json:"z"`
	M                uint8  `json:"m"`
}

//ListGeometryColumns retrieves the list of all GeometryColumns registered in the GeoPackage
func (h *Handle) ListGeometryColumns() ([]*GeometryColumns, error) {

	rows, err := queryGeometryColumns(h.db, "")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*GeometryColumns

	for rows.Next() {
		item := &GeometryColumns{}

		err := rows.Scan(item)
		if err != nil {
			return nil, err
		}

		result = append(result, item)
	}

	return result, rows.Err()
}

//ListGeometryColumnsForTable retrieves the list of all GeometryColumns registered in the GeoPackage
func (h *Handle) ListGeometryColumnsForTable(tableName string) ([]*GeometryColumns, error) {

	rows, err := queryGeometryColumns(h.db, "WHERE table_name=?", tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*GeometryColumns

	for rows.Next() {
		item := &GeometryColumns{}

		err := rows.Scan(item)
		if err != nil {
			return nil, err
		}

		result = append(result, item)
	}

	return result, rows.Err()
}

//GetGeometryColumns retrieves the single GeometryColumns having the given TableName and ColumnName in the GeoPackage. Returns sql.ErrNoRows if not found.
func (h *Handle) GetGeometryColumns(tableName, columnName string) (*GeometryColumns, error) {
	return getGeometryColumns(h.db, tableName, columnName)
}

type geometryColumnsRows struct {
	rows *sql.Rows
}

func (rs *geometryColumnsRows) Close() error {
	return rs.rows.Close()
}
func (rs *geometryColumnsRows) Err() error {
	return rs.rows.Err()
}
func (rs *geometryColumnsRows) Next() bool {
	return rs.rows.Next()
}
func (rs *geometryColumnsRows) Scan(dest *GeometryColumns) error {
	return rs.rows.Scan(&dest.TableName, &dest.ColumnName, &dest.GeometryTypeName, &dest.SrsID, &dest.Z, &dest.M)
}

func queryGeometryColumns(db sqlQueryer, additionalClause string, args ...interface{}) (*geometryColumnsRows, error) {

	rows, err := db.Query(`SELECT 
		table_name, 
		column_name,
		geometry_type_name,
		srs_id,
		z,
		m
		FROM gpkg_geometry_columns `+additionalClause, args...)

	if err != nil {
		return nil, err
	}

	return &geometryColumnsRows{rows: rows}, nil
}
func querySingleGeometryColumns(db sqlQueryer, additionalClause string, args ...interface{}) (*GeometryColumns, error) {

	var dest GeometryColumns

	err := db.QueryRow(`SELECT 
		table_name, 
		column_name,
		geometry_type_name,
		srs_id,
		z,
		m
		FROM gpkg_geometry_columns `+additionalClause, args...).Scan(&dest.TableName, &dest.ColumnName, &dest.GeometryTypeName, &dest.SrsID, &dest.Z, &dest.M)

	return &dest, err
}
func getGeometryColumns(db sqlQueryer, tableName string, columnName string) (*GeometryColumns, error) {

	var dest GeometryColumns

	err := db.QueryRow(`SELECT 
		table_name, 
		column_name,
		geometry_type_name,
		srs_id,
		z,
		m
		FROM gpkg_geometry_columns WHERE table_name=? AND column_name=?`, tableName, columnName).Scan(&dest.TableName, &dest.ColumnName, &dest.GeometryTypeName, &dest.SrsID, &dest.Z, &dest.M)

	return &dest, err
}
