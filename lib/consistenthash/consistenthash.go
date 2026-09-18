package consistenthash

import (
	"hash/crc32"
	"sort"
)

/*
这个包用来实现一致性hash
*/

type HashFunc func(data []byte) uint32

type NodeMap struct {
	hashFunc    HashFunc
	nodeHashs   []int // 后面要对hash进行排序，go自带的排序方法，收的是int类型
	nodehashMap map[int]string
}

func NewNodeMap(fn HashFunc) *NodeMap {
	m := &NodeMap{
		hashFunc:    fn,
		nodehashMap: make(map[int]string),
	}
	// 如果没有设置hash函数，则默认使用 crc32.ChecksumIEEE
	if m.hashFunc == nil {
		m.hashFunc = crc32.ChecksumIEEE
	}
	return m
}

// 判断节点数量是否为空，集群节点是否初始化
func (m *NodeMap) isEmpty() bool {
	return len(m.nodehashMap) == 0
}

// 添加节点，可以一次加多个
func (m *NodeMap) AddNode(keys ...string) {
	for _, key := range keys {
		if key == "" {
			continue
		}
		hash := int(m.hashFunc([]byte(key)))
		m.nodeHashs = append(m.nodeHashs, hash)
		m.nodehashMap[hash] = key
	}
	sort.Ints(m.nodeHashs)
}

func (m *NodeMap) PickNode(key string) string {
	if m.isEmpty() {
		return ""
	}
	hash := int(m.hashFunc([]byte(key)))
	index := sort.Search(len(m.nodehashMap), func(i int) bool {
		return m.nodeHashs[i] >= hash
	})
	if index == len(m.nodehashMap) {
		index = 0
	}
	return m.nodehashMap[m.nodeHashs[index]]
}
