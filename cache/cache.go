// Copyright (c) 2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/protond/blob/master/LICENSE

package cache

import (
	"errors"

	"github.com/Supernomad/protond/common"
)

const (
	// NoopCache defines a cache that stores nothing and returns a hardcoded event or set of events.
	NoopCache = "noop"

	// MemoryCache defines an in memory cache.
	MemoryCache = "memory"
)

// Cache is the interface that plugins must adhere to for operation as a cache plugin.
type Cache interface {
	// Get should return a list of events associated with the given key, if there is an error during processing the list should be empty.
	Get(key string) []*common.Event

	// Store should add an event to an existing list of events or create a new one.
	Store(key string, event *common.Event)

	// Name returns the name of the cache plugin.
	Name() string
}

// New generates a cache plugin based on the passed in plugin and user defined configuration.
func New(cachePlugin string, config *common.Config, pluginConfig *common.PluginConfig) (Cache, error) {
	switch cachePlugin {
	case NoopCache:
		return newNoop(config)
	case MemoryCache:
		return newMemory(config, pluginConfig)
	}
	return nil, errors.New("specified cache plugin does not exist")
}
