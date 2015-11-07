# raster
Package raster provides tiled raster data (such as MBTiles, GeoPackage, OpenStreetMap servers, ...) manipulation helpers.

The available tools are:
  * [raster_init](https://github.com/xeonx/raster/tree/master/cmd/raster_init): performs conversion between tile datasource (ex: extract a GeoPackage into a tile folder)
  * [raster_server](https://github.com/xeonx/raster/tree/master/cmd/raster_server): serve a tile data source on an HTTP server. It exposes a TMS like server and an OpenLayers webpage displaying the layers. Conversion between latitude/longitude and x/y in global-mercator is performed as described
in http://wiki.openstreetmap.org/wiki/Slippy_map_tilenames . 

Currently available [drivers](https://github.com/xeonx/raster/tree/master/formats) are:
  * [GeoPackage](https://github.com/xeonx/raster/tree/master/formats/gpkg)
  * [MBTiles](https://github.com/xeonx/raster/tree/master/formats/gpkg)
  * [ZXY server (TMS like)](https://github.com/xeonx/raster/tree/master/formats/zxyserver)
  * [Tile folder](https://github.com/xeonx/raster/tree/master/formats/tilefolder)

[![GoDoc](https://godoc.org/github.com/xeonx/raster?status.svg)](https://godoc.org/github.com/xeonx/raster)

## Install
The `raster_init` tool depends on GEOS C library being available in your PATH. See https://github.com/paulsmith/gogeos.

    go get github.com/xeonx/raster/...

## Roadmap
  * more tests and benchmarks 
  * improved documentation (code examples)

## License
This code is licensed under the MIT license. See [LICENSE](https://github.com/xeonx/raster/blob/master/LICENSE).