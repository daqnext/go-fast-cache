package sortedset

import (
	"log"
	"sync"
	"sync/atomic"
)

// SortedSet is a set which keys sorted by bound score
type SortedSet struct {
	dict         sync.Map
	skiplist     *skiplist
	elementCount int64
	lock         sync.Mutex
	slChannel    chan *channelFunc
}

// Make makes a new SortedSet
func Make() *SortedSet {

	s := &SortedSet{
		skiplist:     makeSkiplist(),
		elementCount: 0,
		slChannel:    make(chan *channelFunc, 20000),
	}
	s.handleChannelJob()
	return s
}

type channelFunc struct {
	f      func(member string, score int64)
	member string
	score  int64

	f2 func()
}

func (sortedSet *SortedSet) handleChannelJob() {
	go func() {
		for {
			f := <-sortedSet.slChannel
			sortedSet.lock.Lock()
			f.f2()
			sortedSet.lock.Unlock()
		}
	}()
}

// Add puts member into set,  and returns whether has inserted new node
func (sortedSet *SortedSet) Add(member string, score int64, value interface{}) {
	element, ok := sortedSet.dict.Load(member)
	sortedSet.dict.Store(member, &Element{
		Member: member,
		Score:  score,
		Value:  value,
	})
	if !ok {
		atomic.AddInt64(&sortedSet.elementCount, 1)
		//sortedSet.elementCount++
		log.Println("count after add", sortedSet.elementCount)

		//todo send to channel
		fc := func() {
			sortedSet.skiplist.insert(member, score)
		}
		sortedSet.slChannel <- &channelFunc{sortedSet.skiplist.insert, member, score, fc}
		//sortedSet.skiplist.insert(member, score)
	} else {
		elementScore := element.(*Element).Score
		if score != elementScore {

			//todo send to channel
			fc := func() {
				sortedSet.skiplist.remove(member, elementScore)
				sortedSet.skiplist.insert(member, score)
			}
			sortedSet.slChannel <- &channelFunc{sortedSet.skiplist.remove, member, elementScore, fc}
			//sortedSet.slChannel<-&channelFunc{sortedSet.skiplist.insert,member,score}
			//sortedSet.skiplist.remove(member, elementScore)
			//sortedSet.skiplist.insert(member, score)
		}
		//log.Println("count after cover",sortedSet.elementCount)
	}
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
	elementI, ok := sortedSet.dict.Load(member)
	if !ok {
		return nil, false
	}
	return elementI.(*Element), true
}

// Remove removes the given member from set
//func (sortedSet *SortedSet) Remove(member string) bool {
//	v, ok := sortedSet.dict.Load(member)
//	if ok {
//		sortedSet.skiplist.remove(member, v.(*Element).Score)
//
//		sortedSet.dict.Delete(member)
//		atomic.AddInt64(&sortedSet.elementCount,-1)
//		return true
//	}
//	return false
//}

// GetRank returns the rank of the given member, sort by ascending order, rank starts from 0
//func (sortedSet *SortedSet) GetRank(member string, desc bool) (rank int64) {
//	element, ok := sortedSet.dict.Load(member)
//	if !ok {
//		return -1
//	}
//	r := sortedSet.skiplist.getRank(member, element.(*Element).Score)
//	if desc {
//		r = sortedSet.skiplist.length - r
//	} else {
//		r--
//	}
//	return r
//}

// ForEach visits each member which rank within [start, stop), sort by ascending order, rank starts from 0
//func (sortedSet *SortedSet) ForEach(start int64, stop int64, desc bool, consumer func(node *node) bool) {
//	size := sortedSet.Len()
//	if start < 0 || start >= size {
//		panic("illegal start " + strconv.FormatInt(start, 10))
//	}
//	if stop < start || stop > size {
//		panic("illegal end " + strconv.FormatInt(stop, 10))
//	}
//
//	// find start node
//	var node *node
//	if desc {
//		node = sortedSet.skiplist.tail
//		if start > 0 {
//			node = sortedSet.skiplist.getByRank(size - start)
//		}
//	} else {
//		node = sortedSet.skiplist.header.level[0].forward
//		if start > 0 {
//			node = sortedSet.skiplist.getByRank(start + 1)
//		}
//	}
//
//	sliceSize := int(stop - start)
//	for i := 0; i < sliceSize; i++ {
//		if !consumer(node) {
//			break
//		}
//		if desc {
//			node = node.backward
//		} else {
//			node = node.level[0].forward
//		}
//	}
//}

// Range returns members which rank within [start, stop), sort by ascending order, rank starts from 0
//func (sortedSet *SortedSet) Range(start int64, stop int64, desc bool) []*Element {
//	sliceSize := int(stop - start)
//	slice := make([]*Element, sliceSize)
//	i := 0
//	sortedSet.ForEach(start, stop, desc, func(node *node) bool {
//		element, ok := sortedSet.dict.Load(node.Member)
//		if ok {
//			slice[i] = element.(*Element)
//			i++
//		}
//		return true
//	})
//	return slice
//}

// Count returns the number of  members which score within the given border
//func (sortedSet *SortedSet) Count(min int64, max int64) int64 {
//	var i int64 = 0
//	// ascending order
//	sortedSet.ForEach(0, sortedSet.Len(), false, func(node *node) bool {
//		gtMin := min <= (node.Score) // greater than min
//		if !gtMin {
//			// has not into range, continue foreach
//			return true
//		}
//		ltMax := max >= (node.Score) // less than max
//		if !ltMax {
//			// break through score border, break foreach
//			return false
//		}
//		// gtMin && ltMax
//		i++
//		return true
//	})
//	return i
//}

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

// RemoveByScore removes members which score within the given border
func (sortedSet *SortedSet) RemoveByScore(min int64, max int64) int64 {

	//todo send to channel or add lock
	sortedSet.lock.Lock()
	removed := sortedSet.skiplist.RemoveRangeByScore(min, max)
	sortedSet.lock.Unlock()
	for _, element := range removed {
		sortedSet.dict.Delete(element)
	}
	//atomic.AddInt64(&sortedSet.elementCount, int64(-len(removed)))
	//sortedSet.elementCount-=int64(len(removed))
	sortedSet.elementCount = sortedSet.SLen()
	log.Println("count after RemoveByScore", "map len", sortedSet.elementCount, "list len", sortedSet.SLen())

	return int64(len(removed))
}

// RemoveByRank removes member ranking within [start, stop)
// sort by ascending order and rank starts from 0
func (sortedSet *SortedSet) RemoveByRank(start int64, stop int64) int64 {

	//todo send to channel or add lock
	sortedSet.lock.Lock()
	removed := sortedSet.skiplist.RemoveRangeByRank(start+1, stop+1)
	sortedSet.lock.Unlock()
	for _, element := range removed {
		sortedSet.dict.Delete(element)
	}
	//atomic.AddInt64(&sortedSet.elementCount, int64(-len(removed)))
	//sortedSet.elementCount-=int64(len(removed))
	sortedSet.elementCount = sortedSet.SLen()
	log.Println("count after RemoveByRank", "map len", sortedSet.elementCount, "list len", sortedSet.SLen())

	return int64(len(removed))

}
