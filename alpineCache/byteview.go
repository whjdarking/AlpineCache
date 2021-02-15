package alpineCache

//作为储存中最主要的数据成员Value
type ByteView struct {
	b []byte
}

//满足被储存对象的Len()方法
func (v ByteView) Len() int {
	return len(v.b)
}

//返回一个克隆，防止原地址被外部修改
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

//有时需要转成string类型，方便查看或比较
func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}