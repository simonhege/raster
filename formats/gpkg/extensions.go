// Copyright 2015 Simon HEGE. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gpkg

import (
	"database/sql"
)

//Extensions provides information about a particular extension available for a package, a table or a column.
type Extensions struct {
	TableName     *string `json:"table_name,omitempty"`
	ColumnName    *string `json:"column_name,omitempty"`
	ExtensionName string  `json:"extension_name"`
	Definition    string  `json:"definition"`
	Scope         *string `json:"scope"`
}

//ListExtensions retrieves the list of all Extensions registered in the GeoPackage
func (h *Handle) ListExtensions() ([]*Extensions, error) {

	rows, err := queryExtensions(h.db, "")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*Extensions

	for rows.Next() {
		item := &Extensions{}

		err := rows.Scan(item)
		if err != nil {
			return nil, err
		}

		result = append(result, item)
	}

	return result, rows.Err()
}

type extensionsRows struct {
	rows *sql.Rows
}

func (rs *extensionsRows) Close() error {
	return rs.rows.Close()
}
func (rs *extensionsRows) Err() error {
	return rs.rows.Err()
}
func (rs *extensionsRows) Next() bool {
	return rs.rows.Next()
}
func (rs *extensionsRows) Scan(dest *Extensions) error {
	return rs.rows.Scan(&dest.TableName, &dest.ColumnName, &dest.ExtensionName, &dest.Definition, &dest.Scope)
}

func queryExtensions(db sqlQueryer, additionalClause string, args ...interface{}) (*extensionsRows, error) {

	rows, err := db.Query(`SELECT 
		table_name, 
		column_name,
		extension_name,
		definition,
		scope
		FROM gpkg_extensions `+additionalClause, args...)

	if err != nil {
		return nil, err
	}

	return &extensionsRows{rows: rows}, nil
}
func querySingleExtensions(db sqlQueryer, additionalClause string, args ...interface{}) (*Extensions, error) {

	var dest Extensions

	err := db.QueryRow(`SELECT 
		table_name, 
		column_name,
		extension_name,
		definition,
		scope
		FROM gpkg_extensions `+additionalClause, args...).Scan(&dest.TableName, &dest.ColumnName, &dest.ExtensionName, &dest.Definition, &dest.Scope)

	return &dest, err
}
func getExtensions(db sqlQueryer, tableName string) (*Extensions, error) {

	var dest Extensions

	err := db.QueryRow(`SELECT 
		table_name, 
		column_name,
		extension_name,
		definition,
		scope
		FROM gpkg_extensions WHERE table_name=?`, tableName).Scan(&dest.TableName, &dest.ColumnName, &dest.ExtensionName, &dest.Definition, &dest.Scope)

	return &dest, err
}
