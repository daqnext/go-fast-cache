package meson_network_lts_local_cache

import (
	"log"
	"testing"
	"time"
)

func Test_cache(t *testing.T) {

	GetInstance().Set("a", "aaa", time.Second*5)
	GetInstance().Set("b", "bbb", time.Second*7)
	GetInstance().Set("c", "ccc", time.Second*9)
	GetInstance().Set("d", "ddd", time.Second*11)

	for {
		v, exist := GetInstance().Get("a")
		log.Println(GetInstance().TTL("a"))
		log.Println(v)
		log.Println(exist)
		time.Sleep(1 * time.Second)
	}
}
