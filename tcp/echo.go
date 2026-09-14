package tcp

import (
	"bufio"
	"context"
	"go-redis/lib/logger"
	"go-redis/lib/sync/atomic"
	"go-redis/lib/sync/wait"
	"io"
	"net"
	"sync"
	"time"
)

// 记录所有连接客户端的信息，然后对所有客户端做的服务，就是：你发我什么，我回你什么
// 测试用的Handler
// 客户端实体
type EchoClient struct {
	Conn net.Conn
	// go原生的 waitgroup 相互等待时是没有超时的
	Waiting wait.Wait
}

func (e *EchoClient) Close() error {
	e.Waiting.WaitWithTimeout(10 * time.Second)
	_ = e.Conn.Close()
	return nil
}

// 回复
type EchoHandler struct {
	activeConn sync.Map
	// 这里使用实现的原子布尔
	closing atomic.Boolean
}

func MakeHandler() *EchoHandler {
	return &EchoHandler{}
}

func (handler *EchoHandler) Handle(ctx context.Context, conn net.Conn) {
	// 在关闭过程中的关闭请求，给关掉
	if handler.closing.Get() {
		_ = conn.Close()
	}
	// 将业务连接包装成client结构体
	client := &EchoClient{
		Conn: conn,
	}
	// 记录下来，存入所有正在服务的客户端
	handler.activeConn.Store(client, conn)
	reader := bufio.NewReader(conn)
	for {
		// 根据用户一换行，我就收到一个msg
		msg, err := reader.ReadString('\n')
		if err != nil {
			// EOF是OS中数据的结束符
			if err != io.EOF {
				logger.Info("Connecting close")
				handler.activeConn.Delete(client)
			} else {
				logger.Warn(err)
			}
			return
		}
		// 将发的东西，在waitgroup记录一下（不要关掉），转换为字节写回去
		client.Waiting.Add(1)
		b := []byte(msg)
		_, _ = conn.Write(b)
		client.Waiting.Done()
	}
}

func (handler *EchoHandler) Close() error {
	logger.Info("handler shutting down")
	// 把业务状态改成正在关闭
	handler.closing.Set(true)
	// 将内部方法施加到map上的每一个kv，进行关闭连接操作
	// 内部bool的作用是：是否遍历下一个key
	handler.activeConn.Range(func(key, value interface{}) bool {
		client := key.(*EchoClient)
		_ = client.Conn.Close()
		// true遍历完，false只做第一个
		return true
	})
	return nil
}
