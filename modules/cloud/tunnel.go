package cloud

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zhiting-tech/smartassistant/modules/config"
	"github.com/zhiting-tech/smartassistant/pkg/datatunnel/proto"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/proxy"
	"google.golang.org/grpc"
)

func StartTunnel(ctx context.Context) {
	conf := config.GetConf()
	upstreamMap := map[string]string{
		conf.SmartCloud.Domain: conf.SmartAssistant.HttpAddress(),
	}
	remote := fmt.Sprintf("%s/tunnel", conf.SmartCloud.WebsocketURL())

	inletsClient := proxy.Client{
		Remote:      remote,
		UpstreamMap: upstreamMap,
		Token:       conf.SmartAssistant.Key,
	}

	saID := conf.SmartAssistant.ID
	sleepTime := 2
	for {
		if err := inletsClient.Connect(ctx, saID); err != nil {
			logger.Warning("inletsClient connect err:", err)
			sleepTime = sleepTime * 2
			if sleepTime > 600 {
				sleepTime = 600
			}
		} else {
			sleepTime = 2
		}
		time.Sleep(time.Duration(sleepTime) * time.Second)
		logger.Info("try reconnect...")
	}
}

func StartDataTunnel(ctx context.Context) {
	var level string
	conf := config.GetConf()

	hostname := conf.SmartCloud.Domain
	index := strings.Index(hostname, ":")
	if index > 0 {
		hostname = hostname[:index]
	}

	sleepTime := 2
	target := fmt.Sprintf("%s:%d", hostname, conf.SmartCloud.GRPCPort)
	conn, err := grpc.Dial(target, grpc.WithInsecure())
	if err != nil {
		logger.Warning("grpc connect err:", err)
		return
	}

	if conf.Debug {
		level = "debug"
	}
	sleepTime = 2
	grpcClient := proto.NewDatatunnelControllerClient(conn)
	streamClient := &ControlStreamClient{
		SaID:     conf.SmartAssistant.ID,
		Key:      conf.SmartAssistant.Key,
		LogLevel: level,
	}
	for {

		stream, err := grpcClient.ControlStream(ctx)
		if err != nil {
			logger.Warning("ControlStream err:", err)
			sleepTime = sleepTime * 2
			if sleepTime > 600 {
				sleepTime = 600
			}
		} else {
			sleepTime = 2
		}

		if stream != nil {
			streamClient.HandleStream(stream)
		}

		time.Sleep(time.Duration(sleepTime) * time.Second)
	}
}
