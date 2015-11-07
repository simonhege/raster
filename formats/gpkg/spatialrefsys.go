// Copyright 2015 Simon HEGE. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gpkg

import (
	"database/sql"
)

//SpatialRefSys defines a spatial reference system.
type SpatialRefSys struct {
	SrsName                string  `sqler:"srs_name" json:"srs_name"`
	SrsID                  int64   `sqler:"srs_id,pk" json:"srs_id"`
	Organization           string  `sqler:"organization" json:"organization"`
	OrganizationCoordsysID int64   `sqler:"organization_coordsys_id" json:"organization_coordsys_id"`
	Definition             string  `sqler:"definition" json:"definition"`
	Description            *string `sqler:"description,nullable" json:"description,omitempty"`
}

//ListSpatialRefSys retrieves the list of all SpatialRefSys registered in the GeoPackage
func (h *Handle) ListSpatialRefSys() ([]*SpatialRefSys, error) {

	rows, err := querySpatialRefSys(h.db, "")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*SpatialRefSys

	for rows.Next() {
		item := &SpatialRefSys{}

		err := rows.Scan(item)
		if err != nil {
			return nil, err
		}

		result = append(result, item)
	}

	return result, rows.Err()

}

//GetSpatialRefSys retrieves the single SpatialRefSys having the given SrsID in the GeoPackage. Returns sql.ErrNoRows if not found.
func (h *Handle) GetSpatialRefSys(SrsID int64) (*SpatialRefSys, error) {
	return getSpatialRefSys(h.db, SrsID)
}

type spatialRefSysRows struct {
	rows *sql.Rows
}

func (rs *spatialRefSysRows) Close() error {
	return rs.rows.Close()
}
func (rs *spatialRefSysRows) Err() error {
	return rs.rows.Err()
}
func (rs *spatialRefSysRows) Next() bool {
	return rs.rows.Next()
}
func (rs *spatialRefSysRows) Scan(dest *SpatialRefSys) error {
	return rs.rows.Scan(&dest.SrsName, &dest.SrsID, &dest.Organization, &dest.OrganizationCoordsysID, &dest.Definition, &dest.Description)
}

func querySpatialRefSys(db sqlQueryer, additionalClause string, args ...interface{}) (*spatialRefSysRows, error) {

	rows, err := db.Query(`SELECT 
		srs_name, 
		srs_id,
		organization,
		organization_coordsys_id,
		definition,
		description
		FROM gpkg_spatial_ref_sys `+additionalClause, args...)

	if err != nil {
		return nil, err
	}

	return &spatialRefSysRows{rows: rows}, nil
}
func querySingleSpatialRefSys(db sqlQueryer, additionalClause string, args ...interface{}) (*SpatialRefSys, error) {

	var dest SpatialRefSys

	err := db.QueryRow(`SELECT 
		srs_name, 
		srs_id,
		organization,
		organization_coordsys_id,
		definition,
		description
		FROM gpkg_spatial_ref_sys `+additionalClause, args...).Scan(&dest.SrsName, &dest.SrsID, &dest.Organization, &dest.OrganizationCoordsysID, &dest.Definition, &dest.Description)

	return &dest, err
}
func getSpatialRefSys(db sqlQueryer, SrsID int64) (*SpatialRefSys, error) {

	var dest SpatialRefSys

	err := db.QueryRow(`SELECT 
		srs_name, 
		srs_id,
		organization,
		organization_coordsys_id,
		definition,
		description
		FROM gpkg_spatial_ref_sys WHERE srs_id=?`, SrsID).Scan(&dest.SrsName, &dest.SrsID, &dest.Organization, &dest.OrganizationCoordsysID, &dest.Definition, &dest.Description)

	return &dest, err
}
