package alpineCache

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)
//测试Getter接口有没有正常关联上GetterFunc
func TestGetter(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	expect := []byte("key")
	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
		t.Fatal("Getter失败")
	}
}

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func TestGet(t *testing.T) {
	//访问一个键时调用了几次load
	loadCounts := make(map[string]int, len(db))
	//给一个getter函数，内容是获取db里的数据
	alpine := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[not from cache] search key", key)
			if v, ok := db[key]; ok {
				//如果第一次查询loadcounts，创建并置0
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key]++
				return []byte(v), nil //转成byte传出去
			}
			return nil, fmt.Errorf("%s not exist in db", key)
		}))

	for k, v := range db {
		//是否能从本地通过Getter获取数据
		if view, err := alpine.Get(k); err != nil || view.String() != v {
			t.Fatalf("failed to get value of %v", k)
		}
		//后续再试图Get时，是否能通过缓存直接读到
		if _, err := alpine.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		}
	}

	if view, err := alpine.Get("unknown"); err == nil {
		t.Fatalf("the value of unknow should be empty, but %s got", view)
	}
}
