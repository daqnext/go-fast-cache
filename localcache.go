package meson_network_lts_local_cache

import (
	"github.com/daqnext/meson.network-lts-local-cache/sortedset"
	"github.com/daqnext/meson.network-lts-local-cache/ttltype"
	"time"
)

type LocalCache struct {
	S *sortedset.SortedSet
}

var localCache *LocalCache

func GetInstance() *LocalCache {
	if localCache == nil {
		localCache = &LocalCache{
			S: sortedset.Make(),
		}
	}
	return localCache
}

//func (lc *LocalCache) SetCountLimit(limit uint){
//	lc.CountLimit=limit
//}

func (lc *LocalCache) Get(key interface{}) (value interface{}, exist bool) {
	//check expire
	e, exist := lc.S.Get(key)
	if !exist {
		return nil, false
	}
	if e.Score < time.Now().Unix() {
		return nil, false
	}
	return e.Value, true
}

// Set Set key value with expire time, ttl.Keep,ttl.Infinity,or time.Duration. if key not exist and set ttl ttl.Keep,it will use default ttl 5min
func (lc *LocalCache) Set(key interface{}, value interface{}, ttl time.Duration) {

	expireTime := int64(ttltype.Infinity)

	if ttl == ttltype.Keep {
		//keep
		var exist bool
		expireTime, exist = lc.TTL(key)
		if !exist {
			expireTime = time.Now().Add(time.Minute * 5).Unix()
		}
	} else if ttl < ttltype.Infinity {
		//no expire
		expireTime = -1
	} else {
		//new expire
		expireTime = time.Now().Add(ttl).Unix()
	}
	lc.S.Add(key.(string), expireTime, value)
}

//func (lc *LocalCache) SetEx(key interface{}, ttl time.Duration) bool {
//	if !lc.IsExist(key){
//		return false
//	}
//
//	expireTime:=int64(ttltype.Infinity)
//
//	if ttl == ttltype.Keep {
//		//keep
//		var exist bool
//		expireTime,exist=lc.TTL(key)
//		if !exist {
//			expireTime=time.Now().Add(time.Minute*5).Unix()
//		}
//	} else if ttl < ttltype.Infinity {
//		//no expire
//		expireTime=-1
//	} else {
//		//new expire
//		expireTime=time.Now().Add(ttl).Unix()
//	}
//	lc.S.Add(key.(string),expireTime,value)
//}

// IsExist is key exist
func (lc *LocalCache) IsExist(key interface{}) bool {
	//check expire
	e, exist := lc.S.Get(key.(string))
	if !exist {
		return false
	}
	if e.Score < time.Now().Unix() {
		return false
	}
	return true
}

// Remove remove a key
func (lc *LocalCache) Remove(key interface{}) {
	lc.S.Remove(key.(string))
}

// TTL get ttl of a key with second, if <0 means no expire time
func (lc *LocalCache) TTL(key interface{}) (int64, bool) {
	e, exist := lc.S.Get(key.(string))
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
func (lc *LocalCache) ScheduleDeleteExpire(interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			min := int64(0)
			max := time.Now().Unix()
			//get expired keys
			lc.S.RemoveByScore(min, max)
		}
	}()
}
