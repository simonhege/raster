// Copyright 2015 Simon HEGE. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package zxyserver

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
)

//ZxyServer is the TileReader for OpenStreetMap like servers.
//URL must contains three '%d' indicating where the level, x and y will be placed (in this order).
//
//See http://wiki.openstreetmap.org/wiki/Tile_usage_policy before using the OpenStreetMap servers.
type ZxyServer struct {
	URL string //eg.: http://a.tile.openstreetmap.org/%d/%d/%d.png
}

//TileFormat exposes the image format of the source (png or jpg)
func (r ZxyServer) TileFormat() string {
	ext := path.Ext(r.URL)
	if len(ext) > 0 && ext[0] == '.' {
		ext = ext[1:]
	}
	return ext
}

//GetURL returns the URL for the given level/x/y.
func (r ZxyServer) GetURL(level, x, y int) string {
	var ymax = 1 << uint(level)
	var yosm = ymax - y - 1

	return fmt.Sprintf(r.URL, level, x, yosm)
}

//GetRaw retrieves the tile for a given level/x/y.
func (r ZxyServer) GetRaw(level, x, y int) ([]byte, error) {
	url := r.GetURL(level, x, y)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	rawImg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return rawImg, nil
}

//Contains returns true if the reader already contains the tile for a given level/x/y
func (r ZxyServer) Contains(level int, x, y int) (bool, error) {
	url := r.GetURL(level, x, y)

	resp, err := http.Head(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return true, nil
}
