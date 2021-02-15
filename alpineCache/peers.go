package alpineCache

//每一个节点对应一个getter，所以想从其它节点获取缓存时，得先通过key找到对应得节点得getter
//注意pool本身实现了peerpicker接口。后面pool被直接存在group里
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

//每个pool里面的的getter成员要继承这个接口
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}