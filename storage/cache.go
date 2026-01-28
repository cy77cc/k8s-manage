package storage

import (
	"container/list"
	"sync"
)

type entry[K comparable, V any] struct {
	key   K
	value V
}

type Cache[K comparable, V any] struct {
	capacity int
	ll       *list.List
	mu       sync.RWMutex
	cache    map[K]*list.Element
}

// NewCache 创建一个 LRU 实例
func NewCache[K comparable, V any](capacity int) *Cache[K, V] {

	// 考虑边界情况capacity可能是空的
	if capacity <= 0 {
		capacity = 0
	}

	// TODO: 返回一个初始化好的 *Cache
	return &Cache[K, V]{
		capacity: capacity,
		ll:       list.New(),
		cache:    make(map[K]*list.Element),
	}
}

func (l *Cache[K, V]) Get(key K) (V, bool) {
	// TODO: 实现 Cache 的 Get 逻辑
	var zero V

	// 1. 获取key对应的元素

	l.mu.RLock()
	defer l.mu.RUnlock()

	// 如果此时没有cache应该直接返回空
	if len(l.cache) == 0 {
		return zero, false
	}

	val, ok := l.cache[key]
	if !ok {
		return zero, false
	}

	// 2. 访问过的元素要放到最前面
	l.ll.MoveToFront(val)
	ent := val.Value.(*entry[K, V])

	return ent.value, true
}

func (l *Cache[K, V]) Put(key K, value V) {
	// TODO: 实现 LRU 的 Put 逻辑（需要处理更新与淘汰）

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.capacity <= 0 {
		return
	}

	// 如果key已经存在了，应该是更新
	if elem, ok := l.cache[key]; ok {
		ent := elem.Value.(*entry[K, V])
		ent.value = value
		l.ll.MoveToFront(elem)
		return
	}

	// 1. 看一下有没有大于容量
	if len(l.cache) >= l.capacity {
		back := l.ll.Back()
		if back != nil {
			ent := back.Value.(*entry[K, V])
			delete(l.cache, ent.key)
			l.ll.Remove(back)
		}
	}

	// 插入新元素
	ent := &entry[K, V]{key: key, value: value}
	elem := l.ll.PushFront(ent)
	l.cache[key] = elem
}

func (l *Cache[K, V]) Len() int {
	if l.ll == nil {
		return 0
	}
	return l.ll.Len()
}
