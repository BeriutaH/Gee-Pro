package lru

import "container/list"

type Value interface {
	Len() int
}

// entry 是双向链表节点的数据类型，在链表中仍保存每个值对应的 key 的好处在于，淘汰队首节点时，需要用 key 从字典中删除对应的映射
type entry struct {
	key   string
	Value Value
}

type Cache struct {
	maxBytes  int64                         //允许使用的最大内存
	userBytes int64                         // 当前已使用的内存
	ll        *list.List                    // 标准库实现的双向链表
	cache     map[string]*list.Element      //键是字符串，值是双向链表中对应节点的指针
	OnEvicted func(key string, value Value) // 某条记录被移除时的回调函数，可以为 nil
}

// New 是Cache的构造函数
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}
