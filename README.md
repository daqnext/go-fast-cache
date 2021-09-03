# go-fast-cache

local-cache is a high-speed, thread-safe, key-value caching system. All data keep in memory.

## usage
```go
//new instance
//new a localcache instance with default config
//DefaultDeleteExpireIntervalSecond(Schedule job for delete expired key interval) is 5 seconds
//DefaultCountLimit(Max key-value pair count) is 100,0000
lc := localcache.New() 

//set
//Set(key string, value interface{}, ttlSecond int64)
//ttlSecond should >0, MaxTTL is 7200sec(2 hour)
lc.Set("foo", "bar", 300)
lc.Set("a", 1, 300)
lc.Set("b", Person{"Jack", 18}, 300)
lc.Set("b*", &Person{"Jack", 18}, 300)
lc.Set("c", true, 100)

//get
//Get(key string) (value interface{}, ttl int64, exist bool)
log.Println("---get---")
log.Println(lc.Get("foo"))
log.Println(lc.Get("a"))
log.Println(lc.Get("b"))
log.Println(lc.Get("b*"))
log.Println(lc.Get("c"))

//set cover
log.Println("---cover set---")
log.Println(lc.Get("c"))
lc.Set("c", false, 60)
log.Println(lc.Get("c"))
```

### custom DeleteExpireIntervalSecond and key-value pair CountLimit
```go
//new instance
lc := localcache.NewWithInterval(20) //custom schedule job interval(second) for delete expired key
lc.SetCountLimit(10000) //custom the max key-value pair count
```

## Benchmark
### set
```
cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
BenchmarkLocalCache_SetPointer
BenchmarkLocalCache_SetPointer-8   	 1000000	      1618 ns/op	     379 B/op	      10 allocs/op
PASS
```

### get
```
cpu: Intel(R) Core(TM) i7-7700HQ CPU @ 2.80GHz
BenchmarkLocalCache_GetPointer
BenchmarkLocalCache_GetPointer-8   	 9931429	       129.7 ns/op	       0 B/op	       0 allocs/op
PASS
```