package models

import (
	"hash/crc32"
	"sync"
)

type ShardedMap struct {
	shards []map[string][]ForestRoad
	locks  []sync.Mutex
	size   int
}

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

func NewShardedMap(size int) *ShardedMap {
	shards := make([]map[string][]ForestRoad, size)
	locks := make([]sync.Mutex, size)

	for i := range shards {
		shards[i] = make(map[string][]ForestRoad)
	}

	return &ShardedMap{shards: shards, locks: locks, size: size}
}

func (sm *ShardedMap) Get(key string) ([]ForestRoad, bool) {
	idx := sm.hashKey(key)
	sm.locks[idx].Lock()
	defer sm.locks[idx].Unlock()
	val, ok := sm.shards[idx][key]
	return val, ok
}

func (sm *ShardedMap) Set(key string, value ForestRoad) {
	idx := sm.hashKey(key)
	sm.locks[idx].Lock()
	defer sm.locks[idx].Unlock()

	if _, exists := sm.shards[idx][key]; !exists {
		sm.shards[idx][key] = []ForestRoad{}
	}

	sm.shards[idx][key] = append(sm.shards[idx][key], value)
}

func (sm *ShardedMap) hashKey(key string) int {
	return int(crc32.ChecksumIEEE([]byte(key))) % sm.size
}
