# Raster init

Command raster_init provides a CLI tool to create or update a tile data source (such as MBTiles or Geopackage database) based on an other data source (such as a ZXY server) and an area of interest.

Warning: see http://wiki.openstreetmap.org/wiki/Tile_usage_policy before using the OpenStreetMap servers. Bulk downloading is strongly discouraged.
See also http://wiki.openstreetmap.org/wiki/TMS for a list of TMS server.

## Install

The `raster_init` tool depends on GEOS C library being available in your PATH. See https://github.com/paulsmith/gogeos.

    go get github.com/xeonx/raster/cmd/raster_init

## Run

	raster_init -src="http://a.tile.openstreetmap.org/%d/%d/%d.png" -dst="world.mbtiles"

Usage:

    -aoi string
        Area of interest (in WKT) (default "POLYGON((-180 -85.0511, 180 -85.0511, 180 85.0511, -180 85.0511, -180 -85.0511))")
    -dst string
        Destination data source name
    -dstdriver string
        Destination driver
    -dstlayer string
        Destination data source layer name (default "data")
    -levelmax int
        maximum zoom level (default 3)
    -levelmin int
        minimum zoom level (default 0)
    -replace
        force replace of existing tiles
    -src string
        Source data source name
    -srcdriver string
        Source driver
    -srclayer string
        Source data source layer name (default is first layer in the source)

	
## License

This code is licensed under the MIT license. See [LICENSE](https://github.com/xeonx/raster/blob/master/LICENSE).