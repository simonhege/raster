# Raster server

Command raster_server provides an HTTP server for tile data sources.

## Install

	go get github.com/xeonx/raster/cmd/raster_server

## Run

	raster_server -db="mydb.db" -http=":8085"

Usage
	-http string
	    HTTP service address (e.g., '127.0.0.1:8085' or just ':8085') (default ":8085")
    -src string
        Source data source name
    -srcdriver string
        Source driver
    -srclayer string
        Source data source layer name (default is first layer in the source)

Then open your browser at
	http://localhost:8085/map.html
	
Tiles are served at
	http://localhost:8085/tiles/0/0/0.png
	
## Docs

<http://godoc.org/github.com/xeonx/raster/cmd/raster_server>
	
## License

This code is licensed under the MIT license. See [LICENSE](https://github.com/xeonx/raster/blob/master/LICENSE).