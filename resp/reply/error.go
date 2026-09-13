package reply

// 返回错误
type UnknowErrReply struct {
}

var unknowErrBytes = []byte("-Err unknown\r\n")

func (u UnknowErrReply) Error() string {
	return "Err unknown"
}

func (u UnknowErrReply) ToBytes() []byte {
	return unknowErrBytes
}

// 返回参数异常
type ArgNumErrReply struct {
	cmd string
}

func (r *ArgNumErrReply) Error() string {
	return "-Err wrong number of arguments for '" + r.cmd + "' command\r\n"
}

func (r *ArgNumErrReply) ToBytes() []byte {
	return []byte("-Err wrong number of arguments for '" + r.cmd + "' command\r\n")
}

func MakeArgNumErrReply(cmd string) *ArgNumErrReply {
	return &ArgNumErrReply{cmd: cmd}
}

// 语法异常

type SyntaxErrReply struct {
}

var syntaxErrBytes = []byte("-Err syntax error\r\n")
var theSyntaxErrReply = &SyntaxErrReply{}

func (r *SyntaxErrReply) Error() string {
	return "-Err syntax error"
}

func (r *SyntaxErrReply) ToBytes() []byte {
	return syntaxErrBytes
}

func MakeSyntaxErrReply() *SyntaxErrReply {
	return theSyntaxErrReply
}

// 数据类型错误
type WrongTypeErrReply struct {
}

var wrongTypeErrBytes = []byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n")

func (r *WrongTypeErrReply) Error() string {
	return "-WRONGTYPE Operation against a key holding the wrong kind of value"
}

func (r *WrongTypeErrReply) ToBytes() []byte {
	return wrongTypeErrBytes
}

// 接口协议错误
type ProtocolErrReply struct {
	Msg string
}

func (r *ProtocolErrReply) Error() string {
	return "ERR Protocol error: " + r.Msg
}

func (r *ProtocolErrReply) ToBytes() []byte {
	return []byte("-ERR Protocol '" + r.Msg + "'\r\n")
}
