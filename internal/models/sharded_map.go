package models

import (
	"hash/crc32"
	"sync"
)

// ShardedMap is a concurrent map implementation that uses sharding to improve performance.
type ShardedMap struct {
	shards []map[string][]ForestRoad
	locks  []sync.Mutex
	size   int
}

// GetFeaturesFromShardedMap returns a map of features from the sharded map.
// Keys are the SeNorge cluster coordinates, and values are slices of ForestRoad.
func (sm *ShardedMap) GetFeaturesFromShardedMap() map[string][]ForestRoad {
	result := make(map[string][]ForestRoad)

	for i := 0; i < sm.size; i++ {
		sm.locks[i].Lock()
		for key, features := range sm.shards[i] {
			result[key] = features
		}
		sm.locks[i].Unlock()
	}

	return result
}

// GetHashSetFromShardedMap returns a map of keys from the sharded map.
func (sm *ShardedMap) GetHashSetFromShardedMap() map[string]bool {
	result := make(map[string]bool)

	for i := 0; i < sm.size; i++ {
		sm.locks[i].Lock()
		for key := range sm.shards[i] {
			result[key] = true
		}
		sm.locks[i].Unlock()
	}

	return result
}

// NewShardedMap creates a new ShardedMap with the specified number of shards.
func NewShardedMap(size int) *ShardedMap {
	shards := make([]map[string][]ForestRoad, size)
	locks := make([]sync.Mutex, size)

	for i := range shards {
		shards[i] = make(map[string][]ForestRoad)
	}

	return &ShardedMap{shards: shards, locks: locks, size: size}
}

// Get retrieves the value associated with the given key from the sharded map.
func (sm *ShardedMap) Get(key string) ([]ForestRoad, bool) {
	idx := sm.hashKey(key)
	sm.locks[idx].Lock()
	defer sm.locks[idx].Unlock()
	val, ok := sm.shards[idx][key]
	return val, ok
}

// Set adds a value to the sharded map under the specified key.
func (sm *ShardedMap) Set(key string, value ForestRoad) {
	idx := sm.hashKey(key)
	sm.locks[idx].Lock()
	defer sm.locks[idx].Unlock()

	if _, exists := sm.shards[idx][key]; !exists {
		sm.shards[idx][key] = []ForestRoad{}
	}

	sm.shards[idx][key] = append(sm.shards[idx][key], value)
}

// hashKey computes the hash of the key and returns the index of the shard.
func (sm *ShardedMap) hashKey(key string) int {
	return int(crc32.ChecksumIEEE([]byte(key))) % sm.size
}
