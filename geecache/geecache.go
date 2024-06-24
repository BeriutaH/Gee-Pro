package geecache

import (
	"GeeCache/geecache/singleflight"
	"fmt"
	"log"
	"sync"
)

// 负责与外部交互，控制缓存存储和获取的主流程

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

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
	peers     PeerPicker
	// 使用 singleflight.Group 确保每个键仅被获取一次
	loader *singleflight.Group
}

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
		loader:    &singleflight.Group{},
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
	// 每个密钥仅被获取一次（本地或远程,无论并发调用者的数量是多少）
	view, err := g.loader.Do(key, func() (any, error) {
		if g.peers != nil {
			// 使用 PickPeer 方法选择节点，若非本机节点，则调用 getFromPeer 从远程获取
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("GeeCache] 从对端获取失败", err)
			}
		}
		// 若是本机节点或失败，则回退到 getLocally
		return g.getLocally(key)
	})
	if err == nil {
		return view.(ByteView), nil
	}
	return
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

// RegisterPeer 注册一个 PeerPicker 来选择远程对等体
func (g *Group) RegisterPeer(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// getFromPeer 使用实现了 PeerGetter 接口的 httpGetter 从访问远程节点，获取缓存值。
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, err
}
