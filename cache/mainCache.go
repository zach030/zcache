package cache

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"zcache"
	"zcache/singleflight"
	"zcache/zlog"
)

// if not cached, custom get from source data
type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

func init() {
	cachePool = make(map[string]*GroupCache, 0)
}

// 多了一个 getter方法，用于内存缓存未命中时访问DB使用
// picker 分布式的节点选择方法
type GroupCache struct {
	name   string
	getter Getter
	cache  *BaseCache
	picker zcache.PeerPicker
	loader *singleflight.Group
}

var (
	lock      sync.RWMutex
	cachePool map[string]*GroupCache
)

func NewGroupCache(name string, size int64, getter Getter) *GroupCache {
	lock.Lock()
	defer lock.Unlock()
	mc := &GroupCache{
		name:   name,
		getter: getter,
		cache:  &BaseCache{cacheSize: size},
		loader: &singleflight.Group{},
	}
	cachePool[name] = mc
	fmt.Printf("cache pool is :%+v\n", cachePool)
	return mc
}

func GetCache(name string) *GroupCache {
	lock.Lock()
	defer lock.Unlock()
	return cachePool[name]
}

func (g *GroupCache) GetName() string {
	return g.name
}

func (g *GroupCache) RegisterPicker(picker zcache.PeerPicker) {
	if g.picker != nil {
		panic("call register picker more than once")
	}
	g.picker = picker
}

func (g *GroupCache) Get(key string) (View, error) {
	// 先从本节点缓存读
	if key == "" {
		return View{}, errors.New("empty key")
	}
	if v, ok := g.cache.get(key); ok {
		log.Println("cache hit")
		return v, nil
	}
	// 从其他节点读
	return g.load(key)
}

func (g *GroupCache) load(key string) (v View, err error) {
	//todo 增加singlefly
	viewi, err := g.loader.DoSingle(key, func() (interface{}, error) {
		if g.picker != nil {
			// 挑选一个节点
			if getter, ok := g.picker.Pick(key); ok {
				// 从节点获取数据
				if data, err := g.GetFromPeer(getter, key); err == nil {
					return data, err
				}
				zlog.Info("failed to load from peer")
			}
			zlog.Info("failed to select peer")
		}
		// 从数据库读
		return g.getFromDB(key)
	})
	if err != nil {
		return viewi.(View), err
	}
	return
}

func (g *GroupCache) getFromDB(key string) (View, error) {
	// 通过自定义getter方法，读DB
	v, err := g.getter.Get(key)
	if err != nil {
		zlog.Error("get from source failed,err:" + err.Error())
		return View{}, err
	}
	value := View{data: cloneBytes(v)}
	// 加入缓存
	g.addCache(key, value)
	return value, err
}

func (g *GroupCache) addCache(key string, v View) {
	g.cache.set(key, v)
}

func (g *GroupCache) GetFromPeer(peerGetter zcache.PeerGetter, key string) (View, error) {
	bytes, err := peerGetter.Get(g.name, key)
	if err != nil {
		return View{}, err
	}
	return View{data: bytes}, nil
}
