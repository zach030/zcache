package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"sync"
	"zcache"
	"zcache/cache"
	"zcache/consistent_hash"
	"zcache/node"
	"zcache/zlog"
)

const (
	defaultSpace    = "zach"
	defaultReplicas = 50
)

type CacheContext struct {
	group  string
	method string
	key    string
	value  []byte
}

// cache cacheServer 使用一致性哈希实现挑选节点方法
type CacheServer struct {
	context     *CacheContext
	cacheServer *gin.Engine
	port        int64
	url         string
	namespace   string
	peers       *consistent_hash.Map // 一致性哈希算法，选择节点
	lock        sync.Mutex
	nodeMap     map[string]*node.ClientNode // 映射节点与节点客户端
}

func NewCacheServer(port int64, namespace string) *CacheServer {
	s := &CacheServer{
		context:     &CacheContext{},
		cacheServer: gin.Default(),
		port:        port,
		namespace:   namespace,
		nodeMap:     make(map[string]*node.ClientNode),
		url:         fmt.Sprintf("http://localhost:%v", port),
	}
	if namespace == "" {
		s.namespace = defaultSpace
	}
	return s
}

func (s *CacheServer) Start(group *cache.GroupCache, peers ...string) {
	group.RegisterPicker(s)
	s.SetPeers(peers...)
	g := s.cacheServer.Group(s.namespace, s.FilterParamsCtx())
	{
		g.GET("/:__group/get/:__key", s.GetHandle)
		g.GET("/:__group/put/:__key/:__value", s.PutHandle)
	}
	s.cacheServer.Run(fmt.Sprintf(":%v", s.port))
}

func (s *CacheServer) SetPeers(peers ...string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.peers = consistent_hash.New(defaultReplicas, nil)
	s.peers.Add(peers...)
	for _, peer := range peers {
		zlog.Info("client node url is:" + fmt.Sprintf("%s/%s", peer, s.namespace))
		s.nodeMap[peer] = &node.ClientNode{
			BaseURL: fmt.Sprintf("%s/%s", peer, s.namespace),
		}
	}
}

func (s *CacheServer) Pick(key string) (zcache.PeerGetter, bool) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if peer := s.peers.Get(key); peer != "" && peer != s.url {
		zlog.Info("[Select] peer:" + peer)
		return s.nodeMap[peer], true
	}
	return nil, false
}

func (s *CacheServer) FilterParamsCtx() gin.HandlerFunc {
	return func(c *gin.Context) {
		group := c.Param("__group")
		if len(group) == 0 {
			http.Error(c.Writer, "empty group", http.StatusNotFound)
			return
		}
		key := c.Param("__key")
		if len(key) == 0 {
			http.Error(c.Writer, "empty key", http.StatusNotFound)
			return
		}
		zlog.Info("Get group:" + group + ",key:" + key)
		s.context.group = group
		s.context.key = key
	}
}

func (s *CacheServer) GetHandle(c *gin.Context) {
	s.context.method = "get"
	group := cache.GetCache(s.context.group)
	if group == nil {
		http.Error(c.Writer, "no such group: "+s.context.group, http.StatusNotFound)
		return
	}
	zlog.Info("get group:" + group.GetName())
	v, err := group.Get(s.context.key)
	if err != nil {
		zlog.Error("get from group failed")
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
		return
	}
	zlog.Info("get value: " + v.String())
	c.Writer.Header().Set("Content-Type", "application/octet-stream")
	c.Writer.Write(v.ByteSlice())
	return
}

func (s *CacheServer) PutHandle(c *gin.Context) {
	s.context.method = "put"
}
