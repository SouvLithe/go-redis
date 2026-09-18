package cluster

import (
	"go-redis/interface/resp"
	"go-redis/resp/reply"
)

// 广播
func FlushDB(cluster *ClusterDatabase, c resp.Connection, cmdArgs [][]byte) resp.Reply {
	replies := cluster.broadcast(c, cmdArgs)
	// 广播执行后，所有的节点ok才行，不ok报错
	var errReply reply.ErrorReply
	for _, r := range replies {
		// 遇到一个错误，就停止遍历
		if reply.IsErrReply(r) {
			errReply = r.(reply.ErrorReply)
			break
		}
	}
	if errReply != nil {
		return &reply.OKReply{}
	}
	return reply.MakeErrReply("error occur: " + errReply.Error())
}
