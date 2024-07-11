package codec

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
)

type GobCodec struct {
	conn io.ReadWriteCloser // 是由构建函数传入, 通过 TCP 或者 Unix 建立 socket 时得到的链接实例
	buf  *bufio.Writer      // 防止阻塞而创建的带缓冲的 Writer, 提升性能
	dec  *gob.Decoder
	enc  *gob.Encoder
}

func (g *GobCodec) Close() error {
	return g.conn.Close()
}

func (g *GobCodec) ReadHeader(header *Header) error {
	return g.dec.Decode(header)
}

func (g *GobCodec) ReadBody(body any) error {
	return g.dec.Decode(body)
}

func (g *GobCodec) Write(header *Header, body any) (err error) {
	defer func() {
		_ = g.buf.Flush()
		if err != nil {
			_ = g.Close()
		}
	}()
	if err = g.enc.Encode(header); err != nil {
		log.Println("rpc 编解码器: gob 编码标头错误: ", err)
		return err
	}
	if err = g.enc.Encode(body); err != nil {
		log.Println("rpc 编解码器: gob 编码主体错误: ", err)
		return err
	}
	return nil
}

var _ Codec = (*GobCodec)(nil)

func NewGobCodec(conn io.ReadWriteCloser) Codec {
	buf := bufio.NewWriter(conn)
	return &GobCodec{
		conn: conn,
		buf:  buf,
		dec:  gob.NewDecoder(conn),
		enc:  gob.NewEncoder(buf),
	}
}
