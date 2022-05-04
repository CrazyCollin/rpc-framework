package gorpc

import (
	"encoding/json"
	"fmt"
	"gorpc/codec"
	"io"
	"log"
	"net"
	"reflect"
	"sync"
)

const MagicNumber = 0x3bef5c

type Option struct {
	MagicNumber int
	CodecType   codec.Type
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}

type Server struct {
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
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server:option error:", err)
		return
	}
	if opt.MagicNumber != MagicNumber {
		log.Println("rpc server:invalid magic number:", opt.MagicNumber)
		return
	}
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Println("rpc server:invalid codec type:", opt.MagicNumber)
		return
	}
	//传入conn
	server.serveCodec(f(conn))
}

var invalidRequest = struct {
}{}

//
// serveCodec
// @Description: 处理实际消息体
// @receiver server
// @param cc
//
func (server *Server) serveCodec(cc codec.Codec) {
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for true {
		req, err := server.readRequest(cc)
		if err != nil {
			if req == nil {
				break
			}
			req.h.Error = err
			server.sendResponse(cc, req.h, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		go server.handleRequest(cc, req, sending, wg)
	}
	wg.Wait()
	_ = cc.Close()
}

//
// request
// @Description: 封装请求
//
type request struct {
	h            *codec.Header
	argv, replyv reflect.Value
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
			log.Println("rpc server:read head error:", err)
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

	req.argv = reflect.New(reflect.TypeOf(""))
	if err = cc.ReadBody(req.argv.Interface()); err != nil {
		log.Println("rpc server:read argv err:", err)
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
func (server *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println(req.h, req.argv.Elem())
	req.replyv = reflect.ValueOf(fmt.Sprintf("gorpc resp %d", req.h.Seq))
	server.sendResponse(cc, req.h, req.replyv.Interface(), sending)
}

func Accept(lis net.Listener) {
	DefaultServer.Accept(lis)
}
