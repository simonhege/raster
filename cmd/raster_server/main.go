// Copyright 2015 Simon HEGE. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//Command raster_server provides an HTTP server for MBTiles compliant database.
//
//You can run it using
//	raster_server -db="mydb.db" -http=":8085"
//
//Tiles are served at http://localhost:8085/tiles/level/x/y.png and an OpenLayers map is available at http://localhost:8085/
//
package main

import (
	"flag"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"

	"github.com/xeonx/raster"
	_ "github.com/xeonx/raster/formats/gpkg"
	_ "github.com/xeonx/raster/formats/mbtiles"
	_ "github.com/xeonx/raster/formats/zxyserver"
)

var src = flag.String("src", "", "Source data source name")
var srcDriver = flag.String("srcdriver", "", "Source driver")
var srcLayer = flag.String("srclayer", "", "Source data source layer name")

var addr = flag.String("http", ":8085", "HTTP service address (e.g., '127.0.0.1:8085' or just ':8085')")

var page = `<!doctype html>
<html lang="en">
  <head>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/ol3/3.5.0/ol.css" type="text/css">
    <style>
      html, body, .map {
        margin: 0;
        padding: 0;
        width: 100%;
        height: 100%;
      }
    </style>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/ol3/3.5.0/ol.js" type="text/javascript"></script>
    <title>{{.Name}} - MBTiles server</title>
  </head>
  <body>
    <div id="map" class="map"></div>
    <script type="text/javascript">
      var map = new ol.Map({
        target: 'map',
        layers: [
			new ol.layer.Tile({
			  source: new ol.source.XYZ({
				url: '/tiles/{z}/{x}/{-y}.{{.Reader.TileFormat}}'
			  })
			})
        ],
        view: new ol.View({
          center: [0,0],
          zoom: 1
        })
      });
    </script>
  </body>
</html>`
var t = template.Must(template.New("page").Parse(page))

type mapPageHandler struct {
	Name   string
	Reader raster.TileReader
}

func (h mapPageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := t.Execute(w, h)
	if err != nil {
		log.Fatal("Execute template: ", err)
	}
}

type closer interface {
	Close() error
}

func main() {
	flag.Parse()

	//Open the data source
	if len(*srcDriver) == 0 {
		*srcDriver = raster.FindDriverName(*src)
	}
	input, err := raster.Open(*srcDriver, *src)
	if err != nil {
		log.Fatal("Open data source: ", err)
	}
	if c, ok := input.(closer); ok {
		defer c.Close()
	}
	var tileReader raster.TileReader
	if len(*srcLayer) > 0 {
		tileReader, err = input.OpenTileLayer(*srcLayer)
	} else {
		tileReader, err = raster.OpenTileLayerAt(input, 0)
	}
	if err != nil {
		log.Fatal("Open layer: ", err)
	}

	log.Print("Connected to data set '", *src, "'")

	//Configure HTTP handlers
	http.Handle("/tiles/", &raster.Server{
		TileReader: tileReader,
		ZeroIsTop:  false,
	})
	http.Handle("/", mapPageHandler{
		Name:   *src,
		Reader: tileReader,
	})

	//Start HTTP server
	log.Print("Starting to listen on ", *addr)
	err = http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
