// Copyright 2014-2015 Simon HEGE. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package mbtiles

import (
	"path/filepath"

	"github.com/xeonx/raster"
)

func init() {
	raster.Register("mbtiles", mbtilesDriver{})
}

type mbtilesDriver struct {
}

func (d mbtilesDriver) OpenTileSource(dataSourceName string) (raster.TileSource, error) {
	return Open(dataSourceName)
}

func (d mbtilesDriver) CanOpen(dataSourceName string) bool {
	if filepath.Ext(dataSourceName) == ".mbtiles" {
		return true
	}
	return false
}
