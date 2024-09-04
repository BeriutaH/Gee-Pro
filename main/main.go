package main

import (
	geerpc "GeeRPC"
	"GeeRPC/registry"
	"GeeRPC/xclient"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type Foo int

type Args struct{ Num1, Num2 int }

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func (f Foo) Sleep(args Args, reply *int) error {
	time.Sleep(time.Second * time.Duration(args.Num1))
	*reply = args.Num1 + args.Num2
	return nil
}

func startServer(addr string, wg *sync.WaitGroup) {
	var foo Foo
	//if err := geerpc.Register(&foo); err != nil {
	//	log.Println("register error:", err)
	//}
	//// pick a free port
	//l, err := net.Listen("tcp", ":9999")
	//if err != nil {
	//	log.Fatal("network error:", err)
	//}
	//log.Println("start rpc server on", l.Addr())
	//geerpc.HandleHTTP()
	//addr <- l.Addr().String() // 将服务器的地址存进管道中
	//err1 := http.Serve(l, nil)
	//if err1 != nil {
	//	log.Println("服务端错误: ", err1)
	//}
	l, _ := net.Listen("tcp", ":0")
	server := geerpc.NewServer()
	_ = server.Register(&foo)
	//addr <- l.Addr().String() // 将服务器的地址存进管道中
	// 添加函数 startRegistry，稍微修改 startServer，添加调用注册中心的 Heartbeat 方法的逻辑，定期向注册中心发送心跳保活
	registry.Heartbeat(addr, "tcp@"+l.Addr().String(), 0)
	wg.Done()
	server.Accept(l)

}

func startRegistry(wg *sync.WaitGroup) {
	l, _ := net.Listen("tcp", ":9999")
	registry.HandleHTTP()
	wg.Done()
	_ = http.Serve(l, nil)
}

func foo(xc *xclient.XClient, ctx context.Context, typ, serviceMethod string, args *Args) {
	var reply int
	var err error
	switch typ {
	case "call":
		err = xc.Call(ctx, serviceMethod, args, &reply)
	case "broadcast":
		err = xc.Broadcast(ctx, serviceMethod, args, &reply)
	}
	if err != nil {
		log.Printf("%s %s error: %v", typ, serviceMethod, err)
	} else {
		log.Printf("%s %s success: %d + %d = %d", typ, serviceMethod, args.Num1, args.Num2, reply)
	}
}

// call 调用单个实例
func call(registry string) {
	//// addrCh 是一个无缓冲的通道，必须等到startServer(addr)执行完成，即服务器启动了才有值（服务器的地址）
	//client, _ := geerpc.DialHTTP("tcp", <-addrCh)
	//log.Println("创建连接 TCP")
	//defer client.Close()
	//time.Sleep(time.Second)
	//var wg sync.WaitGroup
	//for i := 0; i < 5; i++ {
	//	wg.Add(1)
	//	go func(i int) {
	//		defer wg.Done()
	//		args := &Args{Num1: i, Num2: i * i}
	//		var reply int
	//		//client.Call(context.Background(), "Foo.Sum", args, &reply)
	//		if err := client.Call(context.Background(), "Foo.Sum", args, &reply); err != nil {
	//			log.Println("调用 Foo.Sum 错误: ", err)
	//		}
	//		log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
	//	}(i)
	//}
	//wg.Wait()
	d := xclient.NewGeeRegistryDiscovery(registry, 0)
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer xc.Close()
	// 发送请求并接收响应
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			foo(xc, context.Background(), "call", "Foo.Sum", &Args{Num1: i, Num2: i * i})
		}(i)
	}
	wg.Wait()
}

// broadcast 调用所有服务实例
func broadcast(registry string) {
	d := xclient.NewGeeRegistryDiscovery(registry, 0)
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer xc.Close()
	// 发送请求并接收响应
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			foo(xc, context.Background(), "broadcast", "Foo.Sum", &Args{Num1: i, Num2: i * i})
			// 预计 2 - 5 次超时
			ctx, _ := context.WithTimeout(context.Background(), time.Second*2)
			foo(xc, ctx, "broadcast", "Foo.Sleep", &Args{Num1: i, Num2: i * i})
		}(i)
	}
	wg.Wait()
}

func main() {
	log.SetFlags(0)
	registryAddr := "http://localhost:9999/_geerpc_/registry"
	var wg sync.WaitGroup
	wg.Add(1)
	go startRegistry(&wg)
	wg.Wait()

	time.Sleep(time.Second) // 确保注册中心启动后，再启动 RPC 服务端
	wg.Add(2)
	go startServer(registryAddr, &wg)
	go startServer(registryAddr, &wg)
	wg.Wait()

	time.Sleep(time.Second)
	call(registryAddr)
	broadcast(registryAddr)
	fmt.Println("测试")
}
