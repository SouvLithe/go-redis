package database

import (
	"go-redis/interface/resp"
	"go-redis/lib/wildcard"
	"go-redis/resp/reply"
)

/*
实现DEL EXISTS KEYS FLUSHDB TYPE RENAME RENAMENX
*/

// DEL k1,k2,k3
func execDel(db *DB, args [][]byte) resp.Reply {
	keys := make([]string, len(args))
	for i, v := range args {
		keys[i] = string(v)
	}
	// 记录一下操作参数数量
	deleted := db.Removes(keys...)
	return reply.MakeIntReply(int64(deleted))
}

// EXISTS k1,k2,k3
func execExists(db *DB, args [][]byte) resp.Reply {
	result := int64(0)
	for _, arg := range args {
		key := string(arg)
		_, exists := db.GetEntity(key)
		if exists {
			result++
		}
	}
	return reply.MakeIntReply(result)
}

// FLUSHDB
func execFlushDB(db *DB, args [][]byte) resp.Reply {
	db.Flush()
	return reply.MakeOKReply()
}

// TYPE k1
func execType(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exist := db.GetEntity(key)
	if !exist {
		return reply.MakeStatusReply("none") // TCP :none/r/n
	}
	switch entity.Data.(type) {
	case []byte:
		return reply.MakeStatusReply("string")
	}
	// 如果有其它数据结构的话，把case放在这里
	return reply.UnknowErrReply{}
}

// RENAME k1 k2 k1:v -> k2,v
func execRename(db *DB, args [][]byte) resp.Reply {
	src := string(args[0])
	dest := string(args[1])
	entity, exists := db.GetEntity(src)
	if !exists {
		return reply.MakeErrReply("no such key")
	}
	db.PutEntity(dest, entity)
	db.Remove(src)
	return reply.MakeOKReply()
}

// RENAMENX k1 k2,就是检查一下k2的值是否本来就存在于map
func execRenameNX(db *DB, args [][]byte) resp.Reply {
	src := string(args[0])
	dest := string(args[1])

	_, ok := db.GetEntity(dest)
	if ok {
		return reply.MakeIntReply(0)
	}

	entity, exists := db.GetEntity(src)
	if !exists {
		return reply.MakeErrReply("no such key")
	}
	db.PutEntity(dest, entity)
	db.Remove(src)
	return reply.MakeIntReply(1)
}

// KEYS *
func execKeys(db *DB, args [][]byte) resp.Reply {
	pattern := wildcard.CompilePattern(string(args[0]))
	// 所有符合通配符的key放入result中
	result := make([][]byte, 0)
	db.data.ForEach(func(key string, value interface{}) bool {
		if pattern.IsMatch(key) {
			result = append(result, []byte(key))
		}
		return true
	})
	return reply.MakeMultiBulkReply(result)
}

func init() {
	RegisterCommand("DEL", execDel, -2)
	RegisterCommand("EXISTS", execExists, -2)
	RegisterCommand("flushdb", execFlushDB, -1) // FLUSHDB a b c
	RegisterCommand("Type", execType, 2)        // TYPE k1
	RegisterCommand("RENAME", execRename, 3)    // RENAME k1 k2
	RegisterCommand("RENAMENX", execRenameNX, 3)
	RegisterCommand("KEYS", execKeys, 2) // KEY *
}
