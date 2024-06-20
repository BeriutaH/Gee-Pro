package consistenthash

import (
	"fmt"
	"strconv"
	"testing"
)

//go test -v ./geecache/consistenthash/ -run TestHashing

func TestHashing(t *testing.T) {
	hash := New(3, func(key []byte) uint32 {
		i, _ := strconv.Atoi(string(key))
		return uint32(i)
	})
	hash.Add("6", "5", "2")
	// 所有的hash节点
	fmt.Println(hash.keys) // [2 5 6 12 15 16 22 25 26]
	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "5",
		"27": "2",
	}
	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}
	hash.Add("8")
	testCases["27"] = "8"
	fmt.Println(hash.keys) // [2 5 6 8 12 15 16 18 22 25 26 28]
	for k, v := range testCases {
		if n := hash.Get(k); n != v {
			t.Errorf("要求 %s,，应该得到 %s", k, n)
		}
	}
}
