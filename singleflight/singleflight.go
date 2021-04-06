package singleflight

import "sync"

// 进行中或已结束的请求
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	mu sync.Mutex       // protects m
	m  map[string]*call
}

// 针对相同的 key，无论 Do 被调用多少次，函数 fn 都只会被调用一次，等待 fn 调用结束了，返回返回值或错误。
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	if c, ok := g.m[key]; ok {
		c.wg.Wait()   // 如果请求正在进行中，则等待
		return c.val, c.err  // 请求结束，返回结果
	}
	c := new(call)
	c.wg.Add(1)       // 发起请求前加锁
	g.m[key] = c      // 添加到 g.m，表明 key 已经有对应的请求在处理

	c.val, c.err = fn() // 调用 fn，发起请求
	c.wg.Done()         // 请求结束

	delete(g.m, key)    // 更新 g.m

	return c.val, c.err // 返回结果
}