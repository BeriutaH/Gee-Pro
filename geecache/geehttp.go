package geecache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// 提供被其他节点访问的能力(基于http)

const defaultBasePath = "/_geecache/" // 节点间通讯地址的前缀

type HTTPPool struct {
	self     string // 记录自己的地址，包括主机名/IP 和端口
	basePath string
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath, // 默认是 /_geecache/
	}
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP 处理所有http请求
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 判断访问路径的前缀是否跟 basePath 相等
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// /<basepath>/<groupname>/<key>  将结果按"/"进行分割，最多分割成2部分
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)

	log.Println("parts: ", parts)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	log.Println("组名: ", groupName)
	key := parts[1]

	group := GetGroup(groupName) // 根据组名获取组的对象
	if group == nil {
		http.Error(w, "未查询到组: "+groupName, http.StatusNotFound)
		return
	}
	view, err := group.Get(key) // 获取缓存数据
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// application/octet-stream 通用的二进制数据类型，表示响应体是未解释的二进制数据
	// 浏览器和其他客户端通常会将这样的内容视为文件进行下载，而不是直接显示
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice()) // 将获取的缓存数据返回
}
