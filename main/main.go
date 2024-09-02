package main

import (
	geerpc "GeeRPC"
	"GeeRPC/xclient"
	"context"
	"log"
	"net"
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

func startServer(addr chan string) {
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
	addr <- l.Addr().String() // 将服务器的地址存进管道中
	server.Accept(l)

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
	println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
	if err != nil {
		log.Printf("%s %s error: %v", typ, serviceMethod, err)
	} else {
		log.Printf("%s %s success: %d + %d = %d", typ, serviceMethod, args.Num1, args.Num2, reply)
	}
}

// call 调用单个实例
func call(addr1, addr2 string) {
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
	d := xclient.NewMultiServerDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
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
func broadcast(addr1, addr2 string) {
	d := xclient.NewMultiServerDiscovery([]string{"tcp@" + addr1, "tcp@" + addr2})
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
	ch1 := make(chan string)
	ch2 := make(chan string)
	go startServer(ch1)
	go startServer(ch2)
	addr1 := <-ch1
	addr2 := <-ch2

	time.Sleep(time.Second)
	call(addr1, addr2)
	broadcast(addr1, addr2)
	//println("-------start--------")
	//go call(ch) // 协程的方式执行call, 等待服务器启动再接收addr来进行调用
	//println("--------end-------")
	//startServer(ch)

}
