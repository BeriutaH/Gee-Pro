package gee

import (
	"html/template"
	"log"
	"net/http"
	"path"
	"strings"
)

// HandlerFunc 定义路由映射的处理方法
type HandlerFunc func(ctx *Context)

type (
	// Engine 作为最顶层，拥有RouterGroup所有的功能 并实现了ServeHTTP这个接口
	Engine struct {
		*RouterGroup
		router *router        // 路由映射表
		groups []*RouterGroup // 存储所有组
		// html 渲染
		htmlTemplate *template.Template
		funcMap      template.FuncMap // 自定义渲染函数
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
	// 添加中间件相关信息
	var middlewares []HandlerFunc
	for _, group := range e.groups {
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, req)
	c.handlers = middlewares // 将中间件添加到Context中
	c.engine = e
	e.router.handle(c)
}

func (e *Engine) SetFuncMap(funcMap template.FuncMap) {
	e.funcMap = funcMap
}

func (e *Engine) LoadHTMLGlob(pattern string) {
	// template.Must 简化模板创建的代码, 接收一个模板和一个错误
	// New 创建模板实例， Funcs 向模板添加自定义函数， ParseGlob 解析与给定模式匹配的所有模板文件，并将它们加载到模板中
	// 例如pattern = templates/* 则表示templates下所有的文件都将添加到模板实例中
	e.htmlTemplate = template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern))
}

// Use 将多个已定义的中间件添加到组中
func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	log.Println("执行Use，添加中间件")
	group.middlewares = append(group.middlewares, middlewares...)
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

func (group *RouterGroup) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	// 拼接URL路由路径
	absolutePath := path.Join(group.prefix, relativePath)
	// http.FileSystem 是一个接口，表示文件系统的抽象，内部实现Open方法
	// http.StripPrefix 第一个参数用于删除请求 URL 中的指定前缀，第二个参数是实际的文件路径前缀
	fileServer := http.StripPrefix(absolutePath, http.FileServer(fs)) // http.FileServer 静态文件服务器，提供目录中的文件和子目录
	return func(c *Context) {
		file := c.Param("filepath")
		log.Printf("文件路径为: %s", file)
		if _, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Req)
	}
}

// Static 文件服务器
func (group *RouterGroup) Static(relativePath string, root string) {
	log.Printf("当前文件路径: %s", root)
	handler := group.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "/*filepath")
	// 注册 GET 处理程序
	group.GET(urlPattern, handler)
}
