// Copyright 2015 Simon HEGE. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package raster

import (
	"errors"
	"fmt"
	"sort"
	"sync"
)

//FindDriverName returns the name of a driver able to import the given data source
func FindDriverName(dataSourceName string) string {

	driverNames := Drivers()

	for _, driverName := range driverNames {
		driversMu.Lock()
		driveri, ok := drivers[driverName]
		driversMu.Unlock()
		if !ok {
			continue
		}

		if driveri.CanOpen(dataSourceName) {
			return driverName
		}
	}

	return ""
}

//Driver is the interface that must be implemented by a tile data source driver.
type Driver interface {
	CanOpen(dataSourceName string) bool
	OpenTileSource(dataSourceName string) (TileSource, error)
}

var (
	driversMu sync.Mutex
	drivers   = make(map[string]Driver)
)

//ErrLayerNotFund is returned by OpenTileLayer when the requested layer is not found in the source
var ErrLayerNotFund = errors.New("raster: layer not found")

// Register makes a tile source driver available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func Register(name string, driver Driver) {
	driversMu.Lock()
	defer driversMu.Unlock()
	if driver == nil {
		panic("raster: Register driver is nil")
	}
	if _, dup := drivers[name]; dup {
		panic("raster: Register called twice for driver " + name)
	}
	drivers[name] = driver
}

// Drivers returns a sorted list of the names of the registered drivers.
func Drivers() []string {
	driversMu.Lock()
	defer driversMu.Unlock()
	var list []string
	for name := range drivers {
		list = append(list, name)
	}
	sort.Strings(list)
	return list
}

// Open opens a tile source specified by its tile source driver name and a
// driver-specific data source name, usually consisting of at least a
// tile source name and connection information.
func Open(driverName, dataSourceName string, options ...func(*TileSource) error) (TileSource, error) {

	driversMu.Lock()
	driveri, ok := drivers[driverName]
	driversMu.Unlock()
	if !ok {
		return nil, fmt.Errorf("raster: unknown driver %q (forgotten import?)", driverName)
	}

	ts, err := driveri.OpenTileSource(dataSourceName)
	if err != nil {
		return nil, err
	}

	for _, o := range options {
		if err := o(&ts); err != nil {
			return nil, err
		}
	}

	return ts, err
}

//TileSource is the interface that must be implemented to access multiple tiled data.
//
//A TileSource must be safe for use by multiple goroutines.
type TileSource interface {
	//ListTileLayers list all available tile layers
	ListTileLayers() ([]string, error)

	//OpenTileLayer opens the tile layer for reading
	OpenTileLayer(name string) (TileReader, error)
}

//OpenTileLayerAt opens the nth layer of a given source
func OpenTileLayerAt(source TileSource, n int) (TileReader, error) {
	layers, err := source.ListTileLayers()
	if err != nil {
		return nil, err
	}

	if n < 0 || n >= len(layers) {
		return nil, ErrLayerNotFund
	}

	return source.OpenTileLayer(layers[n])
}

//WritableTileSource is the interface that must be implemented to allow writing of tiled data
//
//A WritableTileSource must be safe for use by multiple goroutines.
type WritableTileSource interface {
	TileSource
	//CreateTileLayer creates a layer in the data source, or opens it for writing
	CreateTileLayer(name string) (TileReadWriter, error)
}

//TileReader represents any service able to provide a tile for given level, x and y.
//
//A TileReader must be safe for use by multiple goroutines.
type TileReader interface {
	//TileFormat exposes the image format of the source (png or jpg)
	TileFormat() string
	//GetRaw retrieves the tile for a given level/x/y.
	GetRaw(level, x, y int) ([]byte, error)
	//Contains returns true if the reader already contains the tile for a given level/x/y
	Contains(level int, x, y int) (bool, error)
}

//TileReadWriter represents any service able to save a tile for given level, x and y.
//
//A TileReader must be safe for use by multiple goroutines.
type TileReadWriter interface {
	TileReader
	//SetRaw stores the tile for a given level/x/y. No check is performed on the image format.
	SetRaw(level, x, y int, img []byte) error
	//Clear removes all stored tiles at a given level.
	Clear(level int) error
}
