package geecache

import (
	"fmt"
	"log"
	"sync"
)

// 负责与外部交互，控制缓存存储和获取的主流程

// Getter 根据某个键加载数据
type Getter interface {
	Get(key string) ([]byte, error)
}

// GetterFunc 适配器模式，避免了为每个实现创建单独的结构体
type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name      string
	getter    Getter
	mainCache cache
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup 是 Group 的构造函数
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("Getter 不能为空")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,                          // 缓存的命名空间, 每个 Group 拥有一个唯一的名称 name
		getter:    getter,                        // 缓存未命中时获取源数据的回调
		mainCache: cache{cacheBytes: cacheBytes}, // 并发缓存
	}
	groups[name] = g
	return g
}

// GetGroup 返回先前使用 NewGroup 创建的命名组，如果不存在这样的组，则返回 nil
func GetGroup(name string) *Group {
	mu.RLock() // 只读锁，读的本质也是并发的操作map，如果不加锁，并发量大可能会panic
	defer mu.RUnlock()
	g := groups[name]
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	/*
		1. 从 mainCache 中查找缓存，如果存在则返回缓存值
		2. 缓存不存在，则调用 load 方法，load 调用 getLocally（分布式场景下会调用 getFromPeer 从其他节点获取）
	*/
	if key == "" {
		return ByteView{}, fmt.Errorf("key是必填参数")
	}
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] 命中")
		return v, nil
	}
	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	return g.getLocally(key)
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key) // 获取源数据
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)} // 将源数据添加到缓存
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value) // cache 添加到并发缓存
}
