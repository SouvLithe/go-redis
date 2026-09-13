package resp

// 表示各种客户端对服务端的回复
type Reply interface {
	// TCP交流使用字节流，将回复的内容转成字节
	ToBytes() []byte
}
