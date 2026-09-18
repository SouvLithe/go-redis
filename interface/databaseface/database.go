package databaseface

import "go-redis/interface/resp"

type CmdLine = [][]byte

type Database interface {
	Exec(client resp.Connection, args [][]byte) resp.Reply
	Close()
	AfterClientClose(client resp.Connection)
}

// 用来指代redis的各种数据类型
type DataEntity struct {
	Data interface{}
}
