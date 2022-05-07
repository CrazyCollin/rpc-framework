package xclient

import (
	"log"
	"net/http"
	"strings"
	"time"
)

type GoRegistryDiscovery struct {
	*MultiServersDiscovery
	registry string
	timeout  time.Duration
	//记录从注册中心更新服务列表的时间，超时则重新获取
	lastUpdate time.Time
}

const defaultUpdateTimeout = 10 * time.Second

func NewGoRegistryDiscovery(registry string, timeout time.Duration) *GoRegistryDiscovery {
	if timeout == 0 {
		timeout = defaultUpdateTimeout
	}
	return &GoRegistryDiscovery{
		MultiServersDiscovery: NewMultiServersDiscovery(make([]string, 0)),
		registry:              registry,
		timeout:               timeout,
	}
}

//
// Refresh
// @Description: 刷新服务列表，超时则重新获取
// @receiver g
// @return error
//
func (g *GoRegistryDiscovery) Refresh() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	//服务列表没有超时，不需要刷新
	if g.lastUpdate.Add(g.timeout).After(time.Now()) {
		return nil
	}
	log.Println("rpc registry:refresh servers from registry ", g.registry)
	//从注册中心拿alive-servers列表
	resp, err := http.Get(g.registry)
	if err != nil {
		log.Println("rpc registry refresh error:", err)
		return err
	}
	servers := strings.Split(resp.Header.Get("X-GoRpc-Servers"), ",")
	g.servers = make([]string, 0, len(servers))
	for _, server := range servers {
		if strings.TrimSpace(server) != "" {
			g.servers = append(g.servers, strings.TrimSpace(server))
		}
	}
	//更新上次拿注册列表时间
	g.lastUpdate = time.Now()
	return nil
}

func (g *GoRegistryDiscovery) Update(servers []string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.servers = servers
	g.lastUpdate = time.Now()
	return nil
}

func (g *GoRegistryDiscovery) Get(mode SelectMode) (string, error) {
	//刷新服务列表
	if err := g.Refresh(); err != nil {
		return "", err
	}
	return g.MultiServersDiscovery.Get(mode)
}

func (g *GoRegistryDiscovery) GetAll() ([]string, error) {
	//刷新服务列表
	if err := g.Refresh(); err != nil {
		return nil, err
	}
	return g.MultiServersDiscovery.GetAll()
}
