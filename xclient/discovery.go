package xclient

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"
)

type SelectMode int

const (
	RandomSelect SelectMode = iota
	RoundRobinSelect
)

type Discovery interface {
	Refresh() error
	Update(servers []string) error
	Get(mode SelectMode) (string, error)
	GetAll() ([]string, error)
}

// MultiServersDiscovery 是针对没有注册中心的多服务器的发现，用户需要明确提供服务器地址
type MultiServersDiscovery struct {
	r       *rand.Rand   // 生成随机数
	mu      sync.RWMutex // 保护以下
	servers []string
	index   int // 记录罗宾算法选定的位置
}

// Refresh 刷新无意义，故忽略
func (d *MultiServersDiscovery) Refresh() error {
	return nil
}

// Update 动态更新发现的服务器
func (d *MultiServersDiscovery) Update(servers []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.servers = servers
	return nil
}

// Get 根据模式获取服务器
func (d *MultiServersDiscovery) Get(mode SelectMode) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	n := len(d.servers)
	if n == 0 {
		return "", errors.New("rpc discovery: 没有可用的服务器")
	}
	switch mode {
	case RandomSelect:
		return d.servers[d.r.Intn(n)], nil
	case RoundRobinSelect:
		s := d.servers[d.index%n] // 服务器可能会更新，因此模式 n 以确保安全
		d.index = (d.index + 1) % n
		return s, nil
	default:
		return "", errors.New("rpc discovery: 不支持选择模式")
	}
}

// GetAll 返回发现中的所有服务器
func (d *MultiServersDiscovery) GetAll() ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	servers := make([]string, len(d.servers), len(d.servers))
	copy(servers, d.servers)
	return servers, nil
}

// NewMultiServerDiscovery 创建 MultiServersDiscovery 实例
func NewMultiServerDiscovery(servers []string) *MultiServersDiscovery {
	d := &MultiServersDiscovery{
		servers: servers,
		r:       rand.New(rand.NewSource(time.Now().UnixNano())), // 初始化时使用时间戳设定随机数种子，避免每次产生相同的随机数序列
	}
	d.index = d.r.Intn(math.MaxInt32 - 1) // 记录 Round Robin 算法已经轮询到的位置，为了避免每次从 0 开始，初始化时随机设定一个值
	return d
}

var _ Discovery = (*MultiServersDiscovery)(nil)
