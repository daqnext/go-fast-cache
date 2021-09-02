# meson.network-lts-local-cache

```
lc := localcache.New() // or use NewWithInterval(intervalSecond int) custom the schedule job interval
lc.SetCountLimit(10000) //if not set default is 100000
//set
lc.Set("foo", "bar", 300)
lc.Set("a", 1, 300)
lc.Set("b", Person{"Jack", 18}, 300)
lc.Set("b*", &Person{"Jack", 18}, 300)
lc.Set("c", true, 100)
//get
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