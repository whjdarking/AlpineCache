package hash

import (
	"strconv"
	"testing"
)

func TestHashing(t *testing.T) {
	//创建一个实例，节点每个虚拟成三份。hash函数为了方便，直接传入一个返回key值的函数作为hash函数
	hash := New(3, func(key []byte) uint32 {
		i, _ := strconv.Atoi(string(key))
		return uint32(i)
	})

	// 手动算出我们应该得到如后面的节点：2, 4, 6, 12, 14, 16, 22, 24, 26
	//（这里比如4会变成4,14,24这样）
	hash.Add("6", "4", "2")

	//手动算出2 11 23 27应该归属哪个真实节点
	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}

	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}

	//再添加一个真实节点
	hash.Add("8")

	testCases["27"] = "8"

	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}

}