package main

import (
	"context"
	"flag"
	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/internal/api"
	"github.com/zhiting-tech/smartassistant/internal/cloud"
	"github.com/zhiting-tech/smartassistant/internal/config"
	"github.com/zhiting-tech/smartassistant/internal/plugin"
	"github.com/zhiting-tech/smartassistant/internal/task"
	"github.com/zhiting-tech/smartassistant/internal/websocket"
	"github.com/zhiting-tech/smartassistant/pkg/reverseproxy"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var configFile = flag.String("c", "app.yaml", "config file")

func main() {
	flag.Parse()
	conf := config.InitConfig(*configFile)
	ctx, cancel := context.WithCancel(context.Background())
	if conf.Debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	// 优先使用单例模式，循环引用通过依赖注入解耦
	taskManager := task.GetManager()
	pluginManager := plugin.GetManager()
	wsServer := websocket.NewWebSocketServer()
	pluginManager.SetStateChangeCB(wsServer.OnDeviceStateChange)
	httpServer := api.NewHttpServer(wsServer.AcceptWebSocket)

	go wsServer.Run(ctx)
	go pluginManager.Run(ctx)
	go httpServer.Run(ctx)

	// 等待其他服务启动完成
	time.Sleep(3 * time.Second)
	go taskManager.Run(ctx)

	reverseproxy.RegisterUpstream("wangpan", "wangpan:8089")
	// 如果已配置，则尝试连接 SmartCloud
	if len(conf.SmartCloud.Domain) > 0 {
		go cloud.StartTunnel(ctx)
	}

	logrus.Info("SmartAssistant started")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sig:
		// Exit by user
	}
	logrus.Info("shutting down.")
	cancel()
	time.Sleep(3 * time.Second)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
