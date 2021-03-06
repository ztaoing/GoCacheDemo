/**
* @Author:zhoutao
* @Date:2021/1/3 下午2:30
* @Desc:
 */

package GoCacheDemo

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/ztaoing/GoCacheDemo/consistenthash"
	"github.com/ztaoing/GoCacheDemo/pb"
	"log"
	"net/http"
	"strings"
	"sync"
)

// 分布式缓存需要实现节点之间的通信，建立基本的HTTP的通信机制是比较常见和简单的做法。
// 如果一个节点启动了http服务，那么这个节点就可以被其他节点访问

const (
	defaultBathPath = "/gocache/"
	defaultReplicas = 50
)

type HTTPPool struct {
	self        string // record the address it`s self ,including host name and port
	basePath    string // http.example.com/gocache/
	mu          sync.Mutex
	peers       *consistenthash.Map
	httpGetters map[string]*HttpGetter //映射远程节点对应的HttpGetter。每一个远程节点对应一个HttpGetter
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBathPath,
	}
}

func (h *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", h.self, fmt.Sprintf(format, v...))
}

func (h *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, h.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}

	h.Log("%s %s", r.Method, r.URL.Path)
	// <basepath>/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path[len(h.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group:"+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	//write the value to proto buffer
	body, err := proto.Marshal(&pb.Response{Value: view.ByteSlice()})
	if err != nil {
		// todo:为什么是服务端错误
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(body)
}

func (h *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if peer := h.peers.Get(key); peer != "" && peer != h.self {
		h.Log("Pick peer %s", peer)
		return h.httpGetters[peer], true
	}
	return nil, false
}

//实例化一个一致性hash算法，并且添加了传入的节点，并为每一个节点创建一个http客户端httpGetter
func (h *HTTPPool) Set(peers ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.peers = consistenthash.NewMap(defaultReplicas, nil)
	h.peers.AddMap(peers...)
	h.httpGetters = make(map[string]*HttpGetter, len(peers))
	for _, peer := range peers {
		h.httpGetters[peer] = &HttpGetter{
			baseUrl: peer + h.basePath,
		}
	}
}
