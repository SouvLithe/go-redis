package reply

// 保存一些固定的回复
// 1. 回复Ping
type PongReply struct {
}

var pongbytes = []byte("+PONG\r\n")

func (r PongReply) ToBytes() []byte {
	return pongbytes
}

// 上面的其实已经可以了，但是很多还是习惯暴露一个make方法
// 外面需要使用pongreply，就不用new一个结构体，可以直接调用这个make方法
// 好处就是避免裸函数，可将结构体变为私有
func MakePongReply() *PongReply {
	return &PongReply{}
}

// 2. 回复OK
type OKReply struct {
}

var okbytes = []byte("+OK\r\n")

func (r OKReply) ToBytes() []byte {
	return okbytes
}

// 所有调用复用同一个全局对象，单例
// 而上面那种是，每次调用语义上新建一个对象
var theOKReply = new(OKReply)

func MakeOKReply() *OKReply {
	return theOKReply
}

// 3. 空字符串回复
type NullBulkReply struct {
}

var nullBulkBytes = []byte("$-1\r\n")

func (n NullBulkReply) ToBytes() []byte {
	return nullBulkBytes
}

func MakeNullBulkReply() *NullBulkReply {
	return &NullBulkReply{}
}

// 4. 空数组回复
type EmptyMultiBulkReply struct {
}

var emptyMultiBulkBytes = []byte("*0\r\n")

func (e EmptyMultiBulkReply) ToBytes() []byte {
	return emptyMultiBulkBytes
}

// 5. 真的空回复
type NoReply struct {
}

var noBytes = []byte("")

func (n NoReply) ToBytes() []byte {
	return noBytes
}
