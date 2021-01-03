/**
* @Author:zhoutao
* @Date:2021/1/3 上午11:45
* @Desc:
 */

package GoCacheDemo

//只读
type ByteView struct {
	b []byte //存储缓存值,byte类型能够支持任意类型的数据类型的存储，例如字符串、图片等
}

func (v ByteView) Len() int {
	return len(v.b)
}

func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}

// b是只读的，使用ByteSlice()方法返回一个拷贝，防止缓存值被外部程序修改
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}
