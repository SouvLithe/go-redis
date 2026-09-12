package tcp

import (
	"context"
	"go-redis/interface/tcp"
	"go-redis/lib/logger"
	"net"
)

// 启动tcp server的地址，加最大连接数、超时等，在这里加
type Config struct {
	Address string
}

func ListenAndServeWithSignal(
	cfg *Config,
	handler tcp.Handler) error {
	closeChan := make(chan struct{})
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return err
	}
	logger.Info("Start listen")
	ListenAndServe(listener, handler, closeChan)
	return nil
}

func ListenAndServe(listener net.Listener,
	handler tcp.Handler,
	closeChan <-chan struct{}) {
	ctx := context.Background()
	for true {
		conn, err := listener.Accept()
		if err != nil {
			break
		}

		logger.Info("Accepted link")

		// 一个协程对应一个连接
		go func() {
			handler.Handle(ctx, conn)
		}()
	}
}
