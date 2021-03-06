package gorpc

import (
	"go/ast"
	"log"
	"reflect"
	"sync/atomic"
)

type methodType struct {
	//方法本身
	method reflect.Method
	//第一个参数类型
	ArgType reflect.Type
	//第二个参数类型
	ReplyType reflect.Type
	//后续统计方法调用次数时使用
	numCalls uint64
}

func (m *methodType) NumCalls() uint64 {
	return atomic.LoadUint64(&m.numCalls)
}

func (m *methodType) newArgv() reflect.Value {
	var argv reflect.Value
	if m.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(m.ArgType.Elem())
	} else {
		argv = reflect.New(m.ArgType).Elem()
	}
	return argv
}

func (m *methodType) newReplyv() reflect.Value {
	replyv := reflect.New(m.ReplyType.Elem())
	switch m.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(m.ReplyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(m.ReplyType.Elem(), 0, 0))
	}
	return replyv
}

type service struct {
	//映射注册服务结构体名称
	name string
	//结构体类型
	typ reflect.Type
	//结构体实例本身
	rcvr reflect.Value
	//存储注册服务结构体所注册的方法
	method map[string]*methodType
}

//
// newService
// @Description: 构造service实例
// @param rcvr 需要映射为服务的结构体实例
// @return *service
//
func newService(rcvr interface{}) *service {
	s := new(service)
	s.rcvr = reflect.ValueOf(rcvr)
	s.name = reflect.Indirect(s.rcvr).Type().Name()
	s.typ = reflect.TypeOf(rcvr)
	if !ast.IsExported(s.name) {
		log.Fatalf("rpc server:%s is not a valid service name", s.name)
	}
	s.registerMethods()
	return s
}

//
// registerMethods
// @Description:
// @receiver s
//
func (s *service) registerMethods() {
	s.method = make(map[string]*methodType)
	for i := 0; i < s.typ.NumMethod(); i++ {
		method := s.typ.Method(i)
		mType := method.Type
		//过滤符合条件的方法
		if mType.NumIn() != 3 || mType.NumOut() != 1 {
			continue
		}
		if mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}
		argType, replyType := mType.In(1), mType.In(2)
		if !isExportedOrBuiltinType(argType) || !isExportedOrBuiltinType(replyType) {
			continue
		}
		s.method[method.Name] = &methodType{
			method:    method,
			ArgType:   argType,
			ReplyType: replyType,
		}
		log.Printf("rpc server:register %s.%s success\n", s.name, method.Name)
	}
}

//
// call
// @Description: 通过反射值调用方法
// @receiver s
// @param m
// @param argv
// @param replyv
// @return error
//
func (s *service) call(m *methodType, argv, replyv reflect.Value) error {
	atomic.AddUint64(&m.numCalls, 1)
	f := m.method.Func
	returnValues := f.Call([]reflect.Value{s.rcvr, argv, replyv})
	if errInter := returnValues[0].Interface(); errInter != nil {
		return errInter.(error)
	}
	return nil
}

func isExportedOrBuiltinType(t reflect.Type) bool {
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}
