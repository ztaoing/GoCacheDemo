/**
* @Author:zhoutao
* @Date:2021/1/3 上午9:58
* @Desc:
 */

package lru

import "container/list"

type Cache struct {
	maxBytes  int64      // the max allowed memory can be used
	nBytes    int64      //  current used memory
	ll        *list.List // doubly-linked list
	cache     map[string]*list.Element
	OnEvicted func(key string, value Value) // call OnEvicted() when a key evicted
}

type entry struct {
	key   string
	value Value
}

//为了通用性，允许值是实现了Value接口的任意类型
type Value interface {
	Len() int // return the how many memory needed
}

func NewCache(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

//2 steps: first:get the node from map by key,second: move the node to the tail of the doubly-linked list
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// delete the node from list then delete it from map and update the memory size
func (c *Cache) RemoveOldest() {
	//the last element of list
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) AddCache(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nBytes += int64(value.Len()) - int64(kv.value.Len())
		// update new value
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key: key, value: value})
		c.cache[key] = ele
		c.nBytes += int64(len(key)) + int64(value.Len())
	}
	// remove the least recent used node
	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}

// the length of the doubly-linked list
func (c *Cache) Len() int {
	return c.ll.Len()
}
