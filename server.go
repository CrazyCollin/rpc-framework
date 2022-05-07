package gorpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"gorpc/codec"
	"io"
	"log"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"
)

const MagicNumber = 0x3bef5c

type Option struct {
	MagicNumber int
	CodecType   codec.Type
	//连接超时时间
	ConnectTimeout time.Duration
	//处理超时时间
	HandleTimeout time.Duration
}

var DefaultOption = &Option{
	MagicNumber:    MagicNumber,
	CodecType:      codec.GobType,
	ConnectTimeout: 10 * time.Second,
}

type Server struct {
	serviceMap sync.Map
}

func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()

func (server *Server) Accept(lis net.Listener) {
	for true {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc server:accept error:", err)
			return
		}
		//log.Println("rpc server:accept connection")
		go server.ServeConn(conn)
	}
}

//
// ServeConn
// @Description: 解析消息，得到序列类型
// @receiver server
// @param conn
//
func (server *Server) ServeConn(conn io.ReadWriteCloser) {
	defer func() {
		_ = conn.Close()
	}()
	var opt Option
	//依照Option格式解码
	//log.Println("rpc server:start decode option")
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server:option error:", err)
		return
	}
	//log.Println("rpc server:decode option success")
	if opt.MagicNumber != MagicNumber {
		log.Println("rpc server:invalid magic number:", opt.MagicNumber)
		return
	}
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Println("rpc server:invalid codec type:", opt.MagicNumber)
		return
	}
	//log.Println("rpc server:init server success")
	//传入conn
	server.serveCodec(f(conn), &opt)
}

var invalidRequest = struct {
}{}

//
// serveCodec
// @Description: 处理实际消息体
// @receiver server
// @param cc
//
func (server *Server) serveCodec(cc codec.Codec, opt *Option) {
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for true {
		//log.Println("rpc server:start to read a request")
		req, err := server.readRequest(cc)
		if err != nil {
			if req == nil {
				break
			}
			req.h.Error = err.Error()
			server.sendResponse(cc, req.h, invalidRequest, sending)
			continue
		}
		//log.Println("rpc server:read a request,start handle the request...")
		wg.Add(1)
		go server.handleRequest(cc, req, sending, wg, opt.ConnectTimeout)
	}
	//log.Println("rpc server:server shutdown error!!!")
	wg.Wait()
	_ = cc.Close()
}

//
// Register
// @Description: 向Server中注册服务
// @receiver server
// @param rcvr
// @return error
//
func (server *Server) Register(rcvr interface{}) error {
	s := newService(rcvr)
	if _, dup := server.serviceMap.LoadOrStore(s.name, s); dup {
		return errors.New("rpc server:service already registered:" + s.name)
	}
	return nil
}

//
// Register
// @Description: 默认Server注册Service
// @param rcvr
// @return error
//
func Register(rcvr interface{}) error {
	return DefaultServer.Register(rcvr)
}

func (server *Server) findService(serviceMethod string) (svc *service, mtype *methodType, err error) {
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		err = errors.New("rpc server: service/method request ill-formed: " + serviceMethod)
		return
	}
	serviceName, methodName := serviceMethod[:dot], serviceMethod[dot+1:]
	//找出服务实例
	svci, ok := server.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("rpc error:can't find service" + serviceName)
		return
	}
	svc = svci.(*service)
	//找出方法实例
	mtype = svc.method[methodName]
	if mtype == nil {
		err = errors.New("rpc server:can't find method" + methodName)
	}
	return
}

//
// request
// @Description: 封装请求
//
type request struct {
	h            *codec.Header
	argv, replyv reflect.Value
	mtype        *methodType
	svc          *service
}

//
// readRequestHeader
// @Description: 读消息头
// @receiver server
// @param cc
// @return *codec.Header
// @return error
//
func (server *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server: read header error:", err)
		}
		return nil, err
	}
	return &h, nil
}

//
// readRequest
// @Description: 读出消息体中的请求
// @receiver server
// @param cc
// @return *request
// @return error
//
func (server *Server) readRequest(cc codec.Codec) (*request, error) {
	h, err := server.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}
	req := &request{h: h}
	//根据header找服务方法
	req.svc, req.mtype, err = server.findService(h.ServiceMethod)
	if err != nil {
		return req, err
	}
	//创造入参实例
	req.argv = req.mtype.newArgv()
	req.replyv = req.mtype.newReplyv()

	argvi := req.argv.Interface()
	//todo 没太搞懂
	if req.argv.Type().Kind() != reflect.Ptr {
		argvi = req.argv.Addr().Interface()
	}

	if err = cc.ReadBody(argvi); err != nil {
		log.Println("rpc server:read body err:", err)
		return req, err
	}
	return req, nil

}

//
// sendResponse
// @Description: 发送回复
// @receiver server
// @param cc
// @param h
// @param body
// @param sending
//
func (server *Server) sendResponse(cc codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := cc.Write(h, body); err != nil {
		log.Println("rpc server:write response err:", err)
	}
}

//
// handleRequest
// @Description: 处理请求
// @receiver server
// @param cc
// @param req
// @param sending
// @param wg
//
func (server *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup, timeout time.Duration) {
	defer wg.Done()
	called := make(chan struct{})
	sent := make(chan struct{})
	go func() {
		//调用服务方法
		err := req.svc.call(req.mtype, req.argv, req.replyv)
		called <- struct{}{}
		if err != nil {
			req.h.Error = err.Error()
			server.sendResponse(cc, req.h, invalidRequest, sending)
			sent <- struct{}{}
			return
		}
		server.sendResponse(cc, req.h, req.replyv.Interface(), sending)
		sent <- struct{}{}
	}()
	if timeout == 0 {
		<-called
		<-sent
		return
	}
	select {
	//超时返回
	case <-time.After(timeout):
		req.h.Error = fmt.Sprintf("rpc server:handle request timeout:expect within:%s", timeout)
		server.sendResponse(cc, req.h, invalidRequest, sending)
	//call调用成功
	case <-called:
		<-sent
	}
}

const (
	connected        = "200 Connected to Go RPC"
	defaultRPCPath   = "/_gorpc_"
	defaultDebugPath = "/debug/gorpc"
)

//
// ServeHTTP
// @Description: 处理HTTP请求
// @receiver server
// @param w
// @param req
//
func (server *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "CONNECT" {
		w.Header().Set("Content-Type", "text/plain;charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = io.WriteString(w, "405 must connect\n")
		return
	}
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		log.Print("rpc hijacking ", req.RemoteAddr, ": ", err.Error())
		return
	}
	_, _ = io.WriteString(conn, "HTTP/1.0 "+connected+"\n\n")
	server.ServeConn(conn)
}

// HandleHTTP is a convenient approach for default server to register HTTP handlers
func (server *Server) HandleHTTP() {
	DefaultServer.HandleHTTP()
	http.Handle(defaultRPCPath, server)
	http.Handle(defaultDebugPath, debugHTTP{server})
	log.Println("rpc server debug path:", defaultDebugPath)
}

func Accept(lis net.Listener) {
	DefaultServer.Accept(lis)
}
