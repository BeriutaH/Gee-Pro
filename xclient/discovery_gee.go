package xclient

import "time"

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
