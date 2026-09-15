package database

import (
	"go-redis/interface/resp"
	"go-redis/resp/reply"
)

func Ping(db *DB, args [][]byte) resp.Reply {
	return reply.MakePongReply()
}

/*
go提供的方法，不管在那写个init方法，会保证在所在包初始化时，就会跑这个init方法
本方法作用：初始kv -> "ping", Ping到command包下的叫cmdTable的map中
*/
func init() {
	RegisterCommand("ping", Ping, 1)
}
