/**
* @Author:zhoutao
* @Date:2021/1/3 下午4:11
* @Desc:
 */

package GoCacheDemo

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

// 根据传入的key选择相应节点PeerGetter
type PeerPicker interface {
	PickPeer(key string) (PeerGetter, bool)
}

// 从对应的group查找缓存之
type PeerGetter interface {
	Get(group, key string) ([]byte, error)
}

type HttpGetter struct {
	baseUrl string
}

func (h *HttpGetter) Get(group, key string) ([]byte, error) {
	// remote url
	u := fmt.Sprintf("%v,%v/%v", h.baseUrl, url.QueryEscape(group), url.QueryEscape(key))
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server return:%v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body:%v", err)
	}

	return bytes, nil
}
