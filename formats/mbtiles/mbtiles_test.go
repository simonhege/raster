// Copyright 2014-2015 Simon HEGE. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package mbtiles

import (
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

var dataToTMS = []struct {
	longitudeDeg float64
	latitudeDeg  float64
	level        int
	x            int
	y            int
}{
	{7.27, 47.4, 0, 0, 0},
	{7.27, 47.4, 1, 1, 1},
	{7.27, 47.4, 2, 2, 2},
	{7.27, 47.4, 3, 4, 5},
	{7.27, 47.4, 4, 8, 10},
	{7.27, 47.4, 5, 16, 20},
	{7.27, 47.4, 6, 33, 41},
}

func TestConvertToTMS(t *testing.T) {

	for _, tt := range dataToTMS {

		outX := Lon2X(tt.level, tt.longitudeDeg)
		if outX != tt.x {
			t.Errorf("Lon2X(%d, %f) => %d, want %d", tt.level, tt.longitudeDeg, outX, tt.x)
		}
		outY := Lat2Y(tt.level, tt.latitudeDeg)
		if outY != tt.y {
			t.Errorf("Lat2Y(%d, %f) => %d, want %d", tt.level, tt.latitudeDeg, outY, tt.y)
		}
	}
}

var dataToLatLon = []struct {
	longitudeDeg float64
	latitudeDeg  float64
	level        int
	x            int
	y            int
}{
	{-180, -85.05112947, 0, 0, 0},
	{0, 0, 1, 1, 1},
	{0, 0, 2, 2, 2},
	{0, 40.979898, 3, 4, 5},
	{0, 40.979898, 4, 8, 10},
	{0, 40.979898, 5, 16, 20},
	{5.625, 45.089036, 6, 33, 41},
}

func TestConvertToTLatLon(t *testing.T) {

	for _, tt := range dataToLatLon {
		outLat := Y2Lat(tt.level, tt.y)
		if (outLat-tt.latitudeDeg)*(outLat-tt.latitudeDeg) > 1e-5 {
			t.Errorf("Y2Lat(%d, %d) => %f, want %f", tt.level, tt.y, outLat, tt.latitudeDeg)
		}

		outLon := X2Lon(tt.level, tt.x)
		if outLon != tt.longitudeDeg {
			t.Errorf("X2Lon(%d, %d) => %f, want %f", tt.level, tt.x, outLon, tt.longitudeDeg)
		}
	}

}
