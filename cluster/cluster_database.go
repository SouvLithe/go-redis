package cluster

import (
	"context"
	"fmt"
	"go-redis/config"
	database "go-redis/database"
	"go-redis/interface/databaseface"
	"go-redis/interface/resp"
	"go-redis/lib/consistenthash"
	"go-redis/lib/logger"
	"go-redis/resp/reply"
	"runtime/debug"
	"strings"

	pool "github.com/jolestar/go-commons-pool/v2"
)

/*
涉及 lib/consistenthash  resp/client
database/standalone_database.go 只是改了名字 还有 go.mod
在单节点database上套一层集群版的cluster_database，集群版的内核不做业务，只做转发.
此外同级别的cluster还要做到可以相互通信转发请求，就意味着它们要成为彼此的client/server。
*/

type ClusterDatabase struct {
	self           string                      // 记录本节点
	nodes          []string                    // 记录整个集群的节点
	peerPicker     *consistenthash.NodeMap     // 一致性hash
	peerConnection map[string]*pool.ObjectPool // github上的池化工具;用map保存多个连接池，如3个节点要用2个连接池
	db             databaseface.Database
}

func MakeClusterDatabase() *ClusterDatabase {
	cluster := &ClusterDatabase{
		self:           config.Properties.Self,
		db:             database.NewStandaloneDatabase(), // 其实集群版也会启动单机版的database
		peerPicker:     consistenthash.NewNodeMap(nil),   // 如果有其它的hash函数通过这个传入
		peerConnection: make(map[string]*pool.ObjectPool),
	}
	nodes := make([]string, 0, len(config.Properties.Peers)+1)

	// 加入集群内出自己外其余节点
	for _, peer := range config.Properties.Peers {
		nodes = append(nodes, peer)
	}
	// 把自己作为节点加进去
	nodes = append(nodes, config.Properties.Self)
	// 初始化一致性hash
	cluster.peerPicker.AddNode(nodes...)
	// 初始化一个空的ctx
	ctx := context.Background()
	// 加入集群内出自己外其余节点,新建连接池
	for _, peer := range config.Properties.Peers {
		// 这里要记录，否则找不到集群内其它节点
		cluster.peerConnection[peer] = pool.NewObjectPoolWithDefaultConfig(ctx, &connectionFactory{
			Peer: peer,
		})
	}
	cluster.nodes = nodes
	return cluster
}

// 统一签名函数
type CmdFunc func(cluster *ClusterDatabase, c resp.Connection, cmdArgs [][]byte) resp.Reply

// 初始化路由表
var router = makeRouter()

// 对集群层进行操作
// 替代单机版的Exec，解析完直接调用这个
// 语法糖：返回值如果起一个参数名，相当于在函数起始行，进行初始化
func (c *ClusterDatabase) Exec(client resp.Connection, cmdLine [][]byte) (result resp.Reply) {
	defer func() {
		if err := recover(); err != nil {
			logger.Warn(fmt.Sprintf("error occurs: %v\n%s", err, string(debug.Stack())))
			result = &reply.UnknowErrReply{}
		}
	}()
	// 获取指令名称
	cmdName := strings.ToLower(string(cmdLine[0]))
	cmdFunc, ok := router[cmdName]
	if !ok {
		return reply.MakeErrReply("ERR unknown command '" + cmdName + "', or not supported in cluster mode")
	}
	// 调用对应cluster层的方法
	result = cmdFunc(c, client, cmdLine)
	return
}

func (c *ClusterDatabase) Close() {
	c.db.Close()
}

func (c *ClusterDatabase) AfterClientClose(client resp.Connection) {
	c.db.AfterClientClose(client)
}
