// Copyright 2015 Simon HEGE. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gpkg

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"net/http"
	"sort"
	"strconv"

	"github.com/gorilla/mux"
)

//Server allows serving GeoPackage through HTTP
type Server struct {
	router    *mux.Router
	h         *Handle
	urlPrefix string
}

//NewServer creates a new Server based on a GeoPackage handle and an URL prefix
func NewServer(h *Handle, urlPrefix string) *Server {

	s := Server{h: h, router: mux.NewRouter(), urlPrefix: urlPrefix}

	s.router.HandleFunc(urlPrefix+"/gpkg_spatial_ref_sys", handleError(encodeJSON(s.spatialRefSysHandler))).Methods("GET")
	s.router.HandleFunc(urlPrefix+"/gpkg_spatial_ref_sys/{srs_id:(-)?[0-9]+}", handleError(encodeJSON(s.singleSpatialRefSysHandler))).Methods("GET")

	s.router.HandleFunc(urlPrefix+"/gpkg_contents", handleError(encodeJSON(s.contentsHandler))).Methods("GET")
	s.router.HandleFunc(urlPrefix+"/gpkg_contents/{table_name}", handleError(encodeJSON(s.singleContents))).Methods("GET")

	s.router.HandleFunc(urlPrefix+"/gpkg_extensions", handleError(encodeJSON(s.extensionHandler))).Methods("GET")

	s.router.HandleFunc(urlPrefix+"/gpkg_geometry_columns", handleError(encodeJSON(s.geometryColumnsHandler))).Methods("GET")
	s.router.HandleFunc(urlPrefix+"/gpkg_geometry_columns/{table_name}", handleError(encodeJSON(s.geometryColumnsByTableNameHandler))).Methods("GET")
	s.router.HandleFunc(urlPrefix+"/gpkg_geometry_columns/{table_name}/{column_name}", handleError(encodeJSON(s.singleGeometryColumns))).Methods("GET")

	s.router.HandleFunc(urlPrefix+"/gpkg_tile_matrix_set", handleError(encodeJSON(s.tileMatrixSetHandler))).Methods("GET")
	s.router.HandleFunc(urlPrefix+"/gpkg_tile_matrix_set/{table_name}", handleError(encodeJSON(s.singleTileMatrixSet))).Methods("GET")

	s.router.HandleFunc(urlPrefix+"/gpkg_tile_matrix", handleError(encodeJSON(s.tileMatrixHandler))).Methods("GET")
	s.router.HandleFunc(urlPrefix+"/gpkg_tile_matrix/{table_name}", handleError(encodeJSON(s.tileMatrixByTableNameHandler))).Methods("GET")
	s.router.HandleFunc(urlPrefix+"/gpkg_tile_matrix/{table_name}/{level:[0-9]+}", handleError(encodeJSON(s.singleTileMatrix))).Methods("GET")

	s.router.HandleFunc(urlPrefix+"/{table_name}/{level:[0-9]+}/{x:[0-9]+}/{y:[0-9]+}.{ext}", handleError(encodeImage(s.tileHandler))).Methods("GET")

	s.router.HandleFunc(urlPrefix+"/{table_name}", handleError(encodeJSON(s.tableHandler))).Methods("GET")

	s.router.HandleFunc(urlPrefix+"/", handleError(s.indexHandler)).Methods("GET")

	return &s
}

func (s *Server) spatialRefSysHandler(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return s.h.ListSpatialRefSys()
}

func (s *Server) singleSpatialRefSysHandler(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	SrsID, err := strconv.ParseInt(vars["srs_id"], 10, 64)
	if err != nil {
		return nil, newDataError("SRS id", err)
	}

	return s.h.GetSpatialRefSys(SrsID)
}

func (s *Server) extensionHandler(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return s.h.ListExtensions()
}

func (s *Server) contentsHandler(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return s.h.ListContents()
}

func (s *Server) singleContents(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	tableName := vars["table_name"]
	if len(tableName) == 0 {
		return nil, newDataError("Table name", errors.New("empty name not allowed"))
	}

	return s.h.GetContents(tableName)
}

func (s *Server) geometryColumnsHandler(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return s.h.ListGeometryColumns()
}

func (s *Server) geometryColumnsByTableNameHandler(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	tableName := vars["table_name"]
	if len(tableName) == 0 {
		return nil, newDataError("Table name", errors.New("empty name not allowed"))
	}

	return s.h.ListGeometryColumnsForTable(tableName)
}

func (s *Server) singleGeometryColumns(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	tableName := vars["table_name"]
	if len(tableName) == 0 {
		return nil, newDataError("Table name", errors.New("empty name not allowed"))
	}
	columnName := vars["column_name"]
	if len(columnName) == 0 {
		return nil, newDataError("Column name", errors.New("empty name not allowed"))
	}

	return s.h.GetGeometryColumns(tableName, columnName)
}

func (s *Server) tileMatrixSetHandler(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return s.h.ListTileMatrixSet()
}

