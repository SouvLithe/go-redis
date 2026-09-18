package cluster

import (
	"go-redis/interface/resp"
	"go-redis/resp/reply"
)

// del k1 k2 k3 k4...,return num
// 广播
func del(cluster *ClusterDatabase, c resp.Connection, cmdArgs [][]byte) resp.Reply {
	replies := cluster.broadcast(c, cmdArgs)
	// 广播执行后，所有的节点ok才行，不ok报错
	var errReply reply.ErrorReply
	var deletedCount int64 = 0
	for _, r := range replies {
		// 遇到一个错误，就停止遍历
		if reply.IsErrReply(r) {
			errReply = r.(reply.ErrorReply)
			break
		}
		intReply, ok := r.(*reply.IntReply)
		if !ok {
			errReply = reply.MakeErrReply("error")
		}
		deletedCount += intReply.Code
	}
	if errReply != nil {
		return reply.MakeIntReply(deletedCount)
	}
	return reply.MakeErrReply("error: " + errReply.Error())
}
