package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

type GobCodec struct {
	conn io.ReadWriteCloser
	//防阻塞创建带缓冲的Writer
	buf *bufio.Writer
	dec *gob.Decoder
	enc *gob.Encoder
}

func (g *GobCodec) Close() error {
	return g.conn.Close()
}

func (g *GobCodec) ReadHeader(header *Header) error {
	err := g.dec.Decode(header)
	//if err != nil {
	//	log.Println("codec decode header error:", err)
	//}
	return err
}

func (g *GobCodec) ReadBody(body interface{}) error {
	return g.dec.Decode(body)
}

func (g *GobCodec) Write(header *Header, body interface{}) (err error) {
	defer func() {
		_ = g.buf.Flush()
		if err != nil {
			_ = g.Close()
		}
	}()
	//编码消息头
	if err := g.enc.Encode(header); err != nil {
		log.Println("rpc codec:gob error encoding header:", err)
		return err
	}
	//编码消息体
	if err := g.enc.Encode(body); err != nil {
		log.Println("rpc codec:gob error encoding body:", err)
		return err
	}
	return nil
}

func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),
		enc:  gob.NewEncoder(buf),
	}
}
