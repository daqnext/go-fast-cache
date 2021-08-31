package meson_network_lts_local_cache

import (
	"github.com/daqnext/meson.network-lts-local-cache/sortedset"
	"log"
	"strconv"
	"testing"
	"unsafe"
)

func Test_ValSize(t *testing.T) {
	m := map[string]int{}
	key := []byte("123123123123123123123123123123123123123123123123123123")
	val := 1
	m[string(key)] = val
	log.Println(unsafe.Sizeof(m))
	log.Println(unsafe.Sizeof(key))
	log.Println(unsafe.Sizeof(val))
}

func Test_ZSet(t *testing.T) {
	s := sortedset.Make()
	for i := 0; i < 100; i++ {
		s.Add(strconv.Itoa(i), float64(i))
	}

	e := s.RangeByScore(0, 99, 0, -1, false)
	for _, v := range e {
		log.Println(v.Member, v.Score)
	}

}
