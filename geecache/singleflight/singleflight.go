package singleflight

import "sync"

// call 正在进行中，或已经结束的请求
type call struct {
	wg  sync.WaitGroup // 避免重入
	val any
	err error
}

// Group singleflight 的主数据结构，管理不同 key 的请求(call)
type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

// Do 针对相同的 key，无论 Do 被调用多少次，函数 fn 都只会被调用一次，等待 fn 调用结束了，返回返回值或错误
func (g *Group) Do(key string, fn func() (any, error)) (any, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	// 并发时，如果查询的key已经存在了，将会等待第一个请求完成，并直接返回第一个请求的结果
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()         // 等待所有的 goroutine 完成, 阻塞，直到锁被释放
		return c.val, c.err // 请求结束，返回结果
	}
	// 第一个get(key)请求到来时，singleflight会记录当前key正在被处理，后续的请求只需要等待第一个请求处理完成，取返回值即可
	c := new(call)
	c.wg.Add(1)  // 表示我们将启动一个新的 goroutine
	g.m[key] = c // 添加到 g.m，表明 key 已经有对应的请求在处理
	g.mu.Unlock()

	c.val, c.err = fn() // 调用 fn，发起请求
	c.wg.Done()         // 请求结束 每个 goroutine 在完成其任务后应调用 Done 方法，以表示该 goroutine 已经完成

	g.mu.Lock()
	delete(g.m, key) // 更新 g.m, 不删除，如果key对应的值变化，所得的值还是旧值
	g.mu.Unlock()
	return c.val, c.err
}
