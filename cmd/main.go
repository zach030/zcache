package main

import (
	"flag"
	"fmt"
	"zcache/cache"
	"zcache/server"
	"zcache/zlog"
)

var db = map[string]string{
	"zach": "98",
	"sia":  "80",
	"tom":  "70",
}

func createGroup() *cache.GroupCache {
	return cache.NewGroupCache("scores", 100, cache.GetterFunc(func(key string) ([]byte, error) {
		zlog.Info("[SlowDB] search key:" + key)
		if v, ok := db[key]; ok {
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not exist", key)
	}))
}

var (
	addrMap = map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}
)

func main() {
	var port int64
	flag.Int64Var(&port, "port", 8000, "cache server port")
	group := createGroup()
	// 注册主Server
	cacheServer1 := server.NewCacheServer(8001, "")
	cacheServer2 := server.NewCacheServer(8002, "")
	cacheServer3 := server.NewCacheServer(8003, "")

	// 设置节点服务
	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	go cacheServer1.Start(group, addrs...)
	go cacheServer2.Start(group, addrs...)
	go cacheServer3.Start(group, addrs...)

	apiServer := server.NewApiServer(port, map[string]*cache.GroupCache{
		group.GetName(): group,
	})
	apiServer.StartApiServer()
}
