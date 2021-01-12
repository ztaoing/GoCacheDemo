/**
* @Author:zhoutao
* @Date:2021/1/12 下午2:49
* @Desc:
 */

package main

import (
	"flag"
	"fmt"
	"github.com/ztaoing/GoCacheDemo"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "3894",
}

func createGroup() *GoCacheDemo.Group {
	return GoCacheDemo.NewGroup("scores", 2<<10, GoCacheDemo.GetterFunc(func(key string) ([]byte, error) {
		log.Println("[DB search] key", key)
		if v, ok := db[key]; ok {
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not exist", key)
	}))
}

func startCacheServer(addr string, addrs []string, cacheDemo *GoCacheDemo.Group) {
	peers := GoCacheDemo.NewHTTPPool(addr)
	peers.Set(addrs...)

	cacheDemo.RegisterPeers(peers)
	log.Println("cacheDemo is running at ", addr)

	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startApiServer(apiAddr string, cacheDemo *GoCacheDemo.Group) {
	http.Handle("/api", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		view, err := cacheDemo.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(view.ByteSlice())
	}))

	log.Println("server is running at ", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	var port int
	var api bool

	flag.IntVar(&port, "port", 8001, "CacheDemo server port")
	flag.BoolVar(&api, "api", false, "Start a api server")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	CacheDemo := createGroup()
	if api {
		go startApiServer(apiAddr, CacheDemo)
	}
	startCacheServer(addrMap[port], []string(addrs), CacheDemo)
}
