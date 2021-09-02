package meson_network_lts_local_cache

import (
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"

	_ "net/http/pprof"
)

func printMemStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log.Printf("Alloc = %v KB, TotalAlloc = %v KB, Sys = %v KB,Lookups = %v NumGC = %v\n", m.Alloc/1024, m.TotalAlloc/1024, m.Sys/1024, m.Lookups, m.NumGC)
}

func Test_Set(t *testing.T) {
	lc := New(0)
	for i := 0; i < 100000; i++ {
		lc.Set(strconv.Itoa(i), "aaaaaaaaaaaaaaaaaaaaaaa", 60)
	}

	for i := 0; i < 100000; i += 1000 {
		log.Println(lc.Get(strconv.Itoa(i)))
	}

}

func Test_Get(t *testing.T) {
	lc := New(0)
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
	lc := New(0)
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

func Test_BigAmountKey(t *testing.T) {
	type Person struct {
		Name     string
		Age      int
		Location string
	}
	a := Person{"Jack", 18, "America"}
	lc := New(1)
	lc.SetCountLimit(1000000)

	printMemStats()

	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:10000", nil))
	}()

	go func() {
		for {
			time.Sleep(5 * time.Second)
			//runtime.ReadMemStats(&m)
			//log.Printf("memstate %d,%d,%d,%d\n", m.HeapSys, m.HeapAlloc,
			//	m.HeapIdle, m.HeapReleased)

			printMemStats()
		}
	}()

	//wg := sync.WaitGroup{}
	//wg.Add(2)

	printMemStats()
	//go func() {
	//wg.Add(1000000)
	for i := 0; i < 1000000; i++ {
		//j := i % 1000000
		//go func() {
		//time.Sleep(time.Duration(rand.Intn(20000)) * time.Millisecond)
		lc.Set(strconv.Itoa(i), a, int64(rand.Intn(10)+1))
		//wg.Done()
		//}()
	}
	//wg.Done()
	//}()

	go func() {
		for i := 0; i < 100; i++ {
			//time.Sleep(20*time.Second)
			log.Println("Start round", i+1)
			for i := 0; i < 1000000; i++ {
				//j := i % 1000000
				//go func() {
				//time.Sleep(time.Duration(rand.Intn(20000)) * time.Millisecond)
				lc.Set(strconv.Itoa(i), a, int64(rand.Intn(10)+1))
				//wg.Done()
				//}()
			}
		}

	}()

	printMemStats()
	//time.Sleep(time.Millisecond * 5000)
	//go func() {
	//	for i := 0; i < 1000000; i++ {
	//		j := i % 1000000
	//
	//		lc.Get(strconv.Itoa(j))
	//
	//	}
	//wg.Done()
	//}()

	//wg.Wait()

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
	var myMap sync.Map
	for i := 0; i < 1000000; i++ {
		myMap.Store(strconv.Itoa(i), a)
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

func Test_RemoveByRank(t *testing.T) {
	lc := New(0)
	for i := 0; i < 100; i++ {
		lc.Set(strconv.Itoa(i), strconv.Itoa(i), 60+int64(i))
	}

	lc.s.RemoveByRank(0, 10)
	lc.s.RemoveByRank(0, 10)
	lc.s.RemoveByRank(0, 10)
	lc.s.RemoveByRank(0, 10)
	lc.s.RemoveByRank(0, 10)

	log.Println(lc.s.Len())

	e := lc.s.RangeByScore(0, MaxTTL, 0, -1, false)
	for _, v := range e {
		log.Println(v.Member, v.Value)
	}

}

func BenchmarkLocalCache_SetPointer(b *testing.B) {
	lc := New(0)
	type Person struct {
		Name     string
		Age      int
		Location string
	}
	a := &Person{"Jack", 18, "America"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lc.Set(strconv.Itoa(i), a, 300)
	}
}

func BenchmarkLocalCache_SetStruct(b *testing.B) {
	lc := New(0)
	type Person struct {
		Name     string
		Age      int
		Location string
	}
	a := Person{"Jack", 18, "America"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lc.Set(strconv.Itoa(i), a, 300)
	}
}

func BenchmarkLocalCache_GetPointer(b *testing.B) {
	lc := New(0)
	type Person struct {
		Name     string
		Age      int
		Location string
	}
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
	lc := New(0)
	type Person struct {
		Name     string
		Age      int
		Location string
	}
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

func BenchmarkLocalCache_Set(b *testing.B) {
	lc := New(0)
	type Person struct {
		Name     string
		Age      int
		Location string
	}
	a := Person{"Jack", 18, "America"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lc.Set(strconv.Itoa(i), a, 300)
	}
}

func BenchmarkLocalCache_Get(b *testing.B) {
	lc := New(0)
	for i := 0; i < 20000; i++ {
		lc.Set(strconv.Itoa(i), "abcdefghijkl", 300)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lc.Get("a")
	}
}
