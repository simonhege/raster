// Copyright 2015 Simon HEGE. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

//Command raster_init provides a CLI tool to create or update a tile data source (such as MBTiles or Geopackage database) based on an other data source (such as a ZXY server) and an area of interest.
//
//You can run it using
//		raster_init -src="http://example.com/%d/%d/%d.png" -dst="mbtiles.db" -dstdriver="mbtiles"
//
//Warning: see <http://wiki.openstreetmap.org/wiki/Tile_usage_policy> before using the OpenStreetMap servers. Bulk downloading is strongly discouraged.
//See also <http://wiki.openstreetmap.org/wiki/TMS> for a list of TMS server.
//
package main

import (
	"flag"
	"log"

	"github.com/cheggaaa/pb"
	_ "github.com/mattn/go-sqlite3"
	"github.com/paulsmith/gogeos/geos"

	_ "github.com/xeonx/raster/formats/gpkg"
	_ "github.com/xeonx/raster/formats/mbtiles"
	_ "github.com/xeonx/raster/formats/zxyserver"

	"github.com/xeonx/raster"
	"github.com/xeonx/raster/geosconverter"
)

var lvlmin = flag.Int("levelmin", 0, "zoom level")
var lvlmax = flag.Int("levelmax", 3, "zoom level")

var src = flag.String("src", "", "Source data source name")
var srcDriver = flag.String("srcdriver", "", "Source driver")
var srcLayer = flag.String("srclayer", "", "Source data source layer name")

var dst = flag.String("dst", "", "Destination data source name")
var dstDriver = flag.String("dstdriver", "", "Destination driver")
var dstLayer = flag.String("dstlayer", "data", "Destination data source layer name")

var aoi = flag.String("aoi", "POLYGON((-180 -85.0511, 180 -85.0511, 180 85.0511, -180 85.0511, -180 -85.0511))", "Area of interest (in WKT)")
var replace = flag.Bool("replace", false, "force replace of existing tiles")

type closer interface {
	Close() error
}

func main() {
	flag.Parse()

	poly := geos.Must(geos.FromWKT(*aoi))
	bbox, err := geosconverter.GetBoundingBox(poly)
	if err != nil {
		log.Fatal(err)
	}

	//Source
	if len(*srcDriver) == 0 {
		*srcDriver = raster.FindDriverName(*src)
	}
	input, err := raster.Open(*srcDriver, *src)
	if err != nil {
		log.Fatal(err)
	}
	if c, ok := input.(closer); ok {
		defer c.Close()
	}
	var inputReader raster.TileReader
	if len(*srcLayer) > 0 {
		inputReader, err = input.OpenTileLayer(*srcLayer)
	} else {
		inputReader, err = raster.OpenTileLayerAt(input, 0)
	}
	if err != nil {
		log.Fatal(err)
	}

	//Destination
	if len(*dstDriver) == 0 {
		*dstDriver = raster.FindDriverName(*dst)
	}
	outputReader, err := raster.Open(*dstDriver, *dst)
	if err != nil {
		log.Fatal(err)
	}
	output, ok := outputReader.(raster.WritableTileSource)
	if !ok {
		log.Fatal("Output driver does not allow writing")
	}
	if c, ok := output.(closer); ok {
		defer c.Close()
	}
	outputWriter, err := output.CreateTileLayer(*dstLayer)
	if err != nil {
		log.Fatal(err)
	}

	//Initialize the copy
	copier, err := raster.NewCopier(inputReader, outputWriter)
	if err != nil {
		log.Fatal(err)
	}

	polygonFilter := geosconverter.IntersectsFilter(poly)
	if *replace {
		copier.Filter = polygonFilter
	} else {
		copier.Filter = raster.Any(outputWriter.Contains, polygonFilter)
	}

	//Iterate on each requested level and performs the copy
	for level := *lvlmin; level <= *lvlmax; level++ {
		log.Print("Level: ", level)

		if *replace {
			log.Print("Level ", level, " clearing in database")
			err := outputWriter.Clear(level)
			if err != nil {
				log.Fatal(err)
			}
			log.Print("Level ", level, " cleared in database")
		}

		tiles, err := raster.GetTileBlock(bbox, level)
		if err != nil {
			log.Fatal(err)
		}
		log.Print("BBOX (tiles): ", tiles)

		log.Print("Nb tiles in BBOX: ", tiles.Count())

		bar := pb.StartNew(tiles.Count())

		processed, err := copier.CopyBlock(tiles, func(level, x, y int, processed bool) {
			bar.Increment()
		})
		if err != nil {
			log.Fatal(err)
		}
		bar.FinishPrint("End")
		log.Print("Nb tiles processed: ", processed)
	}

}
