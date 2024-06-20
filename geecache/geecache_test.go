package geecache

import (
	"fmt"
	"log"
	"os"
	"testing"
)

func memoryGetter(key string) ([]byte, error) {
	data := map[string]string{
		"foo": "bar",
		"baz": "qux",
	}
	if val, ok := data[key]; ok {
		return []byte(val), nil
	}
	return nil, fmt.Errorf("key not found")
}

func fileGetter(key string) ([]byte, error) {
	return os.ReadFile(key)
}

var db = map[string]string{
	"tom":  "23",
	"jack": "13",
	"sam":  "30",
}

func TestGetter(t *testing.T) {
	var g Getter

	// 使用内存数据源
	g = GetterFunc(memoryGetter)
	data, err := g.Get("foo")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Data from memory:", string(data))
	}

	// 使用文件数据源
	g = GetterFunc(fileGetter)
	data, err = g.Get("./example.txt")
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Data from file:", string(data))
	}
}

// TestGet 测试缓存
func TestGet(t *testing.T) {
	// 用于记录每个键被加载的次数
	loadCounts := make(map[string]int, len(db))
	// 一个缓存组，名为scores， 2048字节
	gee := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("缓存中没有命中，查询数据源 key: ", key)
			if v, ok := db[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				// 记录键被加载的次数
				loadCounts[key] += 1
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s 不存在", key)
		}))
	for k, v := range db {
		// 缓存中获取值,如果出现错误或获取的值不正确,测试失败
		if view, err := gee.Get(k); err != nil || view.String() != v {
			t.Fatal("未能获取 Tom 的 值")
		}
		if _, err := gee.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("命中 %s 丢失", k)
		}
	}
	if view, err := gee.Get("unknown"); err == nil {
		t.Fatalf("测试获取一个不存在的值，如果报错，则测试失败: %s", view)
	}
	t.Logf("总共查询的次数 %v", loadCounts)
}
