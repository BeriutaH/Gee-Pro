package gee

import (
	"log"
	"net/http"
)

// HandlerFunc 定义路由映射的处理方法
type HandlerFunc func(ctx *Context)

type (
	// Engine 作为最顶层，拥有RouterGroup所有的功能 并实现了ServeHTTP这个接口
	Engine struct {
		*RouterGroup
		router *router        // 路由映射表
		groups []*RouterGroup // 存储所有组
	}

	// RouterGroup 操作路由的方法都转移到当前类型下
	RouterGroup struct {
		prefix      string        // 路由组的前缀
		middlewares []HandlerFunc // 支持中间件
		parent      *RouterGroup  // 支持嵌套
		engine      *Engine       // 所有组共享一个Engine实例
	}
)

// New 是 Engine 构造函数
func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine} // 循环嵌套，Engine 作为所有 RouterGroup 的统一管理者
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
}

func (e *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, e)
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := newContext(w, req)
	e.router.handle(c)
}

// Group 定义组来创建新的 RouterGroup
func (group *RouterGroup) Group(prefix string) *RouterGroup {

	engine := group.engine //所有组共享同一个 Engine 实例

	newGroup := &RouterGroup{
		prefix: group.prefix + prefix, // 当前组的前缀再加上新的前缀 变为新建组的前缀
		parent: group,                 // 当前组
		engine: engine,                // 当前Engine实例
	}
	engine.groups = append(engine.groups, newGroup) // 把新建的组添加到engine实例中
	return newGroup
}

func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp                // 路由前缀加上后半段路径，组成完整的路由路径
	log.Printf("Route %4s - %s", method, pattern) // 查看整体的路由路径
	group.engine.router.addRoute(method, pattern, handler)
}

func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}
