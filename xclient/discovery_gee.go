package xclient

import (
	"GeeRPC/registry"
	"log"
	"net/http"
	"strings"
	"time"
)

type GeeRegistryDiscovery struct {
	*MultiServersDiscovery               // 嵌套了 MultiServersDiscovery，很多能力可以复用
	registry               string        // 注册中心的地址
	timeout                time.Duration // 服务列表的过期时间
	lastUpdate             time.Time     // 代表最后从注册中心更新服务列表的时间，默认 10s 过期之后，需要从注册中心更新新的列表
}

const defaultUpdateTimeout = time.Second * 10

func NewGeeRegistryDiscovery(registerAddr string, timeout time.Duration) *GeeRegistryDiscovery {
	if timeout == 0 {
		timeout = defaultUpdateTimeout
	}
	d := &GeeRegistryDiscovery{
		MultiServersDiscovery: NewMultiServerDiscovery(make([]string, 0)),
		registry:              registerAddr,
		timeout:               timeout,
	}
	return d
}

func (d GeeRegistryDiscovery) Update(servers []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.servers = servers
	d.lastUpdate = time.Now()
	return nil
}

func (d GeeRegistryDiscovery) Refresh() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.lastUpdate.Add(d.timeout).After(time.Now()) {
		return nil
	}
	log.Println("rpc registry: 从注册表刷新服务器", d.registry)
	resp, err := http.Get(d.registry)
	if err != nil {
		log.Println("rpc registry 报错: ", err)
		return err
	}
	servers := strings.Split(resp.Header.Get(registry.DefaultHeader), ",")
	d.servers = make([]string, 0, len(servers))
	for _, server := range servers {
		serverTrim := strings.TrimSpace(server) // 去除字符串 server 两端的空白字符
		if serverTrim != "" {
			d.servers = append(d.servers, serverTrim)
		}
	}
	d.lastUpdate = time.Now()
	return nil
}

// Get 和 GetAll 与 MultiServersDiscovery 相似，唯一的不同在于，GeeRegistryDiscovery 需要先调用 Refresh 确保服务列表没有过期
func (d GeeRegistryDiscovery) Get(mode SelectMode) (string, error) {
	if err := d.Refresh(); err != nil {
		return "", err
	}
	return d.MultiServersDiscovery.Get(mode)
}
func (d GeeRegistryDiscovery) GetAll() ([]string, error) {
	if err := d.Refresh(); err != nil {
		return nil, err
	}
	return d.MultiServersDiscovery.GetAll()
}
