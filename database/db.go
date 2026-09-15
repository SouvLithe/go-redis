package database

import (
	"go-redis/datastruct/dict"
	"go-redis/interface/database"
	"go-redis/interface/resp"
	"go-redis/resp/reply"
	"strings"
)

// dict的上层数据结构
type DB struct {
	index int
	data  dict.Dict
}

// 所有redis指令形式的实现
type ExecFunc func(db *DB, args [][]byte) resp.Reply

type CmdLine = [][]byte

func makeDB() *DB {
	db := &DB{
		data: dict.MakeSyncDict(),
	}
	return db
}

func (db *DB) Exec(c resp.Connection, cmdLine CmdLine) resp.Reply {
	// 得知用户使用的指令 PING SET...
	cmdName := strings.ToLower(string(cmdLine[0]))
	// ok是map的作用，告诉有没有找到值
	cmd, ok := cmdTable[cmdName]
	if !ok {
		return reply.MakeErrReply("ERR unknown command: " + cmdName)
	}
	if !validateArity(cmd.arity, cmdLine) {
		return reply.MakeArgNumErrReply(cmdName)
	}
	fun := cmd.exector
	// SET K V -> K V
	return fun(db, cmdLine[1:])
}

// 定长参数 SET K V -> arity = 3
// 变长参数 EXISTS k1,k2,k3,k4.... -> arity = -2
// 检验发过来的指令符不符合要求的个数
func validateArity(arity int, cmdArgs [][]byte) bool {
	argNum := len(cmdArgs)
	if arity >= 0 {
		return argNum == arity
	}
	return argNum >= -arity
}

func (db *DB) GetEntity(key string) (*database.DataEntity, bool) {
	raw, ok := db.data.Get(key)
	if !ok {
		// 读取失败
		return nil, false
	}
	entity, _ := raw.(*database.DataEntity)
	return entity, true
}

func (db *DB) PutEntity(key string, entity *database.DataEntity) int {
	// 形参是空接口，实参在被传入时，会自动转换为空接口
	return db.data.Put(key, entity)
}

func (db *DB) PutIFExists(key string, entity *database.DataEntity) int {
	// 形参是空接口，实参在被传入时，会自动转换为空接口
	return db.data.PutIfExists(key, entity)
}

func (db *DB) PutIFAbsent(key string, entity *database.DataEntity) int {
	// 形参是空接口，实参在被传入时，会自动转换为空接口
	return db.data.PutIfAbsent(key, entity)
}

func (db *DB) Remove(key string) {
	db.data.Remove(key)
}

// 接收变长参数
func (db *DB) Removes(keys ...string) (deleted int) {
	deleted = 0
	for _, key := range keys {
		_, exists := db.data.Get(key)
		if exists {
			db.Remove(key)
			deleted++
		}
	}
	return deleted
}

func (db *DB) Flush() {
	db.data.Clear()
}
