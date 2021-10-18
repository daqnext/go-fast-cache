package sortedset

import (
	"sync"
	"sync/atomic"
)

// SortedSet is a set which keys sorted by bound score
type SortedSet struct {
	dict     sync.Map
	skiplist *skiplist

	elementCount int64
	lock         sync.Mutex
	slChannel    chan func()
}

// Make makes a new SortedSet
func Make() *SortedSet {

	s := &SortedSet{
		skiplist:     makeSkiplist(),
		elementCount: 0,
		slChannel:    make(chan func(), 20000),
	}
	s.handleChannelJob()
	return s
}

func (sortedSet *SortedSet) handleChannelJob() {
	go func() {
		for {
			f := <-sortedSet.slChannel
			sortedSet.lock.Lock()
			f()
			sortedSet.lock.Unlock()
		}
	}()
}

// Add puts member into set,  and returns whether has inserted new node
func (sortedSet *SortedSet) Add(member string, score int64, value interface{}) {
	element, exist := sortedSet.dict.Load(member)
	sortedSet.dict.Store(member, &Element{
		//Member: member,
		Score: score,
		Value: value,
	})
	if !exist {
		atomic.AddInt64(&sortedSet.elementCount, 1)
		//log.Println("count after add", sortedSet.elementCount)
		fc := func() {
			sortedSet.skiplist.insert(member, score)
		}
		sortedSet.slChannel <- fc
	} else {
		elementScore := element.(*Element).Score
		if score != elementScore {
			fc := func() {
				sortedSet.skiplist.remove(member, elementScore)
				sortedSet.skiplist.insert(member, score)
			}
			sortedSet.slChannel <- fc
		}
		//log.Println("count after cover",sortedSet.elementCount)
	}
}

func (sortedSet *SortedSet) Remove(member string) {
	element, exist := sortedSet.dict.Load(member)
	if !exist {
		return
	}
	atomic.AddInt64(&sortedSet.elementCount, -1)
	elementScore := element.(*Element).Score
	fc := func() {
		sortedSet.skiplist.remove(member, elementScore)
	}
	sortedSet.slChannel <- fc
	sortedSet.dict.Delete(member)
}

// Len returns number of members in set
func (sortedSet *SortedSet) Len() int64 {
	return sortedSet.elementCount
}

func (sortedSet *SortedSet) SLen() int64 {
	return sortedSet.skiplist.length
}

func (sortedSet *SortedSet) MapLen() int64 {
	count := int64(0)
	sortedSet.dict.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// Get returns the given member
func (sortedSet *SortedSet) Get(member string) (element *Element, ok bool) {
	elementI, exist := sortedSet.dict.Load(member)
	if !exist {
		return nil, false
	}
	return elementI.(*Element), true
}

// ForEachByScore visits members which score within the given border
func (sortedSet *SortedSet) ForEachByScore(min int64, max int64, offset int64, limit int64, desc bool, consumer func(node *node) bool) {
	// find start node
	var node *node
	if desc {
		node = sortedSet.skiplist.getLastInScoreRange(min, max)
	} else {
		node = sortedSet.skiplist.getFirstInScoreRange(min, max)
	}

	for node != nil && offset > 0 {
		if desc {
			node = node.backward
		} else {
			node = node.level[0].forward
		}
		offset--
	}

	// A negative limit returns all elements from the offset
	for i := 0; (i < int(limit) || limit < 0) && node != nil; i++ {
		if !consumer(node) {
			break
		}
		if desc {
			node = node.backward
		} else {
			node = node.level[0].forward
		}
		if node == nil {
			break
		}
		gtMin := min <= (node.Score) // greater than min
		ltMax := max >= (node.Score)
		if !gtMin || !ltMax {
			break // break through score border
		}
	}
}

// RangeByScore returns members which score within the given border
// param limit: <0 means no limit
func (sortedSet *SortedSet) RangeByScore(min int64, max int64, offset int64, limit int64, desc bool) []*Element {
	if limit == 0 || offset < 0 {
		return make([]*Element, 0)
	}
	slice := make([]*Element, 0)
	sortedSet.ForEachByScore(min, max, offset, limit, desc, func(node *node) bool {
		element, ok := sortedSet.dict.Load(node.Member)
		if ok {
			slice = append(slice, element.(*Element))
		}
		return true
	})
	return slice
}

// RemoveByScore removes members which timestamp < now time
func (sortedSet *SortedSet) RemoveByScore(max int64) int64 {

	sortedSet.lock.Lock()
	removed := sortedSet.skiplist.RemoveRangeByScore(0, max)
	sortedSet.lock.Unlock()
	for _, element := range removed {
		sortedSet.dict.Delete(element)
	}
	sortedSet.elementCount = sortedSet.SLen()
	//log.Println("count after RemoveByScore", "map len", sortedSet.elementCount, "list len", sortedSet.SLen())
	return int64(len(removed))
}

// RemoveByRank removes member ranking within [start, stop)
// sort by ascending order and rank starts from 0
func (sortedSet *SortedSet) RemoveByRank(start int64, stop int64) int64 {

	sortedSet.lock.Lock()
	removed := sortedSet.skiplist.RemoveRangeByRank(start+1, stop+1)
	sortedSet.lock.Unlock()
	for _, element := range removed {
		sortedSet.dict.Delete(element)
	}
	sortedSet.elementCount = sortedSet.SLen()
	//log.Println("count after RemoveByRank", "map len", sortedSet.elementCount, "list len", sortedSet.SLen())
	return int64(len(removed))
}
