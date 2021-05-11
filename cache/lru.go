package cache

import "container/list"

type Cache struct {
	maxSize    int64
	currSize   int64
	ll         *list.List
	cache      map[string]*list.Element
	insertHook func(key string, value Value)
	deleteHook func(key string, value Value)
}

type Value interface {
	Len() int
}

type entry struct {
	key   string
	value Value
}

func NewCache(max int64, inHook func(key string, value Value), outHook func(key string, value Value)) *Cache {
	return &Cache{
		maxSize:    max,
		currSize:   0,
		ll:         list.New(),
		cache:      make(map[string]*list.Element, 0),
		insertHook: inHook,
		deleteHook: outHook,
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}

// lru get
func (c *Cache) Get(key string) (Value, bool) {
	if v, ok := c.cache[key]; ok {
		c.ll.MoveToFront(v)
		ele := v.Value.(*entry)
		return ele.value, ok
	}
	return nil, false
}

func (c *Cache) Set(key string, value Value) {
	// if exist,update it
	if v, ok := c.cache[key]; ok {
		c.ll.MoveToFront(v)
		// update curr size
		oldEle := v.Value.(*entry)
		c.currSize = int64(oldEle.value.Len()) - int64(value.Len())
		oldEle.value = value
		if c.insertHook != nil {
			c.insertHook(key, value)
		}
	} else {
		ele := c.ll.PushFront(&entry{
			key:   key,
			value: value,
		})
		c.currSize += int64(value.Len()) + int64(len(key))
		c.cache[key] = ele
	}
	for c.currSize >= c.maxSize && c.maxSize != 0 {
		c.RemoveOldest()
	}
}

// remove the last one and update size
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.currSize -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.deleteHook != nil {
			c.deleteHook(kv.key, kv.value)
		}
	}
}
