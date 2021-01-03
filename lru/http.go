/**
* @Author:zhoutao
* @Date:2021/1/3 下午2:30
* @Desc:
 */

package lru

import (
	"fmt"
	"github.com/ztaoing/GoCacheDemo"
	"log"
	"net/http"
	"strings"
)

// 分布式缓存需要实现节点之间的通信，建立基本的HTTP的通信机制是比较常见和简单的做法。
// 如果一个节点启动了http服务，那么这个节点就可以被其他节点访问

const defaultBathPaht = "/gocache/"

type HTTPPool struct {
	self     string // record the address it`s self ,including host name and port
	basePath string // http.example.com/gocache/
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBathPaht,
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

	group := GoCacheDemo.GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group:"+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		// todo:为什么是服务端错误
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}
