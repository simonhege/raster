// Copyright 2015 Simon HEGE. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/*
Package geosconverter ease usage of github.com/xeonx/raster in conjunction with github.com/paulsmith/gogeos/geos
*/
package geosconverter

import (
	"github.com/paulsmith/gogeos/geos"
	"github.com/xeonx/geographic"
	"github.com/xeonx/raster"
)

//GetBoundingBox computes the bounding box for a given geometry
func GetBoundingBox(g *geos.Geometry) (geographic.BoundingBox, error) {
	bbox := geographic.BoundingBox{}

	envelope, err := g.Envelope()
	if err != nil {
		return bbox, err
	}

	envelope, err = envelope.Shell()
	if err != nil {
		return bbox, err
	}

	centroid, err := envelope.Centroid()
	if err != nil {
		return bbox, err
	}

	bbox.LatitudeMinDeg, _ = centroid.Y()
	bbox.LongitudeMinDeg, _ = centroid.X()
	bbox.LatitudeMaxDeg = bbox.LatitudeMinDeg
	bbox.LongitudeMaxDeg = bbox.LongitudeMinDeg

	npoints, err := envelope.NPoint()
	if err != nil {
		return bbox, err
	}
	for i := 0; i < npoints; i++ {
		pt := geos.Must(envelope.Point(i))
		lon, _ := pt.X()
		lat, _ := pt.Y()

		if lon < bbox.LongitudeMinDeg {
			bbox.LongitudeMinDeg = lon
		}
		if lon > bbox.LongitudeMaxDeg {
			bbox.LongitudeMaxDeg = lon
		}
		if lat < bbox.LatitudeMinDeg {
			bbox.LatitudeMinDeg = lat
		}
		if lat > bbox.LatitudeMaxDeg {
			bbox.LatitudeMaxDeg = lat
		}
	}

	return bbox, nil
}

//IntersectsFilter creates a Filter allowing to skip tiles outside of a given geometry
func IntersectsFilter(g *geos.Geometry) raster.Filter {
	//TODO: use a prepared geometry

	return func(level, x, y int) (bool, error) {

		tile, err := newTilePolygon(level, x, y)
		if err != nil {
			return false, err
		}
		intersects, err := g.Intersects(tile)
		if err != nil {
			return false, err
		}

		return !intersects, nil //Tiles that do not intersect are excluded
	}
}

//newTilePolygon creates the geos polygon for a given tile.
func newTilePolygon(level, x, y int) (*geos.Geometry, error) {

	latDeg := raster.Y2Lat(level, y)
	lonDeg := raster.X2Lon(level, x)
	lat2Deg := raster.Y2Lat(level, y+1)
	lon2Deg := raster.X2Lon(level, x+1)

	return geos.NewPolygon(
		[]geos.Coord{
			geos.NewCoord(lonDeg, latDeg),
			geos.NewCoord(lon2Deg, latDeg),
			geos.NewCoord(lon2Deg, lat2Deg),
			geos.NewCoord(lonDeg, lat2Deg),
			geos.NewCoord(lonDeg, latDeg),
		})
}
