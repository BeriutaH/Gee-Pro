package codec

import "io"

// 消息编解码

type Header struct {
	ServiceMethod string // 服务名和方法名, 与 Go 语言中的结构体和方法相映射
	Seq           uint64 // 请求的序号或某个请求的 ID，用来区分不同的请求
	Error         string // 错误信息，客户端置为空
}

// Codec 进行编解码的接口
type Codec interface {
	io.Closer // 嵌入式接口 Close() error
	ReadHeader(header *Header) error
	ReadBody(body any) error
	Write(header *Header, body any) error
}

// NewCodecFunc Codec的构造函数
type NewCodecFunc func(conn io.ReadWriteCloser) Codec

type CType string

const (
	GobType  CType = "application/gob"
	JsonType CType = "application/json"
)

var NewCodecFuncMap map[CType]NewCodecFunc

func init() {
	NewCodecFuncMap = make(map[CType]NewCodecFunc)
	NewCodecFuncMap[GobType] = NewGobCodec
}
