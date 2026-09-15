package database

import (
	"go-redis/config"
	"go-redis/interface/resp"
	"go-redis/lib/logger"
	"go-redis/resp/reply"
	"strconv"
	"strings"
)

type Database struct {
	dbSet []*DB // 一组指针
}

// 用于新建database
func NewDatabase() *Database {
	database := &Database{}
	// 从config中拿到配置文件conf中的数据库个数
	if config.Properties.Databases == 0 {
		config.Properties.Databases = 16
	}
	database.dbSet = make([]*DB, config.Properties.Databases)
	for i := range database.dbSet {
		db := makeDB()         // 创建DB
		db.index = i           // 给DB编号
		database.dbSet[i] = db // 将DB存进去
	}
	return database
}

// select 1  or  65525过大
func execSelect(c resp.Connection, database *Database, args [][]byte) resp.Reply {
	dbIndex, err := strconv.Atoi(string(args[0]))
	if err != nil {
		return reply.MakeErrReply("ERR invalid DB index")
	}
	if dbIndex >= len(database.dbSet) {
		return reply.MakeErrReply("ERR DB index out of range")
	}
	c.SelectDB(dbIndex)
	return reply.MakeOKReply()
}

/*
下面3个函数在这一层没有特殊的逻辑，为空即可
*/
// Set k v \ GET K \ SELECT 2
func (database *Database) Exec(client resp.Connection, args [][]byte) resp.Reply {
	// 向上抛出panic
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
		}
	}()
	// 判断选择指令是否合规
	cmdName := strings.ToLower(string(args[0]))
	if cmdName == "select" {
		if len(args) != 2 {
			return reply.MakeArgNumErrReply("select")
		}
		return execSelect(client, database, args[1:])
	}
	// 根据指针获取对应分数据库db，并在其中执行用户命令
	dbIndex := client.GetDBIndex()
	db := database.dbSet[dbIndex]
	return db.Exec(client, args)
}

func (database *Database) Close() {
}

func (database *Database) AfterClientClose(client resp.Connection) {
}
