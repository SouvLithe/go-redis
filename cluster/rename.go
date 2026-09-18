package cluster

import (
	"go-redis/interface/resp"
	"go-redis/resp/reply"
)

/*
本demo没有实现rename在不同节点的功能，因为如果一改key可能去到别的节点
后面可以加
所以只判断rename的key是否在同一节点，是，做rename；否，不做，只返回错误
*/
func Rename(cluster *ClusterDatabase, c resp.Connection, cmdArgs [][]byte) resp.Reply {
	if len(cmdArgs) != 3 {
		return reply.MakeErrReply("ERR Wrong number args")
	}
	src := string(cmdArgs[1])
	dest := string(cmdArgs[2])
	// 拿到源地址 和 目标地址
	srcPeer := cluster.peerPicker.PickNode(src) //192....:6379
	destPeer := cluster.peerPicker.PickNode(dest)

	if srcPeer != destPeer {
		return reply.MakeErrReply("ERR rename must within same peer")
	}
	return cluster.relay(srcPeer, c, cmdArgs)
}
