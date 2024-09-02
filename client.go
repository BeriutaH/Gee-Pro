package GeeRPC

import (
	"GeeRPC/codec"
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Call 代表一个rpc的开始
type Call struct {
	Seq           uint64     // 序号
	ServiceMethod string     // 格式 "<service>.<method>"
	Args          any        // 函数的参数
	Reply         any        // 函数的返回值
	Error         error      // 如果有错误，设置当前值
	Done          chan *Call // 调用结束时，通知调用方
}

func (c *Call) done() {
	c.Done <- c
}

// Client 代表一个 RPC 客户端，客户端最核心部分
// 一个客户端可能关联有多个未完成的调用，并且一个客户端可能被多个 goroutine 同时使用
type Client struct {
	cc      codec.Codec // 消息的编解码器，和服务端类似，用来序列化将要发送出去的请求，以及反序列化接收到的响应
	opt     *Option
	sending sync.Mutex   // 是一个互斥锁，和服务端类似，为了保证请求的有序发送，即防止出现多个请求报文混淆
	header  codec.Header // 每个请求的消息头，header 只有在请求发送时才需要，而请求发送是互斥的，因此每个客户端只需要一个，声明在 Client 结构体中可以复用
	mu      sync.Mutex
	seq     uint64           // 请求唯一编号
	pending map[uint64]*Call // 存储未处理完的请求，键是编号，值是 Call 实例
	// closing  shutdown 任意一个值置为 true，则表示 Client 处于不可用的状态，但有些许的差别
	closing  bool // 用户主动关闭的，即调用 Close 方法
	shutdown bool // 有错误发生
}

var _ io.Closer = (*Client)(nil)

var ErrShutdown = errors.New("连接已关闭")

func NewClient(conn net.Conn, opt *Option) (*Client, error) {
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		err := fmt.Errorf("无效的编解码器类型 %s", opt.CodecType)
		log.Println("rpc 客户端: 编解码器错误: ", err)
		return nil, err
	}
	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		log.Println("rpc 客户端: 选项错误: ", err)
		_ = conn.Close()
		return nil, err
	}
	return newClientCodec(f(conn), opt), nil
}

func newClientCodec(cc codec.Codec, opt *Option) *Client {
	client := &Client{
		seq:     1, // 1 开始，0 表示无效调用
		cc:      cc,
		opt:     opt,
		pending: make(map[uint64]*Call),
	}
	go client.receive()
	return client
}

// parseOptions 便于用户传入服务端地址，创建 Client 实例
func parseOptions(opts ...*Option) (*Option, error) {
	// 如果 opts 为 nil 或者传递 nil 作为参数
	if len(opts) == 0 || opts[0] == nil {
		return DefaultOption, nil
	}
	if len(opts) != 1 {
		return nil, errors.New("number of options is more than 1")
	}
	opt := opts[0]
	opt.MagicNumber = DefaultOption.MagicNumber
	if opt.CodecType == "" {
		opt.CodecType = DefaultOption.CodecType
	}
	return opt, nil
}

// Dial 连接到指定网络地址的 RPC 服务器
func Dial(network, address string, opts ...*Option) (client *Client, err error) {
	return dialTimeout(NewClient, network, address, opts...)
}

func (c *Client) send(call *Call) {
	// 确保客户端将发送完整的请求
	c.sending.Lock()
	defer c.sending.Unlock()

	// 将call注册进client 中
	seq, err := c.registerCall(call)
	if err != nil {
		call.Error = err
		call.done()
		return
	}

	// 构建请求头
	c.header.ServiceMethod = call.ServiceMethod
	c.header.Seq = seq
	c.header.Error = ""

	// 编码并发送请求
	if err := c.cc.Write(&c.header, call.Args); err != nil {
		// 调用可能为零，这通常意味着 Write 部分失败，客户端已收到响应并处理
		call := c.removeCall(seq)
		if call != nil {
			call.Error = err
			call.done()
		}
	}
}

// Go 异步调用该函数, 它返回代表调用的 Call 结构
func (c *Client) Go(serviceMethod string, args, reply any, done chan *Call) *Call {
	if done == nil {
		done = make(chan *Call, 10)
	} else if cap(done) == 0 {
		log.Panic("rpc 客户端: 完成通道无缓冲")
	}
	call := &Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		Done:          done,
	}
	//log.Printf("call 结构体: %+v", call)
	c.send(call)
	return call
}

// Call 超时处理机制，使用 context 包实现，控制权交给用户，控制更为灵活
// 调用指定函数，等待其完成, 并返回其错误状态
func (c *Client) Call(ctx context.Context, serviceMethod string, args, reply any) error {
	/*
		ctx: 用户可以设置 context.WithTimeout 来自定义超时时间
	*/
	call := c.Go(serviceMethod, args, reply, make(chan *Call, 1))
	//log.Println("服务名: ", serviceMethod, args, reply)
	select {
	case <-ctx.Done(): // 超时将执行这里
		c.removeCall(call.Seq)
		return errors.New("rpc 客户端：调用失败: " + ctx.Err().Error())
	case call = <-call.Done:
		return call.Error
	}

}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closing {
		return ErrShutdown
	}
	c.closing = true
	return c.cc.Close()
}

// IsAvailable 检查是否可用
func (c *Client) IsAvailable() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return !c.shutdown && !c.closing
}

