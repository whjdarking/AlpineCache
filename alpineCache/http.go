package alpineCache

import (
	"alpineCache/hash"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const defaultBasePath = "/_cache/"
const defaultReplicas = 50

type HTTPPool struct {
	self        string //自己的地址
	basePath    string //设定后缀path，表示在调用本缓存api
	mu          sync.Mutex
	peers       *hash.Map              //用之前写的hash来选择节点
	httpGetters map[string]*httpGetter //下面实现的客户端，每一个getter对应一个节点
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

//核心的ServeHTTP方法
func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	//约定格式为 /<basepath>/<groupname>/<key>
	//将几个部分分开
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}
	//通过URL带来的key在缓存中查找
	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream") //未知的应用程序文件
	w.Write(view.ByteSlice())
}

//实例化
//为了方便，要求传入所有的peers也就是节点
func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = hash.New(defaultReplicas, nil) //hash实例，使用默认hash函数（传入nil）
	p.peers.Add(peers...)
	//每一个节点都存一个httpGetter
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

//继续使得HTTPPool实现PeerPicker接口
//根据key返回对应的getter
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	//下面的get是自己写的一致性hash的get，返回一个真实节点的名字string
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick node(peer) from %s", peer)
		//根据真实节点的string名，获得对应的getter
		return p.httpGetters[peer], true
	}
	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil) //接口断言

//每一个节点都应该有一个getter，通过这里面的baseURL区分
//存在httppool里名叫getters的map里面
type httpGetter struct {
	baseURL string
}

//实现peers.go里的PeerGetter接口
//这里相当于是客户端，发送get请求给另一个节点。注意这里的httpGetter已经是对应节点的getter了。
func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	//通过http库的Get发送请求到另一个节点，注意我们已经实现了ServeHTTP。这里是go自带http库的设计
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}

var _ PeerGetter = (*httpGetter)(nil) //接口断言
