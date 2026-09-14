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
	//TODO:实现Database
	db = database.NewEchoDatabase()
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

func (h *RespHandler) Handle(ctx context.Context, conn net.Conn) {
	// 将在关闭过程中的关闭请求，给关掉
	if h.closing.Get() {
		_ = conn.Close()
	}
	newConnection := connection.NewConnection(conn)
	h.activeConn.Store(newConnection, 1)
	ch := parser.ParseStream(conn)
	for payload := range ch {
		// error
		if payload.Err != nil {
			// 客户端发送关闭请求
			if payload.Err == io.EOF ||
				payload.Err == io.ErrUnexpectedEOF ||
				strings.Contains(payload.Err.Error(), "use of closed network connection") {
				h.closeClient(newConnection)
				logger.Info("connection closed: " + newConnection.RemoteAddr().String())
				return
			}
			// 协议出问题
			errReply := reply.MakeErrReply(payload.Err.Error())
			err := newConnection.Write(errReply.ToBytes())
			if err != nil {
				h.closeClient(newConnection)
				logger.Info("connection closed: " + newConnection.RemoteAddr().String())
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
		execResult := h.db.Exec(newConnection, r.Args)
		if execResult != nil {
			_ = newConnection.Write(r.ToBytes())
		} else {
			_ = newConnection.Write(unknownErrReplyBytes)
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
