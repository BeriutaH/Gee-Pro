package lru

import (
	"fmt"
	"reflect"
	"testing"
)

type String string

func (s String) Len() int {
	return len(s)
}

func TestGet(t *testing.T) {
	lru := New(int64(0), nil)
	lru.Add("key1", String("1234"))
	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("如果没有获取到并且值不等于1234，缓存命中失败")
	}
	if _, ok := lru.Get("key8"); ok {
		t.Fatalf("如果获取到了，则报错")
	}
}

func TestRemoveOldest(t *testing.T) {
	k1, k2, k3 := "k1", "k2", "k3"
	v1, v2, v3 := "v1", "v2", "v3"
	cap1 := len(k1 + k2 + k3 + v1 + v2)
	lru := New(int64(cap1), nil)
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3))
	if _, ok := lru.Get("k1"); ok || lru.Len() != 2 {
		t.Fatalf("移除k1失败")
	}
}

// 测试回调函数
func TestOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}
	lru := New(int64(10), callback)
	lru.Add("key1", String("123456"))
	lru.Add("k2", String("k2"))
	lru.Add("k3", String("k3"))
	lru.Add("k4", String("k4"))

	expect := []string{"key1", "k2"}
	// 深度比较两个值是否相等
	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("调用 OnEvicted 失败，预期键: %s", expect)
	}
	fmt.Println(keys)
}
