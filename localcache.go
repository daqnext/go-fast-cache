package meson_network_lts_local_cache

import (
	"github.com/liyiheng/zset"
	"sync"
	"time"
)

type LocalCache struct {
	CacheMap  sync.Map
	Zset      *zset.SortedSet
	SizeLimit uint64
	SizeUsed  uint64
	KeyCount  uint64
}

var localCache *LocalCache

func GetInstance() *LocalCache {
	if localCache == nil {
		localCache = &LocalCache{
			Zset: zset.New(),
		}
	}
	return localCache
}

func (lc *LocalCache) Get(key interface{}) (value interface{}, exist bool) {
	//check expire

	return lc.CacheMap.Load(key)
}

func (lc *LocalCache) Set(key interface{}, value interface{}, ttl time.Duration) {
	if ttl == 0 {
		//keep
	} else if ttl < 0 {
		//no expire
	} else {
		//new expire
	}
	lc.CacheMap.Store(key, value)
}

func (lc *LocalCache) SetEx(key interface{}, ttl time.Duration) {
	//set new expire
	if ttl == 0 {
		//keep
	} else if ttl < 0 {
		//no expire
	} else {
		//new expire
	}
}

// IsExist is key exist
func (lc *LocalCache) IsExist(key interface{}) bool {
	_, isExist := lc.CacheMap.Load(key)
	return isExist
}

// Remove remove a key
func (lc *LocalCache) Remove(key interface{}) {
	lc.CacheMap.Delete(key)
	//remove in zset

}

// TTL get ttl of a key with second, if <0 means no expire time
func (lc *LocalCache) TTL(key interface{}) float64 {
	return (time.Second * 1).Seconds()
}

// ScheduleDeleteExpire delete expired keys
func (lc *LocalCache) ScheduleDeleteExpire() {
	//get expired keys

	//delete keys in map

	//delete keys in zset

}
