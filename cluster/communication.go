package cluster

import (
	"context"
	"errors"
	"go-redis/interface/resp"
	"go-redis/lib/utils"
	"go-redis/resp/client"
	"go-redis/resp/reply"
	"strconv"
)

// 本地
// 拿一个连接
func (cluster *ClusterDatabase) getPeerClient(peer string) (*client.Client, error) {
	// 拿到连接池
	factory, ok := cluster.peerConnection[peer]
	if !ok {
		return nil, errors.New("connection factory not found")
	}
	// 获取其中一个连接
	raw, err := factory.BorrowObject(context.Background())
	if err != nil {
		return nil, err
	}
	// 把客户端转成响应类型
	c, ok := raw.(*client.Client)
	if !ok {
		return nil, errors.New("connection factory make wrong type")
	}
	return c, nil
}

// 还一个连接
func (cluster *ClusterDatabase) returnPeerClient(peer string, peerClient *client.Client) error {
	connectionFactory, ok := cluster.peerConnection[peer]
	if !ok {
		return errors.New("connection factory not found")
	}
	return connectionFactory.ReturnObject(context.Background(), peerClient)
}

// 转发
func (cluster *ClusterDatabase) relay(peer string, c resp.Connection, args [][]byte) resp.Reply {
	// 如果目标节点是自己
	if peer == cluster.self {
		return cluster.db.Exec(c, args)
	}
	// 成功借出连接
	peerClient, err := cluster.getPeerClient(peer)
	if err != nil {
		return reply.MakeErrReply(err.Error())
	}
	defer func() {
		_ = cluster.returnPeerClient(peer, peerClient)
	}()
	// 先切库，再转发真正的指令
	selectReply := peerClient.Send(utils.ToCmdLine("SELECT", strconv.Itoa(c.GetDBIndex())))
	if reply.IsErrReply(selectReply) {
		return selectReply // 或者制造一个错误返回
	}
	return peerClient.Send(args)
}

// 广播
func (cluster *ClusterDatabase) broadcast(c resp.Connection, args [][]byte) map[string]resp.Reply {
	results := make(map[string]resp.Reply)
	for _, node := range cluster.nodes {
		result := cluster.relay(node, c, args)
		results[node] = result
	}
	return results
}
