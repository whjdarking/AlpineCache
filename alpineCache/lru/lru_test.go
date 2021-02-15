package lru

import (
	"reflect"
	"testing"
)
//为了测试，我们假设缓存Value是一个string
type String string

//实现Value接口所需的Len()方法
func (s String) Len() int {
	return len(s)
}

//测试Add的更新功能
func TestAdd(t *testing.T) {
	lru := New(int64(0), nil) //0表示大小不限制
	lru.Add("key", String("1"))
	lru.Add("key", String("111"))

	if lru.nBytes != int64(len("key")+len("111")) {
		t.Fatal("expect 6, but now get", lru.nBytes)
	}
}

//测试Get功能
func TestGet(t *testing.T) {
	lru := New(int64(0), nil)
	lru.Add("key1", String("test1"))
	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "test1" {
		t.Fatalf("expect to find key1=test1，but fail")
	}
	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("expect to not find key2, but find")
	}
}

//测试触发maxBytes的自动删除
func TestRemoveOldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := "value1", "value2", "value3"
	cap := len(k1 + k2 + v1 + v2)
	lru := New(int64(cap), nil)
	lru.Add(k1, String(v1))
	lru.Add(k2, String(v2))
	lru.Add(k3, String(v3))

	if _, ok := lru.Get("key1"); ok || lru.Len() != 2 {
		t.Fatalf("expect autodelete because of maxBytes, but fail")
	}
}

//
func TestOnDelete(t *testing.T) {
	keys := make([]string, 0)
	//设置一个回调函数，将key加在一个slice里。注意删除时才会触发回调。
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}
	lru := New(int64(10), callback)
	lru.Add("key1", String("123456"))
	lru.Add("k2", String("k2"))
	lru.Add("k3", String("k3"))
	lru.Add("k4", String("k4"))
	expect := []string{"key1", "k2"}
	if !reflect.DeepEqual(expect, keys) {
		t.Fatalf("something wrong in callback, expect %s", expect)
	}
}
