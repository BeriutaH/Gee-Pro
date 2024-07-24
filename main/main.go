package main

import (
	geerpc "GeeRPC"
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

func startServer(addr chan string) {
	//// 信道 addr，确保服务端端口监听成功，客户端再发起请求
	//// 创建一个 TCP 监听器，监听随机可用端口（":0"）
	//l, err := net.Listen("tcp", ":0")
	//if err != nil {
	//	log.Fatal("网络错误: ", err)
	//}
	//log.Println("rpc 服务开始启动: ", l.Addr())
	//// 将监听器的地址发送到信道 addr，通知客户端服务已经启动
	//addr <- l.Addr().String()
	//geerpc.Accept(l)
	var foo Foo
	if err := geerpc.Register(&foo); err != nil {
		log.Println("register error:", err)
	}
	// pick a free port
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("network error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	addr <- l.Addr().String()
	geerpc.Accept(l)
}

func main() {
	log.SetFlags(0)
	addr := make(chan string)
	go startServer(addr)

	client, _ := geerpc.Dial("tcp", <-addr)
	defer client.Close()
	time.Sleep(time.Second)
	log.Printf("第一次获取的client: %+v", client)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			//args := fmt.Sprintf("geerpc req %d", i)
			//var reply string
			//if err := client.Call("Foo.Sum", args, &reply); err != nil {
			//	log.Fatal("call Foo.Sum error: ", err)
			//}
			args := &Args{Num1: i, Num2: i * i}
			var reply int
			if err := client.Call("Foo.Sum", args, &reply); err != nil {
				log.Println("call Foo.Sum error: ", err)
			}
			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
		}(i)
	}
	wg.Wait()

}