func (s *Server) singleTileMatrixSet(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	tableName := vars["table_name"]
	if len(tableName) == 0 {
		return nil, newDataError("Table name", errors.New("empty name not allowed"))
	}

	return s.h.GetTileMatrixSet(tableName)
}

func (s *Server) tileMatrixHandler(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return s.h.ListTileMatrix()
}

func (s *Server) tileMatrixByTableNameHandler(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	tableName := vars["table_name"]
	if len(tableName) == 0 {
		return nil, newDataError("Table name", errors.New("empty name not allowed"))
	}

	return s.h.ListTileMatrixForTable(tableName)
}

func (s *Server) singleTileMatrix(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	tableName := vars["table_name"]
	if len(tableName) == 0 {
		return nil, newDataError("Table name", errors.New("empty name not allowed"))
	}
	zoomLevel, err := strconv.ParseInt(vars["level"], 10, 64)
	if err != nil {
		return nil, newDataError("Zoom level", err)
	}

	return s.h.GetTileMatrix(tableName, zoomLevel)
}

func (s *Server) tableHandler(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	tableName := vars["table_name"]
	if len(tableName) == 0 {
		return nil, newDataError("Table name", errors.New("empty name not allowed"))
	}
	return s.h.ListFeatures(tableName)
}

func (s *Server) tileHandler(w http.ResponseWriter, r *http.Request) (image.Image, error) {
	vars := mux.Vars(r)
	tableName := vars["table_name"]
	if len(tableName) == 0 {
		return nil, newDataError("Table name", errors.New("empty name not allowed"))
	}
	level, err := strconv.ParseInt(vars["level"], 10, 64)
	if err != nil {
		return nil, newDataError("Level", err)
	}
	x, err := strconv.ParseInt(vars["x"], 10, 64)
	if err != nil {
		return nil, newDataError("X", err)
	}
	y, err := strconv.ParseInt(vars["y"], 10, 64)
	if err != nil {
		return nil, newDataError("Y", err)
	}
	return s.h.GetTile(tableName, level, x, y)
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) error {
	mapLinks := make(map[string]string)

	//List metadata tables
	mapLinks["gpkg_contents"] = s.urlPrefix + "/gpkg_contents"
	mapLinks["gpkg_spatial_ref_sys"] = s.urlPrefix + "/gpkg_spatial_ref_sys"

	if s.h.HasExtensionsTable() {
		mapLinks["gpkg_extensions"] = s.urlPrefix + "/gpkg_extensions"
	}

	if s.h.HasFeatures() {
		mapLinks["gpkg_geometry_columns"] = s.urlPrefix + "/gpkg_geometry_columns"
	}

	if s.h.HasTiles() {
		mapLinks["gpkg_tile_matrix_set"] = s.urlPrefix + "/gpkg_tile_matrix_set"
		mapLinks["gpkg_tile_matrix"] = s.urlPrefix + "/gpkg_tile_matrix"
	}

	//List feature and tiles tables
	tables, err := s.h.ListContents()
	if err != nil {
		return nil
	}
	for _, t := range tables {
		mapLinks[t.TableName] = s.urlPrefix + "/" + t.TableName
	}

	routes := make([]string, 0, len(mapLinks))
	for k := range mapLinks {
		routes = append(routes, k)
	}
	sort.Strings(routes)

	for _, route := range routes {

		fmt.Fprintf(w, "<a href=\"%s\">%s</a></br>", mapLinks[route], route)

	}
	return nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	s.router.ServeHTTP(w, r)
}

type handlerFunc func(w http.ResponseWriter, r *http.Request) error

func handleError(f handlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		err := f(w, r)

		if err != nil {
			log.Print(err)

			httpErr := http.StatusInternalServerError
			//Try to cast to the known error types
			if _, ok := err.(dataError); ok {
				httpErr = http.StatusBadRequest
			} else if err == sql.ErrNoRows {
				httpErr = http.StatusNotFound
				err = errors.New("404 page not found")
			}

			http.Error(w, err.Error(), httpErr)
			return
		}
	}
}
func encodeJSON(f func(w http.ResponseWriter, r *http.Request) (interface{}, error)) handlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {

		data, err := f(w, r)
		if err != nil {
			return err
		}

		w.WriteHeader(http.StatusOK)
		encoder := json.NewEncoder(w)
		err = encoder.Encode(data)
		if err != nil {
			return err
		}

		return nil
	}
}
func encodeImage(f func(w http.ResponseWriter, r *http.Request) (image.Image, error)) handlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {

		vars := mux.Vars(r)
		ext := vars["ext"]
		if len(ext) == 0 {
			return newDataError("Extension", errors.New("empty value not allowed"))
		}

		image, err := f(w, r)
		if err != nil {
			return err
		}

		w.WriteHeader(http.StatusOK)
		if ext == "png" {
			return png.Encode(w, image)
		}

		return jpeg.Encode(w, image, nil)
	}
}

//dataError represents an input data error
type dataError struct {
	Message string
}

func (err dataError) Error() string {
	return err.Message
}

//newDataError returns a DataError for the given input field
func newDataError(field string, err error) error {
	return dataError{
		Message: fmt.Sprintf("%s: %s", field, err),
	}
}
