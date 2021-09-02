package test

import (
	localcache "github.com/daqnext/meson.network-lts-local-cache"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"
)

type Person struct {
	Name     string
	Age      int
	Location string
}

func printMemStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log.Printf("Alloc = %v KB, TotalAlloc = %v KB, Sys = %v KB,Lookups = %v NumGC = %v\n", m.Alloc/1024, m.TotalAlloc/1024, m.Sys/1024, m.Lookups, m.NumGC)
}

func Test_Set(t *testing.T) {
	lc := localcache.NewWithInterval(1)
	for i := 0; i < 100000; i++ {
		lc.Set(strconv.Itoa(i), "aaaaaaaaaaaaaaaaaaaaaaa", 60)
	}

	for i := 0; i < 100000; i += 1000 {
		log.Println(lc.Get(strconv.Itoa(i)))
	}

}

func Test_Get(t *testing.T) {
	lc := localcache.New()
	for i := 0; i < 100000; i++ {
		lc.Set(strconv.Itoa(i), "aaaaaaaaaaaaaaaaaaaaaaa", 60)
	}

	for i := 0; i < 110000; i += 1000 {
		v, ttl, exist := lc.Get(strconv.Itoa(i))
		if !exist {
			log.Println("key:", strconv.Itoa(i), "not exist")
		} else {
			log.Println(strconv.Itoa(i), v, ttl)
		}
	}
}

func Test_Expire(t *testing.T) {
	lc := localcache.New()
	lc.Set("1", "111", 5)
	lc.Set("2", "111", 18)
	lc.Set("3", "111", 23)
	lc.Set("4", "111", -100)
	lc.Set("5", "111", 3000000)
	lc.Set("6", "111", 35)

	count := 0
	for {
		log.Println(lc.Get("1"))
		log.Println(lc.Get("2"))
		log.Println(lc.Get("3"))
		log.Println(lc.Get("4"))
		log.Println(lc.Get("5"))
		log.Println(lc.Get("6"))
		log.Println("-----------")
		count++
		if count > 40 {
			return
		}
		time.Sleep(time.Second)
	}
}

func Test_SetRemove(t *testing.T) {
	a := Person{"Jack", 18, "America"}
	lc := localcache.NewWithInterval(1)
	lc.SetCountLimit(10000)

	log.Println("start")
	printMemStats()

	for i := 0; i < 100; i++ {
		//set
		for j := 0; j < 100; j++ {
			lc.Set(strconv.Itoa(j), a, 1)
		}

		log.Println("round:", i)
		log.Println("finish set")
		printMemStats()

		time.Sleep(2 * time.Second)
	}

	log.Println("finish")
	printMemStats()
}

func Test_BigAmountKey(t *testing.T) {
	a := Person{"Jack", 18, "America"}
	lc := localcache.New()
	lc.SetCountLimit(1000000)

	printMemStats()

	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:10000", nil))
	}()

	go func() {
		for {
			time.Sleep(5 * time.Second)
			printMemStats()
		}
	}()

	printMemStats()
	for i := 0; i < 1000000; i++ {
		lc.Set(strconv.Itoa(i), a, int64(rand.Intn(10)+1))
	}

	go func() {
		for i := 0; i < 100; i++ {
			//time.Sleep(20*time.Second)
			log.Println("Start round", i+1)
			for i := 0; i < 1000000; i++ {
				lc.Set(strconv.Itoa(i), a, int64(rand.Intn(10)+1))
			}
		}

	}()

	printMemStats()

	count := 0
	for {
		time.Sleep(1 * time.Second)
		//log.Println("count length", lc.getLen())
		//log.Println("skiplist length", lc.s.SLen())
		//log.Println("map length", lc.s.MapLen())
		count++
		if count > 45 {
			//return
		}
	}
	//time.Sleep(1*time.Hour)
}

func Test_SyncMap(t *testing.T) {
	printMemStats()

	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:10000", nil))
	}()

	type Person struct {
		Name     string
		Age      int
		Location string
	}
	a := Person{"Jack", 18, "America"}
	type Element struct {
		Member string
		Score  int64
		Value  interface{}
	}

	var myMap sync.Map
	for i := 0; i < 1000000; i++ {
		key := strconv.Itoa(i)
		b := &Element{
			Member: key,
			Score:  10,
			Value:  a,
		}
		myMap.Store(key, b)
	}

	printMemStats()

	for i := 0; i < 1000000; i++ {
		myMap.Delete(strconv.Itoa(i))
	}
	runtime.GC()

	for {
		printMemStats()
		time.Sleep(time.Second)
	}
}

func BenchmarkLocalCache_SetPointer(b *testing.B) {
	lc := localcache.New()
	a := &Person{"Jack", 18, "America"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lc.Set(strconv.Itoa(i), a, 300)
	}
}

func BenchmarkLocalCache_SetStruct(b *testing.B) {
	lc := localcache.New()
	a := Person{"Jack", 18, "America"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lc.Set(strconv.Itoa(i), a, 300)
	}
}

func BenchmarkLocalCache_GetPointer(b *testing.B) {
	lc := localcache.New()
	a := &Person{"Jack", 18, "America"}
	lc.Set("1", a, 300)
	var e *Person
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		it, _, _ := lc.Get("1")
		e = it.(*Person)
	}
	log.Println(e)
}

func BenchmarkLocalCache_GetStruct(b *testing.B) {
	lc := localcache.New()
	a := Person{"Jack", 18, "America"}
	lc.Set("1", a, 300)
	var e Person
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		it, _, _ := lc.Get("1")
		e = it.(Person)
	}
	log.Println(e)
}
