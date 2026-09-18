package database

import (
	"go-redis/interface/databaseface"
	"go-redis/interface/resp"
	"go-redis/lib/utils"
	"go-redis/resp/reply"
)

/*
GET SET SETNX GETSET STRLEN
*/

// GET k1
func execGet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		return reply.MakeNullBulkReply()
	}
	// 如果实现其它类型，需要判断类型转换是否成功，用两个参数接收，第二个就是是否成功
	bytes := entity.Data.([]byte)
	return reply.MakeBulkReply(bytes)
}

// SET k1 v
func execSet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]
	entity := &databaseface.DataEntity{
		Data: value,
	}
	db.PutEntity(key, entity)

	// 添加aof记录
	db.addAof(utils.ToCmdLine2("set", args...))
	return reply.MakeOKReply()
}

// SETNX k1 v1
func execSetNX(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]
	entity := &databaseface.DataEntity{
		Data: value,
	}
	result := db.PutIFAbsent(key, entity)

	// 添加aof记录
	db.addAof(utils.ToCmdLine2("setnx", args...))
	return reply.MakeIntReply(int64(result))
}

// GETSET k1 v1,获取k1原来的值并返回，切将k1原来的值设为v1
func execGetSet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]
	entity, exists := db.GetEntity(key)
	// 添加aof记录
	db.addAof(utils.ToCmdLine2("getset", args...))
	if !exists {
		return reply.MakeNullBulkReply()
	}
	db.PutEntity(key, &databaseface.DataEntity{Data: value})
	return reply.MakeBulkReply(entity.Data.([]byte))
}

// STRLEN
func execStrLen(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, exists := db.GetEntity(key)
	if !exists {
		return reply.MakeNullBulkReply()
	}
	bytes := entity.Data.([]byte)
	return reply.MakeIntReply(int64(len(bytes)))
}

func init() {
	RegisterCommand("Get", execGet, 2) // get k1
	RegisterCommand("Set", execSet, 3) // set k1 v1
	RegisterCommand("SetNX", execSetNX, 3)
	RegisterCommand("GetSet", execGetSet, 3)
	RegisterCommand("StrLen", execStrLen, 2)
}
