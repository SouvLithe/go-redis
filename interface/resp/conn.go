package resp

type Connection interface {
	// 给客户端恢复消息
	Write([]byte) error
	// 数据库总量
	GetDBIndex() int
	// 标识正在使用的数据库
	SelectDB(int)
}
