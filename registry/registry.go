package registry

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type ServerItem struct {
	Addr  string
	start time.Time
}

// GeeRegistry 是一个简单的注册中心，提供以下功能
// 添加服务器并接收心跳以保持其存活
// 返回所有存活的服务器并同步删除死亡的服务器
type GeeRegistry struct {
	timeout time.Duration
	mu      sync.Mutex
	servers map[string]*ServerItem
}

const (
	defaultPath    = "/_geerpc_/registry"
	defaultTimeout = time.Minute * 5
	defaultHeader  = "X-Geerpc-Servers"
)

// New 新建一个具有超时设置的注册表实例
func New(timeout time.Duration) *GeeRegistry {
	return &GeeRegistry{
		servers: make(map[string]*ServerItem),
		timeout: timeout,
	}
}

var DefaultGeeRegister = New(defaultTimeout)

// putServer 添加服务实例，如果服务已经存在，则更新 star
func (r *GeeRegistry) putServer(addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := r.servers[addr]
	if s == nil {
		r.servers[addr] = &ServerItem{Addr: addr, start: time.Now()}
	} else {
		s.start = time.Now() // 如果服务已经存在，则更新 start 以保持活动
	}
}

// aliveServers 返回可用的服务列表，如果存在超时的服务，则删除
func (r *GeeRegistry) aliveServers() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var alive []string
	for addr, s := range r.servers {
		if r.timeout == 0 || s.start.Add(r.timeout).After(time.Now()) {
			alive = append(alive, addr)
		} else {
			delete(r.servers, addr)
		}
	}
	sort.Strings(alive) // 对字符串切片进行字典序排序
	return alive
}

// ServeHTTP GeeRegistry 采用 HTTP 协议提供服务，且所有的有用信息都承载在 HTTP Header 中
func (r *GeeRegistry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Get：返回所有可用的服务列表，Post：添加服务实例或发送心跳，都通过自定义字段 X-Geerpc-Servers 承载
	switch req.Method {
	case "GET":
		// 保持简单，服务器在 req.Header 中
		w.Header().Set(defaultHeader, strings.Join(r.aliveServers(), ","))
	case "POST":
		addr := req.Header.Get(defaultHeader)
		if addr == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.putServer(addr)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *GeeRegistry) HandleHTTP(registryPath string) {
	http.Handle(registryPath, r)
	log.Println("rpc registry path: ", registryPath)
}

func HandleHTTP() {
	DefaultGeeRegister.HandleHTTP(defaultPath)
}

func Heartbeat(registry, addr string, duration time.Duration) {
	if duration == 0 {
		duration = defaultTimeout - time.Duration(1)*time.Minute
	}

	var err error
	err = sendHeartbeat(registry, addr)
	go func() {
		t := time.NewTicker(duration) // 创建新的定时器,会按照指定的时间间隔（duration）触发事件
		for err == nil {
			<-t.C // 会阻塞直到 Ticker 触发一次，然后执行 sendHeartbeat(registry, addr)，并将返回值赋值给 err。
			err = sendHeartbeat(registry, addr)
		}
	}()
}

func sendHeartbeat(registry string, addr string) error {
	log.Println(addr, "心跳注册: ", registry)
	httpClient := &http.Client{}
	req, _ := http.NewRequest("POST", registry, nil)
	req.Header.Set(defaultHeader, addr)
	if _, err := httpClient.Do(req); err != nil {
		log.Println("rpc server: 心跳错误: ", err)
		return err
	}
	return nil
}
