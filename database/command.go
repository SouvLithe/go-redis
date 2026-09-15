package database

import "strings"

var cmdTable = make(map[string]*command)

// 每个指令都是一个command结构体，在这里实现方法，在DB中作用到数据上
type command struct {
	exector ExecFunc
	arity   int // 参数的数量
}

// 注册指令方法
func RegisterCommand(name string, exector ExecFunc, arity int) {
	name = strings.ToLower(name)
	cmdTable[name] = &command{
		exector: exector,
		arity: arity,
	}
}
