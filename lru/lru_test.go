/**
* @Author:zhoutao
* @Date:2021/1/3 上午10:45
* @Desc:
 */

package lru

import (
	"reflect"
	"testing"
)

type str string

func (s str) Len() int {
	return len(s)
}

func TestCache_Get(t *testing.T) {
	lru := NewCache(int64(0), nil)
	//str("1234") 将"1234"转换为str类型
	lru.AddCache("key1", str("1234"))

	if ele, ok := lru.Get("key1"); !ok || string(ele.(str)) != "1234" {
		t.Fatal("cache hit key1 = 1234 failed")
	}

	if _, ok := lru.Get("key2"); ok {
		t.Fatal("cache hit hey2 failed")
	}
}

//当使用内存超过了设定值时，是否会触发"无用"节点的移除
func TestCache_RemoveOldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := "value1", "value2", "value3"

	cap := len(k1 + k2 + v1 + v2)

	lru := NewCache(int64(cap), nil)
	lru.AddCache(k1, str(v1))
	lru.AddCache(k2, str(v2))
	lru.AddCache(k3, str(v3))

	//如果key1没有被移除
	if _, ok := lru.Get("key1"); ok || lru.Len() != 2 {
		t.Fatal("remove the oldest key1 failed")
	}
}

//测试回调函数是否能被调用
func TestCache_onEvicted(t *testing.T) {
	keys := make([]string, 0, 4)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}

	lru := NewCache(int64(5), callback)

	lru.AddCache("k1", str("123456786987978"))
	lru.AddCache("k2", str("v2"))
	lru.AddCache("k3", str("v3"))
	lru.AddCache("k4", str("v4"))

	expected := []string{"k1", "k2"}

	if !reflect.DeepEqual(expected, keys) {
		t.Fatalf("call evicted failed! error:%s,expected:%s", keys, expected)
		//call evicted failed! error:[k1 k2 k3],expected:[k1 k2]
	}

}
