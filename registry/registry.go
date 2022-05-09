package registry

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

//
// GoRegistry
// @Description: 简易的服务注册中心，和服务注册方有心跳连接
//
type GoRegistry struct {
	timeout time.Duration
	mu      sync.Mutex
	servers map[string]*ServerItem
}

type ServerItem struct {
	Addr  string
	start time.Time
}

const (
	defaultPath    = "/_gorpc_/registry"
	defaultTimeout = 5 * time.Minute
)

func NewGoRegistry(timeout time.Duration) *GoRegistry {
	return &GoRegistry{
		timeout: timeout,
		servers: make(map[string]*ServerItem),
	}
}

var DefaultGoRegistry = NewGoRegistry(defaultTimeout)

//
// putServer
// @Description: 在注册中心中注册服务提供方
// @receiver r
// @param addr
//
func (r *GoRegistry) putServer(addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := r.servers[addr]
	if s == nil {
		//注册服务，记录服务地址和时间
		r.servers[addr] = &ServerItem{Addr: addr, start: time.Now()}
	} else {
		//响应心跳连接
		s.start = time.Now()
	}
}

//
// aliveServers
// @Description: 查看当前注册中心保活服务注册方
// @receiver r
// @return []string
//
func (r *GoRegistry) aliveServers() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var alive []string
	for addr, s := range r.servers {
		//判断是否超时
		if r.timeout == 0 || s.start.Add(r.timeout).After(time.Now()) {
			alive = append(alive, addr)
		} else {
			//服务注册方失活，从注册中心列表中删除
			delete(r.servers, addr)
		}
	}
	sort.Strings(alive)
	return alive
}

//
// ServeHTTP
// @Description: 实现Handler接口
// @receiver r
// @param w
// @param req
//
func (r *GoRegistry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		w.Header().Set("X-GoRpc-Servers", strings.Join(r.aliveServers(), ","))
	case "POST":
		addr := req.Header.Get("X-GoRpc-Servers")
		if addr == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.putServer(addr)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *GoRegistry) HandleHTTP(registryPath string) {
	http.Handle(registryPath, r)
	log.Println("rpc registry path:", registryPath)
}

func HandleHTTP() {
	DefaultGoRegistry.HandleHTTP(defaultPath)
}

//
// Heartbeat
// @Description:
// @param registry
// @param addr
// @param duration
//
func Heartbeat(registry, addr string, duration time.Duration) {
	if duration == 0 {
		duration = defaultTimeout - time.Duration(1)*time.Minute
	}
	var err error
	err = sendHeartbeat(registry, addr)
	go func() {
		t := time.NewTicker(duration)
		for err == nil {
			//打点器每4min发送一次心跳
			<-t.C
			err = sendHeartbeat(registry, addr)
		}
	}()
}

//
// sendHeartbeat
// @Description: 向注册中心发出心跳信息
// @param registry
// @param addr
// @return error
//
func sendHeartbeat(registry, addr string) error {
	log.Println(addr, " send heart beat to registry ", registry)
	httpClient := &http.Client{}
	req, _ := http.NewRequest("POST", registry, nil)
	req.Header.Set("X-GoRpc-Servers", addr)
	if _, err := httpClient.Do(req); err != nil {
		log.Println("rpc server:heart beat error:", err)
		return err
	}
	return nil
}
