package test

import (
	locallog "github.com/daqnext/LocalLog/log"
	localcache "github.com/daqnext/go-fast-cache"
	"github.com/daqnext/go-fast-cache/ttltype"
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

var log *locallog.LocalLog

func printMemStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log.Printf("Alloc = %v KB, TotalAlloc = %v KB, Sys = %v KB,Lookups = %v NumGC = %v\n", m.Alloc/1024, m.TotalAlloc/1024, m.Sys/1024, m.Lookups, m.NumGC)
}

func init() {
	var err error
	log, err = locallog.New("logs", 2, 20, 30)
	if err != nil {
		panic(err.Error())
	}
}

func Test_main(t *testing.T) {
	lc := localcache.New(log) // or use NewWithInterval(intervalSecond int) custom the schedule job interval
	lc.SetCountLimit(10000)   //if not set default is 100000

	//set
	lc.Set("foo", "bar", 300)
	lc.Set("a", 1, 300)
	lc.Set("b", Person{"Jack", 18, "London"}, 300)
	lc.Set("b*", &Person{"Jack", 18, "London"}, 300)
	lc.Set("c", true, 100)

	//get
	value, ttlLeft, exist := lc.Get("foo")
	if exist {
		valueStr, ok := value.(string) //value type is interface{}, please convert to the right type before use
		if ok {
			log.Println("key:foo, value:", valueStr)
		}
		log.Println("key:foo, ttl:", ttlLeft)
	}

	//get
	log.Println("---get---")
	log.Println(lc.Get("foo"))
	log.Println(lc.Get("a"))
	log.Println(lc.Get("b"))
	log.Println(lc.Get("b*"))
	log.Println(lc.Get("c"))

	////overwrite
	log.Println("---set overwrite---")
	log.Println(lc.Get("c"))
	lc.Set("c", false, 60)
	log.Println(lc.Get("c"))
}

func Test_Set(t *testing.T) {
	lc := localcache.NewWithInterval(1, log)
	for i := 0; i < 100000; i++ {
		lc.Set(strconv.Itoa(i), "aaaaaaaaaaaaaaaaaaaaaaa", 60)
	}

	for i := 0; i < 100000; i += 1000 {
		log.Println(lc.Get(strconv.Itoa(i)))
	}

}

