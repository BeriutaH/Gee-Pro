package lru

// lru 缓存淘汰策略

import (
	"container/list"
	"fmt"
)

type Value interface {
	Len() int
}

// entry 是双向链表节点的数据类型，在链表中仍保存每个值对应的 key 的好处在于，淘汰队首节点时，需要用 key 从字典中删除对应的映射
type entry struct {
	key   string
	value Value
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

func (c *Cache) Get(key string) (Value Value, ok bool) {
	/*
		1. 从字典中找到对应的双向链表的节点
		2. 将该节点移动到队尾
	*/
	if ele, ok := c.cache[key]; ok {
		// 键对应的链表节点存在，则将对应节点移动到队尾，并返回查找到的值
		c.ll.MoveToFront(ele)    // 将链表中的节点 ele 移动到队尾
		kv := ele.Value.(*entry) // 类型断言，将entry指针存入，并返回一个*entry类型的值
		return kv.value, true
	}
	return
}

// RemoveOldest 移除最近最少访问的节点（队首）
func (c *Cache) RemoveOldest() {
	fmt.Println("调用移除")
	ele := c.ll.Back() // 取到队首节点，从链表中删除
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)                                   // 字典中 c.cache 删除该节点的映射关系
		c.userBytes -= int64(len(kv.key)) + int64(kv.value.Len()) // 更新当前所用的内存
		// 如果OnEvicted 不为 nil，则调用回调函数
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.userBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key: key, value: value})
		c.cache[key] = ele
		c.userBytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.userBytes {
		// 缓存容量超出限制，移除最少访问的节点(刚刚更新的数据也可能会被淘汰)
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
