// Copyright 2015 Simon HEGE. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package zxyserver

import (
	"strings"

	"github.com/xeonx/raster"
)

func init() {
	raster.Register("zxy", newSingleLayerDriver(func(dataSourceName string) raster.TileReader {
		return ZxyServer{URL: dataSourceName}
	}, func(dataSourceName string) bool {
		if !strings.HasPrefix(dataSourceName, "http") {
			return false
		}
		if strings.Count(dataSourceName, "%d") != 3 {
			return false
		}
		return true
	}))
}

func newSingleLayerDriver(createTileReader func(dataSourceName string) raster.TileReader, canOpen func(dataSourceName string) bool) raster.Driver {
	return singleLayerDriver{
		createTileReader: createTileReader,
		canOpen:          canOpen,
	}
}

type singleLayerDriver struct {
	canOpen          func(dataSourceName string) bool
	createTileReader func(dataSourceName string) raster.TileReader
}

func (d singleLayerDriver) OpenTileSource(dataSourceName string) (raster.TileSource, error) {
	return singleLayerSource{
		TileReader: d.createTileReader(dataSourceName),
		Name:       dataSourceName,
	}, nil
}
func (d singleLayerDriver) CanOpen(dataSourceName string) bool {
	return d.canOpen(dataSourceName)
}

type singleLayerSource struct {
	raster.TileReader
	Name string
}

func (s singleLayerSource) ListTileLayers() ([]string, error) {
	var layers []string
	layers = append(layers, s.Name)
	return layers, nil
}
func (s singleLayerSource) OpenTileLayer(name string) (raster.TileReader, error) {
	return s, nil
}
