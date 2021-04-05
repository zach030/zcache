package zcache

import (
	"errors"
	"log"
	"sync"
)

// 用户回调函数，缓存未命中时，获取源数据
// A Getter loads data for a key.
type Getter interface {
	Get(key string) ([]byte, error)
}

// A GetterFunc implements Getter with a function.
type GetterFunc func(key string) ([]byte, error)

// Get implements Getter interface function
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

type Group struct {
	name   string
	getter Getter
	cache  cache
}

func NewGroup(name string, maxSize int64, getter Getter) *Group {
	if getter == nil {
		panic("nil getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:   name,
		getter: getter,
		cache:  cache{cacheBytes: maxSize},
	}
	groups[name] = g
	return g
}

func GetGroup(name string)*Group{
	mu.Lock()
	defer mu.Unlock()
	g := groups[name]
	return g
}

func (g *Group)Get(key string)(ByteView,error){
	if key==""{
		return ByteView{},errors.New("empty key")
	}
	if v,ok:= g.cache.get(key);ok{
		log.Println("cache hit")
		return v,nil
	}
	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err

	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.cache.add(key, value)
}