package cluster

import "go-redis/interface/resp"

/*
转发节点时，会给每个节点重新发送一个select，以防他们不知道用户使用的是什么select
所以select用户用的db号记录在本地即可
*/

// 本地执行方式
func execSelect(cluster *ClusterDatabase, c resp.Connection, cmdArgs [][]byte) resp.Reply {
	return cluster.db.Exec(c, cmdArgs)
}
