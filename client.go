package gorpc

import (
	"errors"
	"gorpc/codec"
	"sync"
)

type Call struct {
	Seq           uint64
	ServiceMethod string
	Args          interface{}
	Reply         interface{}
	Error         error
	Done          chan *Call
}

//
// done
// @Description: 支持异步调用，当调用结束的时候通知调用方
// @receiver call
//
func (call *Call) done() {
	call.Done <- call
}

type Client struct {
	cc      codec.Codec
	opt     *Option
	sending *sync.Mutex
	header  codec.Header
	mu      *sync.Mutex
	seq     uint64
	pending map[uint64]*Call
	//表示用户主动关闭Client
	closing bool
	//表示有错误发生关闭Client
	shutdown bool
}

var ErrShutdown = errors.New("connection is shutdown")

//
// Close
// @Description: 关闭连接
// @receiver client
// @return error
//
func (client *Client) Close(call *Call) error {
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.closing {
		return ErrShutdown
	}
	client.closing = true
	return client.cc.Close()
}

//
// IsAvailable
// @Description: 查看Client是否可用
// @receiver client
// @return bool
//
func (client *Client) IsAvailable() bool {
	client.mu.Lock()
	defer client.mu.Unlock()
	return !client.shutdown && !client.closing
}

//
// registerCall
// @Description: 在Client中注册Call
// @receiver client
// @param call
// @return uint64
// @return error
//
func (client Client) registerCall(call *Call) (uint64, error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.closing || client.shutdown {
		return 0, ErrShutdown
	}
	call.Seq = client.seq
	client.pending[call.Seq] = call
	client.seq++
	return call.Seq, nil
}

//
// removeCall
// @Description: 在Client删除并返回指定Call
// @receiver client
// @param seq
// @return *Call
//
func (client *Client) removeCall(seq uint64) *Call {
	client.mu.Lock()
	defer client.mu.Unlock()
	call := client.pending[seq]
	delete(client.pending, seq)
	return call
}

//
// terminateCalls
// @Description: 通知所有call队列有客户端或者服务端调用错误发生
// @receiver client
// @param err
//
func (client Client) terminateCalls(err error) {
	client.sending.Lock()
	defer client.sending.Unlock()
	client.mu.Lock()
	defer client.mu.Unlock()
	client.shutdown = true
	for _, call := range client.pending {
		call.Error = err
		call.done()
	}
}

func (client Client) receive() {
	var err error
	for err == nil {

	}
}
