package meson_network_lts_local_cache

import (
	"github.com/daqnext/meson.network-lts-local-cache/sortedset"
	"github.com/daqnext/meson.network-lts-local-cache/ttltype"
	"time"
)

type localCache struct {
	s          *sortedset.SortedSet
	countLimit int64
}

// New instance of localCache, param intervalSecond defines the interval of scheduleDeleteExpire job, if intervalSecond <=0,it will use the default value 5 seconds
func New(intervalSecond int) *localCache {
	cache := &localCache{
		s:          sortedset.Make(),
		countLimit: 100000,
	}
	cache.scheduleDeleteExpire(intervalSecond)
	return cache
}

func (lc *localCache) SetCountLimit(limit int64) {
	lc.countLimit = limit
}

func (lc *localCache) Get(key string) (value interface{}, ttl int64, exist bool) {
	//check expire
	e, exist := lc.s.Get(key)
	if !exist {
		return nil, 0, false
	}
	if e.Score < time.Now().Unix() {
		return nil, 0, false
	}
	return e.Value, e.Score - time.Now().Unix(), true
}

// Set Set key value with expire time, ttl.Keep or second. if key not exist and set ttl ttl.Keep,it will use default ttl 5min
func (lc *localCache) Set(key string, value interface{}, ttlSecond int64) {
	currentCount := lc.s.Len()
	//delete 20%
	deleteCount := currentCount / 5
	if deleteCount < 1 {
		deleteCount = 1
	}
	if currentCount >= lc.countLimit {
		lc.s.RemoveByRank(0, deleteCount)
	}

	if ttlSecond > 7200 {
		ttlSecond = 7200
	}
	expireTime := int64(0)

	if ttlSecond == ttltype.Keep {
		//keep
		var exist bool
		expireTime, exist = lc.ttl(key)
		if !exist {
			expireTime = time.Now().Unix() + 300
		}
	} else {
		if ttlSecond < 1 {
			return
		}
		//new expire
		expireTime = time.Now().Unix() + ttlSecond
	}
	lc.s.Add(key, expireTime, value)
}

// IsExist is key exist
//func (lc *localCache) IsExist(key string) bool {
//	//check expire
//	e, exist := lc.s.Get(key)
//	if !exist {
//		return false
//	}
//	if e.Score < time.Now().Unix() {
//		return false
//	}
//	return true
//}

// Remove remove a key
//func (lc *localCache) Remove(key string) {
//	lc.s.Remove(key)
//}

// TTL get ttl of a key with second
func (lc *localCache) ttl(key string) (int64, bool) {
	e, exist := lc.s.Get(key)
	if !exist {
		return 0, false
	}
	ttl := e.Score - time.Now().Unix()
	if ttl < 0 {
		return -1, true
	}
	return ttl, true
}

// ScheduleDeleteExpire delete expired keys
func (lc *localCache) scheduleDeleteExpire(intervalSecond int) {
	if intervalSecond <= 0 {
		intervalSecond = 5
	}
	interval := time.Second * time.Duration(intervalSecond)
	go func() {
		for {
			time.Sleep(interval)
			min := int64(0)
			max := time.Now().Unix()
			//get expired keys
			lc.s.RemoveByScore(min, max)
		}
	}()
}
