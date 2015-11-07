# gpkg

Package gpkg provides a GeoPackage compliant database manipulation helpers.

An sqlite3 driver should be imported in main package, for example "github.com/mattn/go-sqlite3"

GeoPackage is an open, standards-based, platform-independent, portable, self-describing, compact format for transferring geospatial information.
It is defined in OGC 12-128r11. Additional resources can be found at http://www.geopackage.org/
  
## Install

    go get github.com/xeonx/raster/formats/gpkg

## Docs

[![GoDoc](https://godoc.org/github.com/xeonx/raster/formats/gpkg?status.svg)](https://godoc.org/github.com/xeonx/raster/formats/gpkg)
	
## Tests

`go test` is used for testing.

## Roadmap
  * write methods, including GeoPackage creation and update from various data sources
  * more tests and benchmarks 
  * improved documentation (code examples)

## License

This code is licensed under the MIT license. See [LICENSE](https://github.com/xeonx/raster/blob/master/LICENSE).