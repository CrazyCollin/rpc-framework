package xclient

import (
	"context"
	. "gorpc"
	"reflect"
	"sync"
)

type XClient struct {
	d    Discovery
	mode SelectMode
	opt  *Option
	mu   sync.Mutex
	//复用已经建立连接的client，减少资源消耗
	clients map[string]*Client
}

func NewXClient(d Discovery, mode SelectMode, opt *Option) *XClient {
	return &XClient{d: d, mode: mode, opt: opt, clients: make(map[string]*Client)}
}

func (c *XClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	for key, client := range c.clients {
		_ = client.Close()
		delete(c.clients, key)
	}
	return nil
}

func (c *XClient) dial(rpcAddr string) (*Client, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	client, ok := c.clients[rpcAddr]
	//client不可用
	if ok && !client.IsAvailable() {
		_ = client.Close()
		delete(c.clients, rpcAddr)
		client = nil
	}
	//重新创建新client
	if client == nil {
		var err error
		client, err = XDial(rpcAddr, c.opt)
		if err != nil {
			return nil, err
		}
		c.clients[rpcAddr] = client
	}
	return client, nil
}

func (c *XClient) call(rpcAddr string, ctx context.Context, serviceMethod string, args, reply interface{}) error {
	client, err := c.dial(rpcAddr)
	if err != nil {
		return err
	}
	return client.Call(ctx, serviceMethod, args, reply)
}

func (c *XClient) Call(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	rpcAddr, err := c.d.Get(c.mode)
	if err != nil {
		return err
	}
	return c.call(rpcAddr, ctx, serviceMethod, args, reply)
}

//
// Broadcast
// @Description: 将请求广播到所有服务实例当中
// @receiver c
// @param ctx
// @param serviceMethod
// @param args
// @param reply
// @return error
//
func (c *XClient) Broadcast(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	//拿到所有服务实例
	servers, err := c.d.GetAll()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	var mu sync.Mutex
	var e error
	replyDone := reply == nil
	//创建子context
	ctx, cancel := context.WithCancel(ctx)
	for _, rpcAddr := range servers {
		wg.Add(1)
		go func(rpcAddr string) {
			defer wg.Done()
			var clonedReply interface{}
			if reply != nil {
				clonedReply = reflect.New(reflect.ValueOf(reply).Elem().Type()).Interface()
			}
			err := c.call(rpcAddr, ctx, serviceMethod, args, clonedReply)
			//保证并发情况下error和reply能够被正确赋值
			mu.Lock()
			if err != nil && e == nil {
				e = err
				//直接终止在call失败情况下的其他call
				cancel()
			}
			if err == nil && !replyDone {
				reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(clonedReply).Elem())
				replyDone = true
			}
			mu.Unlock()
		}(rpcAddr)
	}
	wg.Wait()
	return e
}
