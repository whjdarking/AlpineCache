package onlySendOnce

import (
	"sync"
)

type call struct {
	waitGroup sync.WaitGroup
	val       interface{}
	err       error
}

type Group struct {
	mu sync.Mutex       //虽然sendOnce的目的是让并发查询不查多次，但作为标志的m还是要锁起来不让它被并发读写
	m  map[string]*call //string对应key
}

//once代表只想执行一次的函数，这里指查询
func (g *Group) Do(key string, once func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock() //保护m不被并发读写
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	//如果有值，说明已经有请求在处理，只需等待
	if calling, ok := g.m[key]; ok {
		g.mu.Unlock()
		calling.waitGroup.Wait()
		return calling.val, calling.err
	}
	calling := new(call)
	calling.waitGroup.Add(1)
	g.m[key] = calling //加入m中，代表这个key已经在处理了，随后就可以释放m
	g.mu.Unlock()

	calling.val, calling.err = once() //核心，执行查询
	calling.waitGroup.Done()

	g.mu.Lock()
	delete(g.m, key) //已经没有在处理了，所以从m中删除
	g.mu.Unlock()

	return calling.val, calling.err
}
