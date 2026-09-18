package aof

import (
	"go-redis/config"
	"go-redis/interface/databaseface"
	"go-redis/lib/logger"
	"go-redis/lib/utils"
	"go-redis/resp/connection"
	"go-redis/resp/parser"
	"go-redis/resp/reply"
	"io"
	"os"
	"strconv"
)

/*
aof操作只有写操作需要，读操作不需要
*/

type CmdLine = [][]byte

const (
	aofQueueSize = 1 << 16
)

// 操作的数据结构
type payload struct {
	cmdLine CmdLine
	dbIndex int
}

type AofHandler struct {
	database    databaseface.Database
	aofChan     chan *payload
	aofFile     *os.File
	aofFileName string
	currentDB   int //表示当前DB，用以避免重复选择指令select
}

// 初始化AofHandler数据结构
func NewAofHandler(database databaseface.Database) (*AofHandler, error) {
	handler := &AofHandler{}
	handler.aofFileName = config.Properties.AppendFilename
	handler.database = database

	// 加载旧有的Aof
	handler.LoadAof()
	// 根据文件名打开文件，如果没有这个文件，那就创建；如果有了，那就打开
	// 这个打开是从头到尾都要用的，所以不用close
	aofile, err := os.OpenFile(handler.aofFileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	handler.aofFile = aofile
	// channel
	handler.aofChan = make(chan *payload, aofQueueSize)
	go func() { handler.handleAof() }()
	return handler, nil
}

// Add payload(set k v) -> aofChan把payload塞到channel里面
func (handler *AofHandler) AddAof(dbIndex int, cmd CmdLine) {
	if config.Properties.AppendOnly && handler.aofChan != nil {
		handler.aofChan <- &payload{
			cmdLine: cmd,
			dbIndex: dbIndex,
		}
	}
}

// handleAof payload(set k v) <- aofChan(落盘)
func (handler *AofHandler) handleAof() {
	handler.currentDB = 0
	for p := range handler.aofChan {
		if p.dbIndex != handler.currentDB {
			data := reply.MakeMultiBulkReply(utils.ToCmdLine("select", strconv.Itoa(p.dbIndex))).ToBytes()
			_, err := handler.aofFile.Write(data)
			if err != nil {
				logger.Error(err)
				continue
			}
			handler.currentDB = p.dbIndex
		}

		data := reply.MakeMultiBulkReply(p.cmdLine).ToBytes()
		_, err := handler.aofFile.Write(data)
		if err != nil {
			logger.Error(err)
			continue
		}
	}
}

// TEST数据； *2\r\n$3\r\nGET\r\n$3\r\nkey\r\n
// *2\r\n$6\r\nSELECT\r\n$1\r\n1\r\n
// *2\r\n$3\r\nGET\r\n$3\r\nkey\r\n
// LoadAof 在启动的时候，从系统加载aof文件
func (handler *AofHandler) LoadAof() {
	file, err := os.Open(handler.aofFileName)
	// 如果有错误直接返回错误，不做加载了
	if err != nil {
		logger.Error(err)
		return
	}
	defer file.Close()
	ch := parser.ParseStream(file)

	fackConn := &connection.Connection{}
	for p := range ch {
		if p.Err != nil {
			// EOF是一个文件的结束符，读到就是文件读完了
			if p.Err == io.EOF {
				break
			}
			logger.Error(p.Err)
			continue
		}
		if p.Data == nil {
			logger.Error("empty payload")
			continue
		}
		// (*reply.MultiBulkReply)叫类型断言 和 强转不同
		// 类型断言尝试去转换，会告诉你失败
		r, ok := p.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error("need multi bulk")
			continue
		}
		rep := handler.database.Exec(fackConn, r.Args)
		if reply.IsErrReply(rep) {
			logger.Error(rep)
		}
	}
}

// TODO: 整合多个同类命令
