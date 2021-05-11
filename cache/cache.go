package cache

import (
	"fmt"
	"sync"
	"zcache/zlog"
)

func DefaultInsertHook(key string, value Value) {
	zlog.Info(fmt.Sprintf("add key:%s, value :%v", key, value))
}

func DefaultDeleteHook(key string, value Value) {
	zlog.Info(fmt.Sprintf("remove key:%s, value :%v", key, value))
}

type BaseCache struct {
	lock      sync.Mutex
	cache     *Cache
	cacheSize int64
}

func (b *BaseCache) get(key string) (value View, ok bool) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if b.cache == nil {
		zlog.Error("cache is nil")
		return
	}
	if value, ok := b.cache.Get(key); ok {
		return value.(View), ok
	}
	return
}

func (b *BaseCache) set(key string, value View) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if b.cache == nil {
		b.cache = NewCache(b.cacheSize, DefaultInsertHook, DefaultDeleteHook)
	}
	b.cache.Set(key, value)
}
