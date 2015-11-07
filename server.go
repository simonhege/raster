// Copyright 2015 Simon HEGE. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package raster

import (
	"log"
	"net/http"
	"regexp"
	"strconv"

	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
)

var urlRegex = regexp.MustCompile(`\A/.*/(\d+)/(\d+)/(\d+)\.(png|jpeg)\z`)

var grayTile []byte

func init() {
	m := image.NewRGBA(image.Rect(0, 0, 256, 256))
	gray := color.RGBA{96, 96, 96, 255}
	draw.Draw(m, m.Bounds(), &image.Uniform{gray}, image.ZP, draw.Src)

	w := new(bytes.Buffer)
	if err := png.Encode(w, m); err != nil {
		log.Fatal(err)
	}
	grayTile = w.Bytes()
}

//Server allows serving tiles through HTTP on URLS like
//http://example.com/any/sub/path/level/x/y.ext where level, x and y are the tile
//identification integers and ext is the requested file format (png or jpeg).
//
//By default, Server does not follow the OSM convention for y value but the MBTiles
//convention.
type Server struct {
	TileReader TileReader
	ZeroIsTop  bool //Flag indicating if the server follow the OSM convention (0,0 is top-left) instead of TMS/MBTiles convention
}

//ServeHTTP implements net/http.Handler
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	//Split URL
	m := urlRegex.FindStringSubmatch(r.URL.Path)
	if m == nil || len(m) != 5 {
		log.Print("Invalid URL: ", r.URL.Path, m)
		http.NotFound(w, r)
		return
	}

	//Decode level, x and y
	level, err := strconv.Atoi(m[1])
	if err != nil {
		log.Print("Error decoding level: ", m[1])
		http.NotFound(w, r)
		return
	}
	x, err := strconv.Atoi(m[2])
	if err != nil {
		log.Print("Error decoding x: ", m[2])
		http.NotFound(w, r)
		return
	}
	y, err := strconv.Atoi(m[3])
	if err != nil {
		log.Print("Error decoding y: ", m[3])
		http.NotFound(w, r)
		return
	}
	//TODO: m[4] is requested extension (jpeg or png). If it is not the stored format, we can convert on the fly.

	//Reverse y to follow osm conventions
	if s.ZeroIsTop {
		y = (1 << uint(level)) - y - 1
	}

	tileData, err := s.TileReader.GetRaw(level, x, y)
	if err != nil {
		log.Print("Error: ", level, x, y, err)
		http.NotFound(w, r)
		return
	}
	if tileData == nil {
		log.Print("Not found: ", level, x, y, err)
		tileData = grayTile
	}
	w.Write(tileData)
}
