package gee

import (
	"testing"
)

// go test -v ./gee -run TestNestedGroup

// 测试嵌套组
func TestNestedGroup(t *testing.T) {
	r := New()
	v1 := r.Group("/v1")
	v2 := v1.Group("/v2")
	v3 := v2.Group("/v3")
	if v2.prefix != "/v1/v2" {
		t.Fatal("v2的完整路由应该为: /v1/v2 ")
	}
	if v3.prefix != "/v1/v2/v3" {
		t.Fatal("v2的完整路由应该为: /v1/v2/v3 ")

	}
	t.Logf("v3 完整路由为: %s", v3.prefix)
}

// 测试文件路由
func TestStaticPath(t *testing.T) {
	r := New()
	r.Static("/assets", "./static")
	r.Run(":9999")
}
