package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"zcache/cache"
	"zcache/zlog"
)

type ApiServer struct {
	port          int64
	server        *gin.Engine
	registerGroup map[string]*cache.GroupCache
}

func NewApiServer(port int64, groups map[string]*cache.GroupCache) *ApiServer {
	return &ApiServer{
		port:          port,
		registerGroup: groups,
		server:        gin.Default(),
	}
}

func (s *ApiServer) StartApiServer() {
	g := s.server.Group("/api")
	g.GET("/get", s.Get)

	s.server.Run(fmt.Sprintf(":%v", s.port))
}

func (s *ApiServer) Get(c *gin.Context) {
	group := c.Request.URL.Query().Get("group")
	if len(group) == 0 {
		zlog.Error("empty group")
		return
	}
	key := c.Request.URL.Query().Get("key")
	if len(key) == 0 {
		zlog.Error("empty key")
		return
	}
	if g, ok := s.registerGroup[group]; ok {
		view, err := g.Get(key)
		if err != nil {
			http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
			return
		}
		c.Writer.Header().Set("Content-Type", "application/octet-stream")
		c.Writer.Write(view.ByteSlice())
		return
	}
	zlog.Error("group is nil")
	return
}
