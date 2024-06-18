package gee

import (
	"reflect"
	"testing"
)

// 命令 go test -v ./gee

func newTestRouter() *router {
	r := newRouter()
	r.addRoute("GET", "/", nil)
	r.addRoute("GET", "/hello/:name", nil)
	r.addRoute("GET", "/hello/b/c", nil)
	r.addRoute("GET", "/hi/:name", nil)
	r.addRoute("GET", "/assets/*filepath", nil)
	return r
}

func TestParsePattern(t *testing.T) {
	// reflect.DeepEqual 接受两个空接口类型的参数，返回一个布尔值，表示两个参数是否深度相等
	ok := reflect.DeepEqual(parsePattern("/p/:name"), []string{"p", ":name"})
	ok = ok && reflect.DeepEqual(parsePattern("/p/*"), []string{"p", "*"})
	ok = ok && reflect.DeepEqual(parsePattern("/p/*name/*"), []string{"p", "*name"})
	// 这三个其中任何一个不为true，都为失败
	if !ok {
		t.Fatal("test parsePattern failed")
	}
}

func TestGetRoute(t *testing.T) {
	r := newTestRouter()
	n, ps := r.getRoute("GET", "/hello/beriuta")
	if n == nil {
		t.Fatal("不应返回nil")
	}
	if n.pattern != "/hello/:name" {
		t.Fatal("应该匹配/hello/:name")
	}
	if ps["name"] != "beriuta" {
		t.Fatal("名称应该等于 beriuta")
	}
	t.Logf("matched path: %s, params['name']: %s\n", n.pattern, ps["name"])

}

func TestGetRoutes(t *testing.T) {
	r := newTestRouter()
	t.Log("--------------------------")
	nodes := r.getRoutes("GET")
	for i, n := range nodes {
		t.Logf("%d, node: %+v", i+1, n)
	}

	if len(nodes) != 5 {
		t.Fatal("测试错误！")
	}

}
