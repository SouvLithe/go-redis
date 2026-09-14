package main

import (
	"fmt"
	"go-redis/config"
	"go-redis/lib/logger"
	"go-redis/resp/handler"
	"go-redis/tcp"
	"os"
)

// 设置配置文件
const configFile string = "redis.conf"

// 没有配置文件，默认的配置
var defaultProperties = &config.ServerProperties{
	Bind: "0.0.0.0",
	Port: 6379,
}

// 判断配置文件在不在
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	return err == nil && !info.IsDir()
}

func main() {
	// 设置日志框架格式
	logger.Setup(&logger.Settings{
		Path:       "logs",
		Name:       "go-redis",
		Ext:        "log",
		TimeFormat: "2006-01-02",
	})

	// 判断配置文件在不在，不在就默认配置
	if fileExists(configFile) {
		config.SetupConfig(configFile)
	} else {
		config.Properties = defaultProperties
	}

	err := tcp.ListenAndServeWithSignal(
		&tcp.Config{
			Address: fmt.Sprintf("%s:%d", config.Properties.Bind, config.Properties.Port),
		},
		handler.MakeHandler())
	if err != nil {
		logger.Error(err)
	}
}
