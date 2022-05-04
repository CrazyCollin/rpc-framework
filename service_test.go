package gorpc

import (
	"gorpc/codec"
	"io"
	"net"
	"reflect"
	"sync"
	"testing"
)

func TestAccept(t *testing.T) {
	type args struct {
		lis net.Listener
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Accept(tt.args.lis)
		})
	}
}

func TestCall_done(t *testing.T) {
	type fields struct {
		Seq           uint64
		ServiceMethod string
		Args          interface{}
		Reply         interface{}
		Error         error
		Done          chan *Call
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			call := &Call{
				Seq:           tt.fields.Seq,
				ServiceMethod: tt.fields.ServiceMethod,
				Args:          tt.fields.Args,
				Reply:         tt.fields.Reply,
				Error:         tt.fields.Error,
				Done:          tt.fields.Done,
			}
			call.done()
		})
	}
}

func TestClient_Close(t *testing.T) {
	type fields struct {
		cc       codec.Codec
		opt      *Option
		sending  *sync.Mutex
		header   codec.Header
		mu       *sync.Mutex
		seq      uint64
		pending  map[uint64]*Call
		closing  bool
		shutdown bool
	}
	type args struct {
		call *Call
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				cc:       tt.fields.cc,
				opt:      tt.fields.opt,
				sending:  tt.fields.sending,
				header:   tt.fields.header,
				mu:       tt.fields.mu,
				seq:      tt.fields.seq,
				pending:  tt.fields.pending,
				closing:  tt.fields.closing,
				shutdown: tt.fields.shutdown,
			}
			if err := client.Close(tt.args.call); (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClient_IsAvailable(t *testing.T) {
	type fields struct {
		cc       codec.Codec
		opt      *Option
		sending  *sync.Mutex
		header   codec.Header
		mu       *sync.Mutex
		seq      uint64
		pending  map[uint64]*Call
		closing  bool
		shutdown bool
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				cc:       tt.fields.cc,
				opt:      tt.fields.opt,
				sending:  tt.fields.sending,
				header:   tt.fields.header,
				mu:       tt.fields.mu,
				seq:      tt.fields.seq,
				pending:  tt.fields.pending,
				closing:  tt.fields.closing,
				shutdown: tt.fields.shutdown,
			}
			if got := client.IsAvailable(); got != tt.want {
				t.Errorf("IsAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_receive(t *testing.T) {
	type fields struct {
		cc       codec.Codec
		opt      *Option
		sending  *sync.Mutex
		header   codec.Header
		mu       *sync.Mutex
		seq      uint64
		pending  map[uint64]*Call
		closing  bool
		shutdown bool
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := Client{
				cc:       tt.fields.cc,
				opt:      tt.fields.opt,
				sending:  tt.fields.sending,
				header:   tt.fields.header,
				mu:       tt.fields.mu,
				seq:      tt.fields.seq,
				pending:  tt.fields.pending,
				closing:  tt.fields.closing,
				shutdown: tt.fields.shutdown,
			}
			client.receive()
		})
	}
}

func TestClient_registerCall(t *testing.T) {
	type fields struct {
		cc       codec.Codec
		opt      *Option
		sending  *sync.Mutex
		header   codec.Header
		mu       *sync.Mutex
		seq      uint64
		pending  map[uint64]*Call
		closing  bool
		shutdown bool
	}
	type args struct {
		call *Call
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    uint64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := Client{
				cc:       tt.fields.cc,
				opt:      tt.fields.opt,
				sending:  tt.fields.sending,
				header:   tt.fields.header,
				mu:       tt.fields.mu,
				seq:      tt.fields.seq,
				pending:  tt.fields.pending,
				closing:  tt.fields.closing,
				shutdown: tt.fields.shutdown,
			}
			got, err := client.registerCall(tt.args.call)
			if (err != nil) != tt.wantErr {
				t.Errorf("registerCall() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("registerCall() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_removeCall(t *testing.T) {
	type fields struct {
		cc       codec.Codec
		opt      *Option
		sending  *sync.Mutex
		header   codec.Header
		mu       *sync.Mutex
		seq      uint64
		pending  map[uint64]*Call
		closing  bool
		shutdown bool
	}
	type args struct {
		seq uint64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Call
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &Client{
				cc:       tt.fields.cc,
				opt:      tt.fields.opt,
				sending:  tt.fields.sending,
				header:   tt.fields.header,
				mu:       tt.fields.mu,
				seq:      tt.fields.seq,
				pending:  tt.fields.pending,
				closing:  tt.fields.closing,
				shutdown: tt.fields.shutdown,
			}
			if got := client.removeCall(tt.args.seq); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("removeCall() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_terminateCalls(t *testing.T) {
	type fields struct {
		cc       codec.Codec
		opt      *Option
		sending  *sync.Mutex
		header   codec.Header
		mu       *sync.Mutex
		seq      uint64
		pending  map[uint64]*Call
		closing  bool
		shutdown bool
	}
	type args struct {
		err error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := Client{
				cc:       tt.fields.cc,
				opt:      tt.fields.opt,
				sending:  tt.fields.sending,
				header:   tt.fields.header,
				mu:       tt.fields.mu,
				seq:      tt.fields.seq,
				pending:  tt.fields.pending,
				closing:  tt.fields.closing,
				shutdown: tt.fields.shutdown,
			}
			client.terminateCalls(tt.args.err)
		})
	}
}

func TestNewServer(t *testing.T) {
	tests := []struct {
		name string
		want *Server
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewServer(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_Accept(t *testing.T) {
	type args struct {
		lis net.Listener
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &Server{}
			server.Accept(tt.args.lis)
		})
	}
}

func TestServer_ServeConn(t *testing.T) {
	type args struct {
		conn io.ReadWriteCloser
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &Server{}
			server.ServeConn(tt.args.conn)
		})
	}
}

func TestServer_handleRequest(t *testing.T) {
	type args struct {
		cc      codec.Codec
		req     *request
		sending *sync.Mutex
		wg      *sync.WaitGroup
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &Server{}
			server.handleRequest(tt.args.cc, tt.args.req, tt.args.sending, tt.args.wg)
		})
	}
}

func TestServer_readRequest(t *testing.T) {
	type args struct {
		cc codec.Codec
	}
	tests := []struct {
		name    string
		args    args
		want    *request
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &Server{}
			got, err := server.readRequest(tt.args.cc)
			if (err != nil) != tt.wantErr {
				t.Errorf("readRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readRequest() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_readRequestHeader(t *testing.T) {
	type args struct {
		cc codec.Codec
	}
	tests := []struct {
		name    string
		args    args
		want    *codec.Header
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &Server{}
			got, err := server.readRequestHeader(tt.args.cc)
			if (err != nil) != tt.wantErr {
				t.Errorf("readRequestHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readRequestHeader() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestServer_sendResponse(t *testing.T) {
	type args struct {
		cc      codec.Codec
		h       *codec.Header
		body    interface{}
		sending *sync.Mutex
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &Server{}
			server.sendResponse(tt.args.cc, tt.args.h, tt.args.body, tt.args.sending)
		})
	}
}

func TestServer_serveCodec(t *testing.T) {
	type args struct {
		cc codec.Codec
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &Server{}
			server.serveCodec(tt.args.cc)
		})
	}
}
