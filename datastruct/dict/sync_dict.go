package dict

import "sync"

// 最底层保存redis数据的结构
type SyncDict struct {
	m sync.Map
}

func MakeSyncDict() *SyncDict {
	return &SyncDict{}
}

func (dict *SyncDict) Get(key string) (val interface{}, exists bool) {
	value, ok := dict.m.Load(key)
	return value, ok
}

func (dict *SyncDict) Len() int {
	lenth := 0
	dict.m.Range(func(key, value interface{}) bool {
		lenth++
		// 返回true，才会进行下一个元素的相加
		return true
	})
	return lenth
}

func (dict *SyncDict) Put(key string, val interface{}) (result int) {
	_, existed := dict.m.Load(key)
	dict.m.Store(key, val)
	if existed {
		// 如果只是修改，并没有插入新的
		return 0
	}
	return 1
}

func (dict *SyncDict) PutIfAbsent(key string, val interface{}) (result int) {
	_, existed := dict.m.Load(key)
	if existed {
		return 0
	}
	dict.m.Store(key, val)
	return 1
}

func (dict *SyncDict) PutIfExists(key string, val interface{}) (result int) {
	_, existed := dict.m.Load(key)
	if existed {
		dict.m.Store(key, val)
		return 1
	}
	return 0
}

func (dict *SyncDict) Remove(key string) (result int) {
	_, existed := dict.m.Load(key)
	if !existed {
		return 0
	}
	dict.m.Delete(key)
	return 1
}

func (dict *SyncDict) ForEach(consumer Consumer) {
	dict.m.Range(func(key, value interface{}) bool {
		consumer(key.(string), value)
		// 不判断在哪里终止，直接返回true就可以了
		return true
	})
}

func (dict *SyncDict) Keys() []string {
	// 1. 长度设为 0，容量预估为 dict.Len()，避免频繁扩容
	result := make([]string, 0, dict.Len())
	dict.m.Range(func(key, value interface{}) bool {
		// 2. 使用 append，动态增长，天然避免越界和空值
		result = append(result, key.(string))
		return true
	})
	return result
}

func (dict *SyncDict) RandomKeys(limit int) []string {
	result := make([]string, dict.Len())
	for i := 0; i < limit; i++ {
		dict.m.Range(func(key, value interface{}) bool {
			result[i] = key.(string)
			return false
		})
	}
	return result
}

func (dict *SyncDict) RandomDistinctKeys(limit int) []string {
	// 预估分配容量，避免切片频繁扩容
	result := make([]string, 0, limit)
	dict.m.Range(func(key, value interface{}) bool {
		result = append(result, key.(string))
		return len(result) < limit
	})
	return result
}

func (dict *SyncDict) Clear() {
	// 直接申请一个新的，旧的让GC回收掉
	*dict = *MakeSyncDict()
}
