package tcp

import (
	"context"
	"go-redis/interface/tcp"
	"go-redis/lib/logger"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// 启动tcp server的地址，加最大连接数、超时等，在这里加
type Config struct {
	Address string
}

func ListenAndServeWithSignal(
	cfg *Config,
	handler tcp.Handler) error {

	closeChan := make(chan struct{})
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		signal := <-sigChan
		switch signal {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()

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

	go func() {
		// 如果channel没有数据进来，要读channel缓冲区的话，就会卡在这
		// 受到信号时，关掉listener和handler
		<-closeChan
		logger.Info("shutting down")
		_ = listener.Close()
		_ = handler.Close()
	}()

	defer func() {
		// 退出时，关掉listener和handler
		// 如果返回值没有处理就像这样扔掉
		_ = listener.Close()
		_ = handler.Close()
	}()

	ctx := context.Background()
	var waitDone sync.WaitGroup
	for true {
		conn, err := listener.Accept()
		if err != nil {
			break
		}

		logger.Info("Accepted link")
		waitDone.Add(1)

		// 一个协程对应一个连接
		go func() {
			// 如果在handler出现panic，会导致走不到Done
			// 所以写进defer而不是写在下面
			defer func() {
				// 完成后等待组减一
				waitDone.Done()
			}()
			handler.Handle(ctx, conn)
		}()
	}
	waitDone.Wait()
}
