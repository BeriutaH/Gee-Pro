package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

// Map 包含所有哈希键
type Map struct {
	hash     Hash           // 采取依赖注入的方式，允许用于替换成自定义的 Hash 函数
	replicas int            // 虚拟节点倍数
	keys     []int          // 哈希环
	hashMap  map[int]string // 虚拟节点与真实节点的映射表, 键是虚拟节点的哈希值, 值是真实节点的名称
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE // 计算 CRC-32 校验和
	}
	return m
}

// Add 允许传入 0 或 多个真实节点的名称
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		// 对每一个真实节点key 创建m.replicas个虚拟节点
		for i := 0; i < m.replicas; i++ {
			// 虚拟节点的名称是：strconv.Itoa(i) + key
			hashKey := int(m.hash([]byte(strconv.Itoa(i) + key))) // m.hash() 计算虚拟节点的哈希值
			m.keys = append(m.keys, hashKey)                      // 添加到环上
			m.hashMap[hashKey] = key                              // 增加虚拟节点和真实节点的映射关系
		}
	}
	// sort.Ints 对[]int 类型的切片排序
	sort.Ints(m.keys) // 环上的哈希值排序
}

// Get 获取哈希中与所提供的键最接近的项
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hashKey := int(m.hash([]byte(key))) // 计算 key 的哈希值
	// 顺时针找到第一个匹配的虚拟节点的下标 idx
	// 如果需要找到的元素确实存在，那么 idx 将是该元素在切片中的索引
	// 如果需要找到的元素不存在，那么 idx 将是该元素应该插入的位置。
	idx := sort.Search(len(m.keys), func(i int) bool {
		// 找到第一个大于 hashKey 的值的索引，假如 hashKey = 6，那就找到 7 元素所在的索引
		return m.keys[i] >= hashKey
	})
	// 取模运算 idx % len(m.keys) 来确保索引在有效范围内
	// m.keys[idx % len(m.keys)] 取 m.keys 切片中对应位置的值
	// 将 m.keys 切片中的值作为键，从 m.hashMap 中获取相应的值
	return m.hashMap[m.keys[idx%len(m.keys)]] // 映射得到真实的节点
}
