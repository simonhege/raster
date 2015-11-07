// Copyright 2015 Simon HEGE. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package tilefolder

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

//TileFolder is a folder of tiles stored as .../level/x/y.format
type TileFolder struct {
	basePath   string
	tileFormat string
}

//NewTileFolder creates a new TileFolder based on the given configuration
func NewTileFolder(basePath string, tileFormat string) (TileFolder, error) {
	return TileFolder{
		basePath:   basePath,
		tileFormat: tileFormat,
	}, nil
}

//TileFormat exposes the image format of the source (png or jpg)
func (f TileFolder) TileFormat() string {
	return f.tileFormat
}

//GetPath returns the path for the given level/x/y.
func (f TileFolder) GetPath(level, x, y int) string {
	return path.Join(f.basePath, strconv.Itoa(level), strconv.Itoa(x), fmt.Sprintf("%d.%s", y, f.tileFormat))
}

//GetRaw retrieves the tile for a given level/x/y.
func (f TileFolder) GetRaw(level, x, y int) ([]byte, error) {
	path := f.GetPath(level, x, y)

	return ioutil.ReadFile(path)
}

//Contains returns true if the reader already contains the tile for a given level/x/y
func (f TileFolder) Contains(level int, x, y int) (bool, error) {
	path := f.GetPath(level, x, y)

	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

//SetRaw stores the tile for a given level/x/y. No check is performed on the image format.
func (f TileFolder) SetRaw(level, x, y int, img []byte) error {

	basePath := path.Join(f.basePath, strconv.Itoa(level), strconv.Itoa(x))
	if err := os.MkdirAll(basePath, 0666); err != nil {
		return err
	}

	path := f.GetPath(level, x, y)

	return ioutil.WriteFile(path, img, 0666)
}

//Clear removes all stored tiles at a given level.
func (f TileFolder) Clear(level int) error {

	return os.RemoveAll(path.Join(f.basePath, strconv.Itoa(level)))
}
