package geecache

// 并发控制

import (
	"GeeCache/geecache/lru"
	"sync"
)

type cache struct {
	// 之所以不用读写锁 cache 的 get 和 add 都涉及到写操作(LRU 将最近访问元素移动到链表头)，所以不能直接改为读写锁
	//如果 cache 侧和 LRU 侧同时使用锁细颗粒度控制，是有优化空间的，可以尝试下
	mu         sync.Mutex
	lru        *lru.Cache
	cacheBytes int64
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// 判断 c.lru 是否为 nil，如果等于 nil 再创建实例
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil) // 延迟初始化
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