// registerCall 将参数 call 添加到 client.pending 中，并更新 client.seq
func (c *Client) registerCall(call *Call) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	log.Printf("将 call 注册到 client.pending 中，client = %+v", c)
	if c.closing || c.shutdown {
		return 0, ErrShutdown
	}
	log.Println("-----------------------------------------------------------")
	call.Seq = c.seq
	c.pending[call.Seq] = call
	c.seq++
	return call.Seq, nil
}

// terminateCalls 服务端或客户端发生错误时调用, 将 shutdown 设置为 true，且将错误信息通知所有 pending 状态的 call
func (c *Client) terminateCalls(err error) {
	c.sending.Lock()
	defer c.sending.Unlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	log.Println("发生错误shutdown 设为 True")
	c.shutdown = true
	for _, call := range c.pending {
		call.Error = err
		call.done()
	}
}

// removeCall 根据 seq，从 client.pending 中移除对应的 call，并返回
func (c *Client) removeCall(seq uint64) *Call {
	c.mu.Lock()
	defer c.mu.Unlock()
	call := c.pending[seq]
	delete(c.pending, seq)
	return call
}

func (c *Client) receive() {
	/*
		接收到的响应有三种情况
		call不存在
			请求没有发送完整或其他原因取消，但服务端仍然处理了
		call存在
			1. 服务端出错，h.Error 不为空
			2. 服务端处理正常，从body中读取reply的值
	*/
	var err error
	for err == nil {
		var h codec.Header
		if err = c.cc.ReadHeader(&h); err != nil {
			log.Println("读取header信息错误: ", err)
			break
		}
		call := c.removeCall(h.Seq)
		switch {
		case call == nil:
			// 这通常意味着写入部分失败,并且调用已被删除
			log.Println("写入部分失败,并且调用已被删除")
			err = c.cc.ReadBody(nil)
		case h.Error != "":
			log.Println("服务端出错，h.Error 不为空")
			call.Error = fmt.Errorf(h.Error)
			err = c.cc.ReadBody(nil)
			call.done()
		default:
			log.Println("服务端处理正常，从body中读取reply的值")
			err = c.cc.ReadBody(call.Reply)
			if err != nil {
				call.Error = errors.New("reading body " + err.Error())
			}
			call.done()
		}
	}
	log.Println("发生错误，因此终止呼叫待处理的呼叫: ", err)
	// 发生错误，因此终止呼叫待处理的呼叫
	c.terminateCalls(err)
}

type clientResult struct {
	client *Client
	err    error
}

type newClientFunc func(conn net.Conn, opt *Option) (client *Client, err error)

func dialTimeout(f newClientFunc, network, address string, opts ...*Option) (client *Client, err error) {
	opt, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}
	// 将 net.Dial 替换为 net.DialTimeout，如果连接创建超时，将返回错误
	conn, err := net.DialTimeout(network, address, opt.ConnectTimeout)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = conn.Close()
		}
	}()
	ch := make(chan clientResult)
	go func() {
		client, err = f(conn, opt)
		ch <- clientResult{client: client, err: err}
	}()
	if opt.ConnectTimeout == 0 {
		result := <-ch
		return result.client, result.err
	}
	select {
	// time.After 用于在指定的时间段之后发送一个事件
	case <-time.After(opt.ConnectTimeout):
		// 如果 time.After() 信道先接收到消息，则说明 NewClient 执行超时，返回错误
		return nil, fmt.Errorf("rpc 客户端：连接超时：预计在 %s", opt.ConnectTimeout)
	case result := <-ch:
		return result.client, result.err
	}
}

// NewHTTPClient 通过 HTTP 作为传输协议新建一个客户端实例
func NewHTTPClient(conn net.Conn, opt *Option) (*Client, error) {
	/*
		conn: 是一个实现了 io.Reader 接口的连接对象
	*/
	// io.WriteString 向 conn（一个实现了 io.Writer 接口的对象）写入字符串
	reqStr := fmt.Sprintf("CONNECT %s HTTP/1.0\r\n\r\n", defaultRPCPath)
	_, _ = io.WriteString(conn, reqStr)
	// 切换到 RPC 协议之前需要成功的 HTTP 响应，bufio.NewReader 创建一个带缓冲的读取器
	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	if resp == nil {
		log.Println("创建缓冲读取器出错了: ", err)
	}
	if err == nil && resp.Status == connected {
		return NewClient(conn, opt)
	}
	if err == nil {
		err = errors.New("意外的 HTTP 响应: " + resp.Status)
	}
	return nil, err
}

// DialHTTP 连接到指定网络地址的 HTTP RPC 服务器，监听默认的 HTTP RPC 路径
func DialHTTP(network, address string, opts ...*Option) (*Client, error) {
	return dialTimeout(NewHTTPClient, network, address, opts...)
}

// XDial 根据第一个参数 rpcAddr 调用不同的函数来连接 RPC 服务器
func XDial(rpcAddr string, opts ...*Option) (*Client, error) {
	/*
		rpcAddr: 是一种通用格式 (protocol@addr)，用于表示 rpc 服务器
		例如: http@10.0.0.1:7001、tcp@10.0.0.1:9999、unix@/tmp/geerpc.sock
	*/
	log.Println("获取的rpc地址: ", rpcAddr)
	parts := strings.Split(rpcAddr, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("rpc 客户端错误: 格式错误: %s, 正确格式为: protocol@addr ", rpcAddr)
	}
	protocol, addr := parts[0], parts[1]
	switch protocol {
	case "http":
		return DialHTTP("tcp", addr, opts...)
	default:
		// tcp、unix 或其他传输协议
		return Dial(protocol, addr, opts...)
	}
}
