/**
* @Author:zhoutao
* @Date:2021/1/3 下午12:42
* @Desc:
 */

package GoCacheDemo

import (
	"fmt"
	"github.com/ztaoing/GoCacheDemo/singleflight"
	"log"
	"sync"
)

// callback 在数据不存在的时候，调用这个函数，得到数据源
// 根据不同的数据源获取数据的实现由用户自行实现

type Getter interface {
	Get(key string) ([]byte, error)
}

// implement Getter
// 接口型函数，方便使用者在调用时技能够传入函数作为参数，也能够传入实现了该接口的结构体作为参数
type GetterFunc func(key string) ([]byte, error)

func (g GetterFunc) Get(key string) ([]byte, error) {
	return g(key)
}

// core struct
type Group struct {
	name      string // 每个group拥有一个唯一的名称：例如有三个group：学生的成绩scores；学生信息info；学生课程courses
	getter    Getter // 缓存未命中时获取源数据的回调方法
	mainCache cache  // 并发缓存
	peers     PeerPicker
	loader    *singleflight.Group
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil getter")
	}
	mu.Lock()
	defer mu.Unlock()

	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}

	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	//使用了只读锁,因为不涉及任何变量的写操作
	mu.RLock()
	g := groups[name]
	mu.RUnlock()

	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[Cache] hit")
		return v, nil
	}
	//缓存不存在
	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {
	// only one request for the key can be do at the same time
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		//分布式场景下，load先从远程节点获取，失败后再回退到getLocally
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[CacheDemo]Failed to get from peer ", err)
			}
		}
		return g.getLocally(key)
	})
	if err != nil {
		return viewi.(ByteView), err
	}
	return

}

//
func (g *Group) getLocally(key string) (ByteView, error) {
	//从remote获取源数据
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{
		b: cloneBytes(bytes),
	}
	//将源数据添加到并发缓存中
	g.putCache(key, value)

	return value, nil
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{
		b: bytes,
	}, nil
}

func (g *Group) putCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeers called more than one time")
	}
	g.peers = peers
}
