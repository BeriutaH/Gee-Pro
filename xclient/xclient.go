package xclient

import (
	geerpc "GeeRPC"
	"context"
	"io"
	"reflect"
	"sync"
)

type XClient struct {
	d       Discovery
	mode    SelectMode
	opt     *geerpc.Option
	mu      sync.Mutex
	clients map[string]*geerpc.Client
}

func NewXClient(d Discovery, mode SelectMode, opt *geerpc.Option) *XClient {
	return &XClient{d: d, mode: mode, opt: opt, clients: make(map[string]*geerpc.Client)}
}

func (xc *XClient) Close() error {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	for key, client := range xc.clients {
		_ = client.Close() // 忽略错误
		delete(xc.clients, key)
	}
	return nil
}

func (xc *XClient) dial(rpcAddr string) (*geerpc.Client, error) {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	client, ok := xc.clients[rpcAddr]
	// 检查 xc.clients 是否有缓存的 Client，如果有，检查是否是可用状态，如果是则返回缓存的 Client
	if ok && !client.IsAvailable() {
		// 如果不可用，则从缓存中删除
		_ = client.Close()
		delete(xc.clients, rpcAddr)
		client = nil
	}
	// 如果步骤 1) 没有返回缓存的 Client，则说明需要创建新的 Client，缓存并返回
	if client == nil {
		var err error
		client, err = geerpc.XDial(rpcAddr, xc.opt)
		if err != nil {
			return nil, err
		}
		xc.clients[rpcAddr] = client
	}
	return client, nil
}

func (xc *XClient) call(rpcAddr string, ctx context.Context, serviceMethod string, args, reply any) error {
	client, err := xc.dial(rpcAddr)
	if err != nil {
		return err
	}
	return client.Call(ctx, serviceMethod, args, reply)
}

// Call  调用命名函数，等待其完成， 并返回其错误状态。xc 将选择合适的服务器。
func (xc *XClient) Call(ctx context.Context, serviceMethod string, args, reply any) error {
	rpcAddr, err := xc.d.Get(xc.mode)
	if err != nil {
		return err
	}
	return xc.call(rpcAddr, ctx, serviceMethod, args, reply)
}

// Broadcast 为每个在发现中注册的服务器调用命名函数
func (xc *XClient) Broadcast(ctx context.Context, serviceMethod string, args, reply any) error {
	/*
		将请求广播到所有的服务实例，如果任意一个实例发生错误，则返回其中一个错误；如果调用成功，则返回其中一个的结果
			1. 为了提升性能，请求是并发的。
			2. 并发情况下需要使用互斥锁保证 error 和 reply 能被正确赋值。
			3. 借助 context.WithCancel 确保有错误发生时，快速失败。
	*/
	servers, err := xc.d.GetAll()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	var mu sync.Mutex // 保护并回复完成
	var e error
	replyDone := reply == nil //如果回复为零，则不需要设置值
	ctx, cancel := context.WithCancel(ctx)
	defer cancel() // 确保在函数结束时调用 cancel
	for _, rpcAddr := range servers {
		wg.Add(1)
		go func(rpcAddr string) {
			defer wg.Done()
			var clonedReply any
			if reply != nil {
				clonedReply = reflect.New(reflect.ValueOf(reply).Elem().Type()).Interface()
			}
			err = xc.call(rpcAddr, ctx, serviceMethod, args, clonedReply)
			mu.Lock()
			defer mu.Unlock()
			if err != nil && e == nil {
				e = err
				cancel() // 如果任何呼叫失败，则取消未完成的呼叫
			}

			if err == nil && !replyDone {
				reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(clonedReply).Elem())
				replyDone = true
			}

		}(rpcAddr)
	}
	wg.Wait()
	return e
}

var _ io.Closer = (*XClient)(nil)
