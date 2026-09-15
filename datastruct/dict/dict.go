package dict

type Consumer func(key string, value interface{}) bool

type Dict interface {
	Get(key string) (val interface{}, exists bool)
	Len() int // 返回字典数据量
	Put(key string, val interface{}) (result int)
	PutIfAbsent(key string, val interface{}) (result int)
	PutIfExists(key string, val interface{}) (result int)
	Remove(key string) (result int)
	ForEach(consumer Consumer)
	Keys() []string
	RandomKeys(limit int) []string         // 返回多个键
	RandomDistinctKeys(limit int) []string // 返回多个不重复的键
	Clear()
}
