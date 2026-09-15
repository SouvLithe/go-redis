package database

import (
	"go-redis/interface/resp"
	"go-redis/resp/reply"
)

// 只回发测试用的db，发给什么指令返回什么指令
type EchoDatabase struct {
}

func NewEchoDatabase() *EchoDatabase {
	return &EchoDatabase{}
}

func (e EchoDatabase) Exec(client resp.Connection, args [][]byte) resp.Reply {
	return reply.MakeMultiBulkReply(args)
}

func (e EchoDatabase) Close() {

}

func (e EchoDatabase) AfterClientClose(client resp.Connection) {

}
