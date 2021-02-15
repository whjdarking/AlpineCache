package lru

import "container/list"

type Cache struct {
	maxBytes int64                         //允许使用最大内存
	nBytes int64                           //当前内存
	linkedList *list.List                  //使用Go自带的双向链表
	cache map[string]*list.Element         //hash表连接每一个list元素
	OnDelete func(key string, value Value) //被移除时的回调函数,如果需要可以自己任意实现
}

//CU
func (c *Cache) Add(key string, value Value) {
	//如果已经存在，移动到front并更新值
	if element, ok := c.cache[key]; ok {
		c.linkedList.MoveToFront(element)
		kv := element.Value.(*listElement)
		c.nBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {  //如果不存在，插入到front，再加入map
		element := c.linkedList.PushFront(&listElement{key, value})
		c.cache[key] = element
		c.nBytes += int64(len(key)) + int64(value.Len())
	}
	//如果超出大小，不断删除直到满足
	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}

//R
func (c *Cache) Get(key string) (value Value, ok bool) {
	if element, ok := c.cache[key]; ok {
		c.linkedList.MoveToFront(element)
		kv := element.Value.(*listElement)
		return kv.value, true
	}
	return
}
//D
func (c *Cache) RemoveOldest() {
	element := c.linkedList.Back()
	if element != nil {
		c.linkedList.Remove(element)  //移除出list
		kv := element.Value.(*listElement)
		delete(c.cache, kv.key)  //移除出map
		c.nBytes = c.nBytes - int64(len(kv.key)) - int64(kv.value.Len())
		if c.OnDelete != nil {
			c.OnDelete(kv.key, kv.value)
		}
	}
}

//测试用，返回链表长度
func (c *Cache) Len() int {
	return c.linkedList.Len()
}


//链表节点的数据类型，保存key是为了方便从v找k
type listElement struct {
	key   string
	value Value
}

//listElement链表节点中的Value，要求实现Len()返回大小
type Value interface {
	Len() int
}

//实例化，传入max大小和自订的回调函数
func New(maxBytes int64, onDelete func(string, Value)) *Cache {
	return &Cache{
		maxBytes: maxBytes,
		linkedList: list.New(),
		cache: make(map[string]*list.Element),
		OnDelete: onDelete,
	}
}
