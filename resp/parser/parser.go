package parser

import (
	"bufio"
	"errors"
	"go-redis/interface/resp"
	"go-redis/lib/logger"
	"go-redis/resp/reply"
	"io"
	"runtime/debug"
	"strconv"
	"strings"
)

/*
	将用户请求，解析成可识别的内容
*/

type Payload struct {
	Data resp.Reply
	Err  error
}

type readState struct {
	readingMultiLine  bool     // 解析单行还是多行数据
	expectedArgsCount int      // 读取指令应该有几个参数
	msgType           byte     // 记录指令类型
	args              [][]byte // 用户传来具体数据
	bulkLen           int64    // 字节组数据块的长度
}

// 判断解析是否完成
func (s *readState) finished() bool {
	return s.expectedArgsCount > 0 && len(s.args) == s.expectedArgsCount
}

// 业务处理和协议解析并发处理，大写的作为对外暴露的接口
// 解析器是每有一个用户，一个解析器;用户断开之后关掉解析器
func ParseStream(reader io.Reader) <-chan *Payload {
	ch := make(chan *Payload)
	go parse0(reader, ch)
	return ch
}

// 核心方法，调用下面的解析方法，完成用户的请求解析
func parse0(reader io.Reader, ch chan<- *Payload) {
	// 出现panic时可以退出来，并打印栈信息到日志
	defer func() {
		if err := recover(); err != nil {
			logger.Error(string(debug.Stack()))
		}
	}()
	bufReader := bufio.NewReader(reader)
	var state readState
	var err error
	var msg []byte
	for true {
		var ioErr bool
		msg, ioErr, err = readLine(bufReader, &state)
		if err != nil {
			if ioErr {
				// 如果解析io错误时，结束服务，将错误放到输出管道里并返回
				ch <- &Payload{Err: err}
				close(ch)
				return
			}
			// 如果只是协议错误，不结束服务，上报错误并继续监听
			ch <- &Payload{Err: err}
			state = readState{}
			continue
		}
		// 判断是否是多行解析模式
		if !state.readingMultiLine {
			if msg[0] == '*' {
				// *3\r\n
				err := parseMultiBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{Err: errors.New("protocol error: " + string(msg))}
					state = readState{}
					continue
				}
				if state.expectedArgsCount == 0 {
					ch <- &Payload{
						Data: &reply.EmptyMultiBulkReply{},
					}
					state = readState{}
					continue
				}
			} else if msg[0] == '$' {
				// $3\r\n $4\r\nPING\r\n $-1\r\n
				err := parseBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{Err: errors.New("protocol error: " + string(msg))}
					state = readState{}
					continue
				}
				if state.bulkLen == -1 {
					ch <- &Payload{
						Data: &reply.NullBulkReply{},
					}
					state = readState{}
					continue
				}
			} else {
				// +-*时
				result, err := parseSingleLineReply(msg)
				ch <- &Payload{Data: result, Err: err}
				state = readState{}
				continue
			}
		} else {
			err := readBody(msg, &state)
			if err != nil {
				ch <- &Payload{Err: errors.New("protocol error: " + string(msg))}
				state = readState{}
				continue
			}
			// 判断用户的请求是否完全解析了
			if state.finished() {
				var result resp.Reply
				if state.msgType == '*' {
					result = reply.MakeMultiBulkReply(state.args)
				} else if state.msgType == '$' {
					result = reply.MakeBulkReply(state.args[0])
				}
				ch <- &Payload{Data: result, Err: err}
				state = readState{}
			}
		}
	}
}

// 示例：*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
func readLine(bufReader *bufio.Reader, state *readState) ([]byte, bool, error) {
	// 1.没有读到$数字,直接按照\r\n切分就可以
	var msg []byte
	var err error

	if state.bulkLen == 0 {
		// \r\n切分
		msg, err = bufReader.ReadBytes('\n') // 返回值包含\n
		if err != nil {
			return nil, true, err
		}
		if len(msg) < 2 || msg[len(msg)-2] != '\r' {
			return nil, false, errors.New("protocol error：" + string(msg))
		}
	} else {
		// 2.读到$数字,严格读取字符个数
		msg = make([]byte, state.bulkLen+2)
		_, err := io.ReadFull(bufReader, msg)
		if err != nil {
			return nil, true, err
		}
		if len(msg) == 0 || msg[len(msg)-2] != '\r' || msg[len(msg)-1] != '\n' {
			return nil, false, errors.New("protocol error：" + string(msg))
		}
		state.bulkLen = 0
	}
	return msg, false, nil
}

// 更新结构体属性
// *2\r\n$6\r\nselect\r\n$1\r\n1\r\n
// 示例：*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
func parseMultiBulkHeader(msg []byte, state *readState) error {
	var err error
	var expectedLine uint64
	expectedLine, err = strconv.ParseUint(string(msg[1:len(msg)-2]), 10, 32)
	if err != nil {
		return errors.New("protocol error：" + string(msg))
	}
	if expectedLine == 0 {
		state.expectedArgsCount = 0
		return nil
	} else if expectedLine > 0 {
		state.msgType = msg[0]
		state.readingMultiLine = true
		state.expectedArgsCount = int(expectedLine)
		state.args = make([][]byte, 0, expectedLine)
		return nil
	} else {
		return errors.New("protocol error：" + string(msg))
	}
}

// 解析 $4\r\nPING\r\n
func parseBulkHeader(msg []byte, state *readState) error {
	var err error
	state.bulkLen, err = strconv.ParseInt(string(msg[1:len(msg)-2]), 10, 64)
	if err != nil {
		return errors.New("protocol error:" + string(msg))
	}
	if state.bulkLen == -1 {
		return nil
	} else if state.bulkLen > 0 {
		state.msgType = msg[0]
		state.readingMultiLine = true
		state.expectedArgsCount = 1
		state.args = make([][]byte, 0, 1)
		return nil
	} else {
		return errors.New("protocol error:" + string(msg))
	}
}

// 解析 +OK\r\n -err\r\n :5\r\n
func parseSingleLineReply(msg []byte) (resp.Reply, error) {
	str := strings.TrimSuffix(string(msg), "\r\n")
	var result resp.Reply
	switch msg[0] {
	case '+':
		result = reply.MakeStatusReply(str[1:])
	case '-':
		result = reply.MakeErrReply(str[1:])
	case ':':
		val, err := strconv.ParseInt(str[1:], 10, 64)
		if err != nil {
			return nil, errors.New("protocol error:" + string(msg))
		}
		result = reply.MakeIntReply(val)
	}
	return result, nil
}

// 解析指令体
// 情况1：*3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
// 情况2：$4\r\nPING\r\n 或 PING\r\n
func readBody(msg []byte, state *readState) error {
	// 切掉\r\n
	line := msg[0 : len(msg)-2]
	var err error

	if line[0] == '$' {
		state.bulkLen, err = strconv.ParseInt(string(line[1:]), 10, 64)
		// $3
		if err != nil {
			return errors.New("protocol error:" + string(msg))
		}
		// $0\r\n
		if state.bulkLen <= 0 {
			state.args = append(state.args, []byte{})
			state.bulkLen = 0
		}
	} else {
		state.args = append(state.args, line)
	}
	return nil
}
