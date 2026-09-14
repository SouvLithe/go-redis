package connection

import (
	"go-redis/lib/sync/wait"
	"net"
	"sync"
	"time"
)

/*
协议层对每一个连接上的客户端的描述
*/
type Connection struct {
	conn         net.Conn
	waitingReply wait.Wait  // 客户端回发结果时，关闭server前要把没有处理完的请求处理完
	mu           sync.Mutex // 操作客户时上锁，避免并发问题
	selectedDB   int        // 可供选择的库数量
}

func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		conn: conn,
	}
}

// 获取连接远程客户端地址
func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Connection) Close() error {
	// 设置超时时间，避免传输数据过程中被掐断
	c.waitingReply.WaitWithTimeout(10 * time.Second)
	_ = c.conn.Close()
	return nil
}

// 给客户端发送数据
func (c *Connection) Write(bytes []byte) error {
	if len(bytes) == 0 {
		return nil
	}
	c.mu.Lock()
	c.waitingReply.Add(1)
	defer func() {
		c.waitingReply.Done()
		c.mu.Unlock()
	}()
	_, err := c.conn.Write(bytes)
	return err
}

func (c *Connection) GetDBIndex() int {
	return c.selectedDB
}

func (c *Connection) SelectDB(dbNum int) {
	c.selectedDB = dbNum
}
