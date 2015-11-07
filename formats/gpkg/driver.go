// Copyright 2015 Simon HEGE. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gpkg

import (
	"path/filepath"

	"github.com/xeonx/raster"
)

func init() {
	raster.Register("gpkg", gpkgDriver{})
}

type gpkgDriver struct {
}

func (d gpkgDriver) OpenTileSource(dataSourceName string) (raster.TileSource, error) {
	return Open(dataSourceName)
}
func (d gpkgDriver) CanOpen(dataSourceName string) bool {
	if filepath.Ext(dataSourceName) == ".gpkg" {
		return true
	}
	return false
}
