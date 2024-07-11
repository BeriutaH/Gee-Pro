package GeeRPC

import (
	"GeeRPC/codec"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"sync"
)

// 消息的编解码方式, 将这部分信息，放到结构体 Option 中承载
// 报文将以这样的形式发送
//| Option{MagicNumber: xxx, CodecType: xxx} | Header{ServiceMethod ...} | Body interface{} |
//| <------      固定 JSON 编码      ------>  | <-------   编码方式由 CodeType 决定   ------->  |

const MagicNumber = 0x3bef5c

type Option struct {
	MagicNumber int
	CodecType   codec.CType
}

var DefaultOption = &Option{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}
var invalidRequest = struct{}{}

// request 请求存储调用的所有信息
type request struct {
	h            *codec.Header // 请求头
	argV, replyV reflect.Value // 请求的参数和答复 是 reflect.Value 类型
}

type Server struct{}

func NewServer() *Server {
	return &Server{}
}

// DefaultServer 默认 Server 实例
var DefaultServer = NewServer()

func (server *Server) Accept(lis net.Listener) {
	// lis 代表一个网络监听器
	// for 循环等待 socket 连接建立
	for {
		// Accept 用于等待并返回下一个连接到该监听器的客户端
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc 服务器: 接收数据错误: ", err)
			return
		}
		// 开启子协程处理，处理过程交给了 ServerConn 方法
		go server.ServeConn(conn)
	}
}

// ServeConn 在单个连接上运行服务器, 阻塞，为连接提供服务，直到客户端挂断
func (server *Server) ServeConn(conn io.ReadWriteCloser) {
	defer conn.Close()
	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc 服务器: 选项错误: ", err)
		return
	}
	if opt.MagicNumber != MagicNumber {
		log.Println("rpc 服务器: 无效的数字: ", opt.MagicNumber)
		return
	}
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Println("rpc 服务器: 编解码器类型: ", opt.CodecType)
		return
	}
	server.serveCodec(f(conn))
}

func (server *Server) serveCodec(cc codec.Codec) {
	// TODO 处理请求是并发的，但是回复请求的报文必须是逐个发送的，并发容易导致多个回复报文交织在一起，客户端无法解析，使用互斥锁(sending)保证
	sending := new(sync.Mutex)
	// 创建一个等待组，用于等待所有请求处理完毕后再关闭连接
	wg := new(sync.WaitGroup)
	// 在一次连接中，允许接收多个请求
	for {
		// // 读取请求
		req, err := server.readRequest(cc)
		if err != nil {
			// 如果请求为nil，意味着没有更多的请求，退出循环
			if req == nil {
				break
			}
			// 设置错误信息
			req.h.Error = err.Error()
			// 发送错误响应
			server.sendResponse(cc, req.h, invalidRequest, sending)
			continue
		}
		// 增加等待组的计数
		wg.Add(1)
		// 异步处理请求
		go server.handleRequest(cc, req, sending, wg)
	}
	// 等待所有请求处理完毕
	wg.Wait()
	cc.Close()

}

// readRequest 读取请求
func (server *Server) readRequest(cc codec.Codec) (*request, error) {
	h, err := server.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}
	req := &request{h: h}
	req.argV = reflect.New(reflect.TypeOf(""))
	if err = cc.ReadBody(req.argV.Interface()); err != nil {
		log.Println("rpc 服务器: 读取 argv 错误: ", err)
	}
	return req, nil
}

// sendResponse 回复请求
func (server *Server) sendResponse(cc codec.Codec, h *codec.Header, body any, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := cc.Write(h, body); err != nil {
		log.Println("rpc 服务器：写入响应错误: ", err)
	}
}

// handleRequest 处理请求
func (server *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()                     // 这相当于 wg.Add(-1)，将计数器减 1
	log.Println(req.h, req.argV.Elem()) // req.argv.Elem() 用于获取指针、接口或其他复合类型的底层值
	// TODO 应调用已注册的 rpc 方法来获取正确的 replyv
	req.replyV = reflect.ValueOf(fmt.Sprintf("geerpc resp %d", req.h.Seq))
	server.sendResponse(cc, req.h, req.replyV.Interface(), sending)
}

func (server *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && !errors.Is(err, io.ErrUnexpectedEOF) {
			log.Println("rpc 服务器: 读取标头错误: ", err)
		}
		return nil, err
	}
	return &h, nil
}

func Accept(lis net.Listener) { DefaultServer.Accept(lis) }
