package main

import (
	geerpc "GeeRPC"
	"GeeRPC/codec"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

func startServer(addr chan string) {
	// 信道 addr，确保服务端端口监听成功，客户端再发起请求
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("网络错误: ", err)
	}
	log.Println("rpc 服务开始启动: ", l.Addr())
	addr <- l.Addr().String()
	geerpc.Accept(l)
}

func main() {
	addr := make(chan string)
	go startServer(addr)

	conn, _ := net.Dial("tcp", <-addr)
	defer conn.Close()
	time.Sleep(time.Second)
	// 把 geerpc.DefaultOption 结构体的数据变成json数据，并写入conn中
	_ = json.NewEncoder(conn).Encode(geerpc.DefaultOption)
	cc := codec.NewGobCodec(conn)
	for i := 0; i < 5; i++ {
		// 发送消息头
		h := &codec.Header{
			ServiceMethod: "FooSum",
			Seq:           uint64(i),
		}
		_ = cc.Write(h, fmt.Sprintf("geerpc req %d", h.Seq))
		_ = cc.ReadHeader(h)
		var reply string
		_ = cc.ReadBody(&reply)
		log.Println("响应数据: ", reply)
	}
}
