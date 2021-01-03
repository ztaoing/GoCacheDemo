/**
* @Author:zhoutao
* @Date:2021/1/3 下午12:50
* @Desc:
 */

package GoCacheDemo

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

func TestGetter(t *testing.T) {
	//记住GetterFunc的类型转换，将一个匿名回调函数转换为了接口 f Getter
	//调用该接口的方法f.Get(key string),实际上技术在调用匿名回调函数
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})
	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Errorf("callback failed")
	}
}

var db = map[string]string{
	"tom":  "567",
	"json": "890",
	"sam":  "345",
}

// 缓存存在的情况下，直接从缓存中获取，使用loadCounts 统计某个键调用回调函数的次数，如果次数大于1，则表示调用了多次回调函数，没有缓存
// 缓存不存在的情况下，能够通过回调获取源数据
func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(db))
	group := NewGroup("scores", 2<<10, GetterFunc(func(key string) ([]byte, error) {
		log.Println("[db] search key", key)
		//从数据库加载
		if v, ok := db[key]; ok {
			if _, ok = loadCounts[key]; !ok {
				loadCounts[key] = 0
			}
			loadCounts[key] += 1
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not exist in db", key)
	}))

	for k, v := range db {
		//在并发缓存中没有找到或者在缓存中与数据库中的不一致
		if view, err := group.Get(k); err != nil || view.String() != v {
			t.Fatalf("failed to get value from tom")
		} // load from db with callback function

		if _, err := group.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		} // hit cache
	}

	if view, err := group.Get("unknown"); err == nil {
		t.Fatalf("the value of unknown is empty:%d", view)
	}
}
