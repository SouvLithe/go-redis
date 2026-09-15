package handler

import (
	"context"
	"go-redis/database"
	"go-redis/lib/logger"
	"go-redis/resp/connection"
	"go-redis/resp/parser"
	"go-redis/resp/reply"
	"io"
	"net"
	"strings"

	// 别名 databaseface
	databaseface "go-redis/interface/database"
	"go-redis/lib/sync/atomic"
	"sync"
)

var (
	unknownErrReplyBytes = []byte("-ERR unknown\r\n")
)

type RespHandler struct {
	activeConn sync.Map
	db         databaseface.Database
	closing    atomic.Boolean
}

func MakeHandler() *RespHandler {
	var db databaseface.Database
	//实现Database 这里更改db调用指令
	db = database.NewDatabase()
	return &RespHandler{
		db: db,
	}
}

// 关闭其中一个客户端连接
func (h *RespHandler) closeClient(client *connection.Connection) {
	_ = client.Close()
	h.db.AfterClientClose(client)
	h.activeConn.Delete(client)
}

// 处理客户端链接，每个链接用一个纤程对应
// ctx 是上层传下来的上下文，conn 是这条 TCP 连接。
func (h *RespHandler) Handle(ctx context.Context, conn net.Conn) {
	// 将在关闭过程中的关闭请求，给关掉
	if h.closing.Get() {
		_ = conn.Close()
	}
	client := connection.NewConnection(conn)

	// 把这个连接登记到 activeConn 集合里。
	h.activeConn.Store(client, struct{}{})

	// 启动协议解析器，返回一个 channel
	ch := parser.ParseStream(conn)

	// 从 channel 里一条条取解析结果，直到 channel 被关闭（连接结束）。
	for payload := range ch {
		// error
		if payload.Err != nil {
			// 客户端发送关闭请求
			if payload.Err == io.EOF ||
				payload.Err == io.ErrUnexpectedEOF ||
				strings.Contains(payload.Err.Error(), "use of closed network connection") {
				h.closeClient(client)
				logger.Info("connection closed: " + client.RemoteAddr().String())
				return
			}
			// 协议出问题
			errReply := reply.MakeErrReply(payload.Err.Error())
			err := client.Write(errReply.ToBytes())
			if err != nil {
				h.closeClient(client)
				logger.Info("connection closed: " + client.RemoteAddr().String())
				return
			}
			// 成功处理错误并告诉用户，持续监听用户请求
			continue
		}
		// exec
		if payload.Data == nil {
			logger.Error("empty payload")
			continue
		}
		r, ok := payload.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error("require multi bulk reply")
			continue
		}
		execResult := h.db.Exec(client, r.Args)
		if execResult != nil {
			_ = client.Write(execResult.ToBytes())
		} else {
			_ = client.Write(unknownErrReplyBytes)
		}
	}
}

// 关闭整个Redis，关掉所有client
func (h *RespHandler) Close() error {
	logger.Info("handler shutting down")
	h.closing.Set(true)
	h.activeConn.Range(
		func(key interface{}, value interface{}) bool {
			client := key.(*connection.Connection)
			_ = client.Close()
			return true
		})
	h.db.Close()
	return nil
}
