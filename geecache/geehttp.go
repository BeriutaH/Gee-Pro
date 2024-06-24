package geecache

import (
	"GeeCache/geecache/consistenthash"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// 提供被其他节点访问的能力(基于http)

const (
	defaultBasePath = "/_geecache/" // 节点间通讯地址的前缀
	defaultReplicas = 50
)

type HTTPPool struct {
	self        string                 // 记录自己的地址，包括主机名/IP 和端口
	basePath    string                 // 默认是 /_geecache/
	mu          sync.Mutex             // 监控锁住 peers 和 httpGetters
	peers       *consistenthash.Map    // 类型是一致性哈希算法的 Map，用来根据具体的 key 选择节点
	httpGetters map[string]*httpGetter // 映射远程节点与对应的 httpGetter。每一个远程节点对应一个 httpGetter，因为 httpGetter 与远程节点的地址 baseURL 有关
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

// Set 更新池的对等列表
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		// 非空并且不等于自身 主机名/IP 和端口
		p.Log("获取 Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil)

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

type httpGetter struct {
	baseURL string
}

func (h *httpGetter) Get(group, key string) ([]byte, error) {
	u := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(group), url.QueryEscape(key))
	res, err := http.Get(u) // 从 ServeHTTP
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("服务端状态码: %v", res.Status)
	}
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("正在读取响应主体: %v", err)
	}
	return bytes, nil
}

var _ PeerGetter = (*httpGetter)(nil) // 检测 httpGetter 是否实现 PeerGetter 接口所有的方法
