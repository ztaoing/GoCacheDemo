/**
* @Author:zhoutao
* @Date:2021/1/3 下午4:11
* @Desc:
 */

package GoCacheDemo

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/ztaoing/GoCacheDemo/pb"
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
	Get(in *pb.Request, out *pb.Response) error
}

type HttpGetter struct {
	baseUrl string
}

func (h *HttpGetter) Get(in *pb.Request, out *pb.Response) error {
	// remote url
	u := fmt.Sprintf("%v,%v/%v", h.baseUrl, url.QueryEscape(in.GetGroup()), url.QueryEscape(in.GetKey()))
	res, err := http.Get(u)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server return:%v", res.Status)
	}

	protos, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading body:%v", err)
	}
	if err = proto.Unmarshal(protos, out); err != nil {
		return fmt.Errorf("decode response body:%v ", err)
	}

	return nil
}
