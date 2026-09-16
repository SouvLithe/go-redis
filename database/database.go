package database

import (
	"go-redis/aof"
	"go-redis/config"
	"go-redis/interface/resp"
	"go-redis/lib/logger"
	"go-redis/resp/reply"
	"strconv"
	"strings"
)

type Database struct {
	dbSet      []*DB // 一组指针
	aofHandler *aof.AofHandler
}

// 用于新建database
func NewDatabase() *Database {
	mdb := &Database{}
	// 从config中拿到配置文件conf中的数据库个数
	if config.Properties.Databases == 0 {
		config.Properties.Databases = 16
	}
	mdb.dbSet = make([]*DB, config.Properties.Databases)
	for i := range mdb.dbSet {
		db := makeDB()    // 创建DB
		db.index = i      // 给DB编号
		mdb.dbSet[i] = db // 将DB存进去
	}

	// 检查是否有配置
	if config.Properties.AppendOnly {
		aofHandler, err := aof.NewAofHandler(mdb)
		if err != nil {
			// 如果新建Aof处理器错误，那用户数据保护就有问题，选择崩掉
			panic(err)
		}
		mdb.aofHandler = aofHandler
		// 为了让db能够使用AddAof方法，使用匿名方法调用aof处理器的AddAof方法
		for _, db := range mdb.dbSet {
			// db = dbSet[0] -> dbSet[15] 出现闭包问题
			// 内部变量引用外部变量，导致这个变量逃逸到了堆上
			sdb := db
			sdb.addAof = func(line CmdLine) {
				mdb.aofHandler.AddAof(sdb.index, line)
			}
		}
	}

	return mdb
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
