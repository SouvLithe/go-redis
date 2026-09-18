package cluster

import "go-redis/interface/resp"

/*
告知指令的路由模式：本地、转发、广播
*/
type CmdLine = [][]byte

// map入参是指令名称，出参是指令方法
func makeRouter() map[string]CmdFunc {
	routerMap := make(map[string]CmdFunc)
	// 默认转发
	routerMap["exists"] = defaultFunc //exist k1
	routerMap["type"] = defaultFunc
	routerMap["set"] = defaultFunc
	routerMap["setnx"] = defaultFunc
	routerMap["get"] = defaultFunc
	routerMap["getset"] = defaultFunc
	// 本地转发
	routerMap["ping"] = ping
	routerMap["rename"] = Rename
	routerMap["renamenx"] = Rename
	// 若对端也运行在集群模式，SELECT 不在路由表中会返回 "unknown command" 错误。
	// execSelect 会委托给 standalone database 本地执行，返回 +OK。
	routerMap["select"] = execSelect
	// 广播
	routerMap["flushdb"] = FlushDB
	routerMap["del"] = del
	return routerMap
}

// 大部分操作都走这个默认方法，一个key走这种
// GET key // Set k1 v1
func defaultFunc(cluster *ClusterDatabase, c resp.Connection, cmdArgs [][]byte) resp.Reply {
	key := string(cmdArgs[1])
	peer := cluster.peerPicker.PickNode(key)
	return cluster.relay(peer, c, cmdArgs)
}
