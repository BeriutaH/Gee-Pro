package GeeRPC

import (
	"GeeRPC/codec"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"io"
	"log"
	"net"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"
)

// 消息的编解码方式, 将这部分信息，放到结构体 Option 中承载
// 报文将以这样的形式发送
//| Option{MagicNumber: xxx, CodecType: xxx} | Header{ServiceMethod ...} | Body interface{} |
//| <------      固定 JSON 编码      ------>  | <-------   编码方式由 CodeType 决定   ------->  |

const (
	MagicNumber      = 0x3bef5c
	connected        = "200 Connected to Gee RPC"
	defaultRPCPath   = "/_geeprc_"
	defaultDebugPath = "/debug/geerpc"
)

type Option struct {
	MagicNumber    int           // 标记这是一个 geerpc 请求
	CodecType      codec.CType   // 客户端可以选择不同的 Codec 来编码主体
	ConnectTimeout time.Duration // 默认值为 10s
	HandleTimeout  time.Duration // 默认值为 0，即不设限
}

var DefaultOption = &Option{
	MagicNumber:    MagicNumber,
	CodecType:      codec.GobType,
	ConnectTimeout: time.Second * 10,
}
var invalidRequest = struct{}{}

// request 请求存储调用的所有信息
type request struct {
	h            *codec.Header // 请求头
	argV, replyV reflect.Value // 请求的参数和答复 是 reflect.Value 类型
	mType        *methodType
	svc          *service
}

// isExportedOrBuiltinType 是导出(大写字母开头)还是内置类型
func isExportedOrBuiltinType(t reflect.Type) bool {
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}

type Server struct {
	serviceMap sync.Map
}

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
	server.serveCodec(f(conn), &opt)
}

func (server *Server) serveCodec(cc codec.Codec, opt *Option) {
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
		go server.handleRequest(cc, req, sending, wg, opt.HandleTimeout)
		//go server.handleRequest(cc, req, sending, wg)
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
	req.svc, req.mType, err = server.findService(h.ServiceMethod)
	if err != nil {
		return req, nil
	}
	req.argV = req.mType.newArgv()
	req.replyV = req.mType.newReplyV()

	// 确保 argvi 是一个指针，ReadBody 需要一个指针作为参数
	argvi := req.argV.Interface()
	// 即通过 newArgv() 和 newReplyv() 两个方法创建出两个入参实例，然后通过 cc.ReadBody() 将请求报文反序列化为第一个入参 argv，
	// 在这里同样需要注意 argv 可能是值类型，也可能是指针类型，所以处理方式有点差异
	if req.argV.Type().Kind() != reflect.Ptr {
		argvi = req.argV.Addr().Interface()
	}

	if err = cc.ReadBody(argvi); err != nil {
		log.Println("rpc 服务器: 读取 argv 错误: ", err)
		return req, err
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

// handleRequest 处理请求 确保 sendResponse 仅调用一次
func (server *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup, timeout time.Duration) {
	defer wg.Done() // 这相当于 wg.Add(-1)，将计数器减 1
	called := make(chan struct{})
	sent := make(chan struct{})
	go func() {
		err := req.svc.call(req.mType, req.argV, req.replyV)
		called <- struct{}{}
		if err != nil {
			req.h.Error = err.Error()
			server.sendResponse(cc, req.h, invalidRequest, sending)
			sent <- struct{}{}
			return
		}
		//called 信道接收到消息，代表处理没有超时，继续执行 sendResponse
		server.sendResponse(cc, req.h, req.replyV.Interface(), sending)
		sent <- struct{}{}
	}()
	if timeout == 0 {
		<-called
		<-sent
		return
	}
	select {
	//  time.After(timeout 先于 called 接收到消息，说明处理已经超时，called 和 sent 都将被阻塞
	// 在 case <-time.After(timeout) 处调用 sendResponse
	case <-time.After(timeout):
		req.h.Error = fmt.Sprintf("rpc 服务器: 请求处理超时: 预计在 %s 内", timeout)
		server.sendResponse(cc, req.h, invalidRequest, sending)
	case <-called:
		<-sent
	}
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

// Register 在服务器中发布一组方法
func (server *Server) Register(rcvr any) error {
	s := newService(rcvr)
	if _, dup := server.serviceMap.LoadOrStore(s.name, s); dup {
		return errors.New("rpc服务: " + s.name + " 已被定义")
	}
	return nil
}

// Register 在 DefaultServer 中发布接收器的方法。
func Register(rcvr any) error {
	return DefaultServer.Register(rcvr)
}

// findService 通过 ServiceMethod 从 serviceMap 中找到对应的 service
func (server *Server) findService(serviceMethod string) (svc *service, mType *methodType, err error) {
	dot := strings.LastIndex(serviceMethod, ".") // 返回最后一个点所在的索引
	if dot < 0 {
		err = errors.New("rpc 服务器: 服务/方法请求格式错误 " + serviceMethod)
		return
	}
	serviceName, methodName := serviceMethod[:dot], serviceMethod[dot+1:]
	svci, ok := server.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("rpc 服务器: 找不到服务 " + serviceName)
	}
	svc = svci.(*service)
	mType = svc.method[methodName]
	if mType == nil {
		err = errors.New("rpc 服务器: 找不到方法" + methodName)
	}
	return
}

// ServeHTTP 实现一个 http.Handler 来响应 RPC 请求
func (server *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "CONNECT" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = io.WriteString(w, "405 必须连接\n")
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

// HandleHTTP 在 rpcPath 上注册一个用于 RPC 消息的 HTTP 处理程序
// 仍然需要调用 http.Serve()，通常在 go 语句中
func (server *Server) HandleHTTP() {
	// 两个参数
	http.Handle(defaultRPCPath, server) // 一个是pattern: /_geeprc_, 一个是handler，实现ServeHTTP方法的接口Handler类型
	http.Handle(defaultDebugPath, debugHTTP{server})
	log.Println("rpc服务器调试路径: ", defaultDebugPath)
}

// HandleHTTP 是默认服务器注册 HTTP 处理程序的一种便捷方式
func HandleHTTP() {
	DefaultServer.HandleHTTP()
}
func Accept(lis net.Listener) { DefaultServer.Accept(lis) }
