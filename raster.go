// Copyright 2015 Simon HEGE. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/*
Package raster provides raster/tiles manipulation helpers.

Conversion between latitude/longitude and x/y in global-mercator is performed as described
in http://wiki.openstreetmap.org/wiki/Slippy_map_tilenames
*/
package raster

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"math"

	"github.com/xeonx/geographic"
)

//n returns 2^level
func n(level int) int {
	return 1 << uint(level)
}

//X2Lon transforms x into a longitude in degree at a given level.
func X2Lon(level int, x int) float64 {
	return float64(x)/float64(n(level))*360.0 - 180.0
}

//Lon2X transforms a longitude in degree into x at a given level.
func Lon2X(level int, longitudeDeg float64) int {
	return int(float64(n(level)) * (longitudeDeg + 180.) / 360.)
}

//Y2Lat transforms y into a latitude in degree at a given level.
func Y2Lat(level int, y int) float64 {
	var yosm = y

	latitudeRad := math.Atan(math.Sinh(math.Pi * (1. - 2.*float64(yosm)/float64(n(level)))))
	return -(latitudeRad * 180.0 / math.Pi)
}

//Lat2Y transforms a latitude in degree into y at a given level.
func Lat2Y(level int, latitudeDeg float64) int {
	latitudeRad := latitudeDeg / 180 * math.Pi
	yosm := int(float64(n(level)) * (1. - math.Log(math.Tan(latitudeRad)+1/math.Cos(latitudeRad))/math.Pi) / 2.)

	return n(level) - yosm - 1
}

//Encode encodes an image in the given format. Only jpg and png are supported.
func Encode(img image.Image, format string) ([]byte, error) {

	var b bytes.Buffer
	if format == "jpg" {
		err := jpeg.Encode(&b, img, nil)
		if err != nil {
			return nil, err
		}
	} else if format == "png" {
		err := png.Encode(&b, img)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("Unsupported image format: '" + format + "'. Only 'jpg' or 'png' allowed.")
	}

	return b.Bytes(), nil
}

//Decode decodes an image into the given format. Only jpg and png are supported.
func Decode(rawImg []byte, format string) (image.Image, error) {

	b := bytes.NewBuffer(rawImg)

	var img image.Image
	var err error
	if format == "jpg" {
		img, err = jpeg.Decode(b)
		if err != nil {
			return nil, err
		}
	} else if format == "png" {
		img, err = png.Decode(b)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("Unsupported image format: '" + format + "'. Only 'jpg' or 'png' allowed.")
	}

	return img, nil
}

//Filter returns true if the given tile is excluded from copy
type Filter func(level, x, y int) (bool, error)

//Any composes Filter in a single Filter excluding a tile if at least one Filter excludes the tile.
func Any(filters ...Filter) Filter {
	return func(level, x, y int) (bool, error) {
		for _, f := range filters {
			b, err := f(level, x, y)
			if err != nil {
				return false, err
			}
			if b {
				return true, nil
			}
		}
		return false, nil
	}
}

//All composes Filter in a single Filter excluding a tile if all Filters excludes the tile.
func All(filters ...Filter) Filter {
	return func(level, x, y int) (bool, error) {
		for _, f := range filters {
			b, err := f(level, x, y)
			if err != nil {
				return false, err
			}
			if !b {
				return false, nil
			}
		}
		return true, nil
	}
}

//transformFct is any transformation function usable inside the encoding or decoding of images.
type transformFct func(src []byte) ([]byte, error)

//GetTansformFct returns a function able to convert between the raw format of from and the one of to.
//A nil value means that no transformation is required.
func getTansformFct(from TileReader, to TileReadWriter) transformFct {
	var transform transformFct
	if from.TileFormat() != to.TileFormat() {
		transform = func(src []byte) ([]byte, error) {
			img, err := Decode(src, from.TileFormat())
			if err != nil {
				return nil, err
			}
			return Encode(img, to.TileFormat())
		}
	}
	return transform
}

//Copier copies tiles from a TileReader to a TileReadWriter.
//An optional filter allow to discard Tiles before copy.
type Copier struct {
	from      TileReader
	to        TileReadWriter
	transform transformFct

	Filter Filter
}

//NewCopier creates a Copier between from and to.
func NewCopier(from TileReader, to TileReadWriter) (*Copier, error) {
	return &Copier{
		from:      from,
		to:        to,
		transform: getTansformFct(from, to),
	}, nil
}

//TileBlock represents a rectangular set of tiles at a given level
type TileBlock struct {
	Level int
	Xmin  int
	Xmax  int
	Ymin  int
	Ymax  int
}

//Count returns the number of tiles within the block.
func (b TileBlock) Count() int {
	return (b.Xmax - b.Xmin + 1) * (b.Ymax - b.Ymin + 1)
}

//GetTileBlock computes the tile block enveloping the bounding box
func GetTileBlock(bbox geographic.BoundingBox, level int) (TileBlock, error) {
	b := TileBlock{
		Level: level,
		Xmin:  Lon2X(level, bbox.LongitudeMinDeg),
		Xmax:  Lon2X(level, bbox.LongitudeMaxDeg),
		Ymin:  Lat2Y(level, bbox.LatitudeMinDeg),
		Ymax:  Lat2Y(level, bbox.LatitudeMaxDeg),
	}

	if b.Ymin > b.Ymax {
		y := b.Ymin
		b.Ymin = b.Ymax
		b.Ymax = y
	}

	if X2Lon(level, b.Xmax) == bbox.LongitudeMaxDeg {
		b.Xmax--
	}

	return b, nil
}

//CopyBlock copies a block of tiles.
//If progressFct is not nil, it is called during the iteration after each tile.
//It returns the count of tiles copied in the destination and the first error encountered, if any.
func (c *Copier) CopyBlock(block TileBlock, progressFct func(level, x, y int, processed bool)) (int, error) {
	processedCount := 0
	for x := block.Xmin; x <= block.Xmax; x++ {
		for y := block.Ymin; y <= block.Ymax; y++ {
			processed, err := c.Copy(block.Level, x, y)
			if err != nil {
				return processedCount, err
			}
			if progressFct != nil {
				progressFct(block.Level, x, y, processed)
			}
			if processed {
				processedCount++
			}
		}
	}
	return processedCount, nil
}

//Copy copies a single of tile.
//It returns the true if the tile was copied in the destination and the first error encountered, if any.
func (c *Copier) Copy(level, x, y int) (bool, error) {

	if c.Filter != nil {
		filtered, err := c.Filter(level, x, y)
		if err != nil {
			return false, err
		}
		if filtered {
			return false, nil
		}
	}

	rawImg, err := c.from.GetRaw(level, x, y)
	if err != nil {
		return false, err
	}

	if c.transform != nil {
		rawImg, err = c.transform(rawImg)
		if err != nil {
			return false, err
		}
	}

	err = c.to.SetRaw(level, x, y, rawImg)
	if err != nil {
		return false, err
	}

	return true, nil
}

//Copy copies a single tile from a reader to a writer.
//It returns the true if the tile was copied in the destination and the first error encountered, if any.
func Copy(from TileReader, to TileReadWriter, level, x, y int) (bool, error) {
	c, err := NewCopier(from, to)
	if err != nil {
		return false, err
	}
	return c.Copy(level, x, y)
}
