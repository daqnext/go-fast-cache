package meson_network_lts_local_cache

import (
	"log"
	"strconv"
	"testing"
	"time"
)

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
	lc.SetCountLimit(10000000)

	for i := 0; i < 10000000; i++ {
		j := i % 10000

		lc.Set(strconv.Itoa(j), a, 1+int64(j))

	}

	e, ttl, exist := lc.Get("1")
	if exist {
		log.Println(e.(Person).Name)
	}
	log.Println(ttl)
	log.Println(exist)
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
