package main

import (
	"GeeCache/geecache"
	"flag"
	"fmt"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *geecache.Group {
	return geecache.NewGroup("scores", 2<<10, geecache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("查询key: ", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s 不存在", key)
		}))
}

func startCacheServer(addr string, addrAll []string, gee *geecache.Group) {
	peers := geecache.NewHTTPPool(addr)
	peers.Set(addrAll...)
	gee.RegisterPeer(peers)
	log.Println("geecache 正在运行 ", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, gee *geecache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			key := request.URL.Query().Get("key")
			view, err := gee.Get(key)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
			writer.Header().Set("Content-Type", "application/octet-stream")
			writer.Write(view.ByteSlice()) // ByteView.b
		}))
	log.Println("前端服务器正在运行 ", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	//geecache.NewGroup("scores", 2<<10, geecache.GetterFunc(
	//	func(key string) ([]byte, error) {
	//		log.Println("查询key: ", key)
	//		if v, ok := db[key]; ok {
	//			return []byte(v), nil
	//		}
	//		return nil, fmt.Errorf("%s 不存在", key)
	//	}))
	//addr := "localhost:9999"
	//peers := geecache.NewHTTPPool(addr)
	//log.Println("geecache 正在运行 ", addr)
	//log.Fatal(http.ListenAndServe(addr, peers))
	var port int // 存储服务的端口号
	var api bool // 是否启动API服务
	// 绑定命令行参数到变量
	flag.IntVar(&port, "port", 8001, "GeeCache 服务端口")
	flag.BoolVar(&api, "api", false, "启动 api 服务？")
	// 解析命令行参数
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}
	var addrAll []string
	for _, v := range addrMap {
		addrAll = append(addrAll, v)
	}
	gee := createGroup()
	if api {
		// 启动一个 API 服务（端口 9999），与用户进行交互, 用户感知
		go startAPIServer(apiAddr, gee)
	}
	// 启动缓存服务器: 创建 HTTPPool，添加节点信息，注册到 gee 中，启动 HTTP 服务（共3个端口，8001/8002/8003）,用户不感知
	log.Println("缓存服务端口>>> ", port)
	startCacheServer(addrMap[port], addrAll, gee)

}
