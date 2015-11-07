// Copyright 2015 Simon HEGE. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package tilefolder

import (
	"os"
	"path"
	"path/filepath"

	"github.com/xeonx/raster"
)

func init() {
	raster.Register("folder", tileFolderDriver{format: "png"})
	raster.Register("jpgfolder", tileFolderDriver{format: "jpg"})
}

type tileFolderDriver struct {
	format string
}

func (d tileFolderDriver) CanOpen(dataSourceName string) bool {
	s, err := os.Stat(dataSourceName)
	if err != nil {
		return false
	}

	if !s.IsDir() {
		return false
	}

	return true
}
func (d tileFolderDriver) OpenTileSource(dataSourceName string) (raster.TileSource, error) {
	return tileFolderSource{dataSourceName: dataSourceName, format: d.format}, nil
}

type tileFolderSource struct {
	dataSourceName string
	format         string
}

func (s tileFolderSource) ListTileLayers() ([]string, error) {
	var layers []string

	err := filepath.Walk(s.dataSourceName, func(path string, info os.FileInfo, err error) error {

		if info.IsDir() {
			layers = append(layers, info.Name())
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return layers, nil
}
func (s tileFolderSource) OpenTileLayer(name string) (raster.TileReader, error) {
	return NewTileFolder(path.Join(s.dataSourceName, name), s.format)
}
func (s tileFolderSource) CreateTileLayer(name string) (raster.TileReadWriter, error) {
	return NewTileFolder(path.Join(s.dataSourceName, name), s.format)
}
