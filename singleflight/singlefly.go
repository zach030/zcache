package singleflight

import "sync"

// 正在进行中或已结束的请求，通过wg
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// 管理不同key的请求
type Group struct {
	lock  sync.Mutex
	calls map[string]*call
}

//第一个get(key)请求到来时，single-fly会记录当前key正在被处理，
//后续的请求只需要等待第一个请求处理完成，取返回值即可
func (g *Group) DoSingle(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.lock.Lock()
	if g.calls == nil {
		g.calls = make(map[string]*call)
	}
	if v, ok := g.calls[key]; ok {
		// 如果正有请求，则wait
		g.lock.Unlock()
		v.wg.Wait()
		// 请求一并返回
		return v.val, nil
	}
	// 发起请求
	c := new(call)
	c.wg.Add(1)
	// 记录正在请求
	g.calls[key] = c
	g.lock.Unlock()

	// 请求返回
	c.val, c.err = fn()
	c.wg.Done()

	g.lock.Lock()
	// 删除当前请求记录
	delete(g.calls, key)
	g.lock.Unlock()

	return c.val, c.err
}
