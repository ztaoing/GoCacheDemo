/**
* @Author:zhoutao
* @Date:2021/1/3 下午3:43
* @Desc:
 */

package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// 虚拟节点扩充了节点的数量，解决了节点较少的情况下数据容易倾斜的问题。而且代价非常小，
// 只需要正价个字典map维护真实节点和虚拟节点之间的映射关系

type Hash func(data []byte) uint32

type Map struct {
	hash     Hash
	replicas int            //虚拟节点倍数
	keys     []int          //哈希环
	hashMap  map[int]string //虚拟节点与真实节点的映射
}

func NewMap(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	//set default hash
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// 添加真实节点
func (m *Map) AddMap(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			// 加入到哈希环中
			m.keys = append(m.keys, hash)
			// 虚拟节点与真实节点的映射
			m.hashMap[hash] = key
		}
	}
	//环上的hash值排序
	sort.Ints(m.keys)
}

// 通过虚拟节点选择真实节点
func (m *Map) Get(key string) string {
	if len(key) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	//顺时针找打第一个匹配的虚拟节点的下标
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	//通过虚拟节点找到真实节点
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
