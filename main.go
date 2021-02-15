package main

import (
	"alpineCache"
	"flag"
	"fmt"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "597",
	"Jack": "7138",
	"Sam":  "217",
}

func createGroup() *alpineCache.Group {
	//传入getterfunc，这里作为测试，就是从本地（也就是上面写的db）读取key返回value
	return alpineCache.NewGroup("scores", 2<<10, alpineCache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}


//apiserver的路径等可以单独修改，对外暴露。这是apiserver发起group.Get请求
//apiserver和cache的区别是apiserver的group没有设定group.RegisterPeers，是无法正常工作的
//但当cacheServer设置后，同一个group的apiserver就可以工作了。
//换句话说这里是为了方便测试，一组group中，一定要有一个api为true，才能正常使用api功能。如果不启用api，直接向cacheServer发送符合cacheServer的URL也行。
func startAPIServer(apiAddr string, group *alpineCache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := group.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())

		}))
	log.Println("frontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil)) //去除http://

}

//传入本地地址，所有节点地址，和group
func startCacheServer(addr string, addrs []string, group *alpineCache.Group) {
	//在本地实例化一个httppool
	peers := alpineCache.NewHTTPPool(addr)
	//传入其它的所有节点
	peers.Set(addrs...)
	//把设置好的pool传给examplegroup
	group.RegisterPeers(peers)
	log.Println("cache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "cache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	exampleGroup := createGroup()

	if api {
		go startAPIServer(apiAddr, exampleGroup)
	}
	startCacheServer(addrMap[port], addrs, exampleGroup)


	//原测试代码，现存在creategroup里
	//alpineCache.NewGroup("scores", 2<<10, alpineCache.GetterFunc(
	//	func(key string) ([]byte, error) {
	//		log.Println("[SlowDB] search key", key)
	//		if v, ok := db[key]; ok {
	//			return []byte(v), nil
	//		}
	//		return nil, fmt.Errorf("%s not exist", key)
	//	}))
	//
	//addr := "localhost:9999"
	//peers := alpineCache.NewHTTPPool(addr)
	//log.Println("alpineCache is running at", addr)
	//log.Fatal(http.ListenAndServe(addr, peers))
}