func Test_Get(t *testing.T) {
	lc := localcache.New(log)
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

func Test_Delete(t *testing.T) {
	lc := localcache.New(log)
	a := &Person{"Jack", 18, "London"}
	lc.Set("a", a, 5)
	lc.Set("b", a, 5)

	v, ttl, exist := lc.Get("a")
	log.Println("get a")
	log.Println(v, ttl, exist)

	log.Println("delete a")
	lc.Delete("a")

	v, ttl, exist = lc.Get("a")
	log.Println("get a")
	log.Println(v, ttl, exist)

	v, ttl, exist = lc.Get("b")
	log.Println("get b")
	log.Println(v, ttl, exist)

	log.Println("origin a")
	log.Println(*a)

	time.Sleep(500 * time.Second)
}

func Test_Expire(t *testing.T) {
	lc := localcache.New(log)
	lc.Set("1", "111", 5)
	lc.Set("2", "111", 18)
	lc.Set("3", "111", 23)
	lc.Set("4", "111", -100)
	lc.Set("5", "111", 3000000)
	lc.Set("6", "111", 35)

	count := 0
	for {
		v, ttl, ok := lc.Get("1")
		log.Printf("1==>%v %v %v", v, ttl, ok)
		v, ttl, ok = lc.Get("2")
		log.Printf("2==>%v %v %v", v, ttl, ok)
		v, ttl, ok = lc.Get("3")
		log.Printf("3==>%v %v %v", v, ttl, ok)
		v, ttl, ok = lc.Get("4")
		log.Printf("4==>%v %v %v", v, ttl, ok)
		v, ttl, ok = lc.Get("5")
		log.Printf("5==>%v %v %v", v, ttl, ok)
		v, ttl, ok = lc.Get("6")
		log.Printf("6==>%v %v %v", v, ttl, ok)
		log.Println("total key", lc.GetLen())
		log.Println("-----------")
		count++
		if count > 40 {
			return
		}
		time.Sleep(time.Second)
	}
}

func Test_SetAndRemove(t *testing.T) {
	a := Person{"Jack", 18, "America"}
	lc := localcache.NewWithInterval(1, log)

	log.Println("start")
	printMemStats()

	for i := 0; i < 20; i++ {
		//set
		for j := 0; j < 10000; j++ {
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
	lc := localcache.New(log)

	log.Println("start")
	printMemStats()

	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:10000", nil))
	}()

	for i := 0; i < 30; i++ {
		log.Println("----------")
		log.Println("round", i)
		log.Println("mem start set")
		printMemStats()

		for i := 0; i < 1000000; i++ {
			lc.Set(strconv.Itoa(i), a, int64(rand.Intn(10)+1))
		}
		log.Println("mem after set")
		printMemStats()
		time.Sleep(time.Second)
	}

	log.Println("~~~~~~")
	log.Println("finish set")
	printMemStats()

	log.Println("do GC")

	runtime.GC()
	log.Println("after GC")
	printMemStats()

	count := 0
	for {
		time.Sleep(1 * time.Second)
		log.Println("---job finished---")
		printMemStats()
		count++
		if count > 45 {
			//return
		}
	}
	//time.Sleep(1*time.Hour)
}

func Test_RandSet(t *testing.T) {
	lc := localcache.New(log)
	a := Person{"Jack", 18, "America"}

	lc.Set("a", a, 15)
	lc.Set("b", a, 19)
	lc.Set("c", a, 60)
	lc.Set("d", a, 63)
	lc.Set("e", a, 65)

	log.Println("before big amount set")
	v, ttl, ok := lc.Get("a")
	log.Printf("a==>%v %v %v", v, ttl, ok)
	v, ttl, ok = lc.Get("b")
	log.Printf("b==>%v %v %v", v, ttl, ok)
	v, ttl, ok = lc.Get("c")
	log.Printf("c==>%v %v %v", v, ttl, ok)
	v, ttl, ok = lc.Get("d")
	log.Printf("d==>%v %v %v", v, ttl, ok)
	v, ttl, ok = lc.Get("e")
	log.Printf("e==>%v %v %v", v, ttl, ok)

	log.Println("start amount set")
	for i := 0; i < 200; i++ {
		for j := 0; j < 10000; j++ {
			num := rand.Intn(9999999999999)
			key := strconv.Itoa(num)
			lc.Set(key, a, int64(rand.Intn(30)+20))
		}
	}

	for i := 0; i < 70; i++ {
		time.Sleep(time.Second)
		log.Println("--------------")
		v, ttl, ok = lc.Get("a")
		log.Printf("a==>%v %v %v", v, ttl, ok)
		v, ttl, ok = lc.Get("b")
		log.Printf("b==>%v %v %v", v, ttl, ok)
		v, ttl, ok = lc.Get("c")
		log.Printf("c==>%v %v %v", v, ttl, ok)
		v, ttl, ok = lc.Get("d")
		log.Printf("d==>%v %v %v", v, ttl, ok)
		v, ttl, ok = lc.Get("e")
		log.Printf("e==>%v %v %v", v, ttl, ok)
		log.Println("total key", lc.GetLen())
	}
}

func Test_KeepTTL(t *testing.T) {
	lc := localcache.New(log)
	a := Person{"Ma Yun", 58, "China"}
	b := Person{"Jack Ma", 18, "America"}

	lc.Set("a", a, 30)
	lc.Set("b", a, 40)
	lc.Set("c", a, 50)

	//log
	v, ttl, ok := lc.Get("a")
	log.Printf("a==>%v %v %v", v, ttl, ok)
	v, ttl, ok = lc.Get("b")
	log.Printf("b==>%v %v %v", v, ttl, ok)
	v, ttl, ok = lc.Get("c")
	log.Printf("c==>%v %v %v", v, ttl, ok)

	time.Sleep(5 * time.Second)

	lc.Set("a", b, 300)
	lc.Set("b", b, ttltype.Keep)

	//log
	for i := 0; i < 10; i++ {
		log.Println("-----------")
		v, ttl, ok = lc.Get("a")
		log.Printf("a==>%v %v %v", v, ttl, ok)
		v, ttl, ok = lc.Get("b")
		log.Printf("b==>%v %v %v", v, ttl, ok)
		v, ttl, ok = lc.Get("c")
		log.Printf("c==>%v %v %v", v, ttl, ok)
		time.Sleep(time.Second)
	}

}

func Test_SetTTL(t *testing.T) {
	lc := localcache.New(log)
	a := Person{"Ma Yun", 58, "China"}

	ttls := []int64{1, 20000, ttltype.Keep, -100, 200, 45, 346547457457457, -20000, 434, 9}
	for i := 0; i < 10; i++ {
		key := strconv.Itoa(i)
		lc.Set(key, a, ttls[i])
	}

	for i := 0; i < 10; i++ {
		log.Println("-----------")
		for j := 0; j < 10; j++ {
			key := strconv.Itoa(j)
			v, ttl, ok := lc.Get(key)
			log.Printf("%s==>%v %v %v", key, v, ttl, ok)
		}
		log.Println("total key", lc.GetLen())
		time.Sleep(time.Second)
	}
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
	lc := localcache.New(log)
	a := &Person{"Jack", 18, "America"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lc.Set(strconv.Itoa(i), a, 300)
	}
}

func BenchmarkLocalCache_SetStruct(b *testing.B) {
	lc := localcache.New(log)
	a := Person{"Jack", 18, "America"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lc.Set(strconv.Itoa(i), a, 300)
	}
}

func BenchmarkLocalCache_GetPointer(b *testing.B) {
	lc := localcache.New(log)
	a := &Person{"Jack", 18, "America"}
	lc.Set("1", a, 300)
	var e *Person
	log.Println(e)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		it, _, _ := lc.Get("1")
		e = it.(*Person)
	}

}

func BenchmarkLocalCache_GetStruct(b *testing.B) {
	lc := localcache.New(log)
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
