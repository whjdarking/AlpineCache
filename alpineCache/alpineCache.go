package alpineCache

import (
	"alpineCache/onlySendOnce"
	"fmt"
	"log"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

//典型的用函数类型实现接口。这样利用强制转换成GetterFunc，我们可以合法传入任何名字的函数去作为Get。同时身为接口，我们保留了传入结构体的合法性。
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

//以下是group及相关操作
//group是整个系统的最中心部件
type Group struct {
	name      string
	getter    Getter     //缓存未命中时应触发Getter接口的实现（仅代表自己的getter）
	mainCache cache      //实例化且加锁的lru
	peers     PeerPicker //通过picker去寻找其它节点，或者说其它节点的getter。我们这里httppool就实现了picker。

	loader *onlySendOnce.Group //
}

var mu sync.RWMutex                  //读写锁
var groups = make(map[string]*Group) //储存所有group

//简单的通过groups和name获取group的操作
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

//实例化
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("Getter is nil")
	}
	mu.Lock()
	defer mu.Unlock()

	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes}, //同目录下的cache，传入大小
		loader:    &onlySendOnce.Group{},
	}
	groups[name] = g
	return g
}

// 传入group的picker
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// 从Cache中查找是否有缓存，有的话返回值，没有的话调用load试图从用户获取
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("successfully find in cache")
		return v, nil
	}

	return g.load(key)
}

//选择从哪里获得数据
func (g *Group) load(key string) (value ByteView, err error) {
	once, err := g.loader.Do(key, func() (interface{}, error) {
		//group里的peer不会为空
		if g.peers != nil {
			//能picker到对象（对应节点的getter）
			//如果得到的结果就是本机，也会返回false的,这样就会走下面的getLocally
			if peer, ok := g.peers.PickPeer(key); ok {
				//正常获得了缓存信息
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("Failed to get from peer", err)
			}
		}

		return g.getLocally(key)
	})
	if err == nil {
		return once.(ByteView), nil //类型断言检查
	}
	return
}

//调用getter里的方法，本地
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)} //重新拷贝一份防止客户端之后依然可以改动这一份数据
	g.populateCache(key, value)             //正式添加入mainCache
	return value, nil
}

//从peer那里调用get
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	//选定了对应节点的getter后，peergetter的返回值就是缓存拉
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}

//正式添加入group里的mainCache
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}
