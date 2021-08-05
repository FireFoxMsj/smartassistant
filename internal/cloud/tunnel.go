package cloud

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/internal/config"
	"github.com/zhiting-tech/smartassistant/pkg/proxy"
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
			logrus.Warning("inletsClient connect err:", err)
			sleepTime = sleepTime * 2
			if sleepTime > 600 {
				sleepTime = 600
			}
		} else {
			sleepTime = 2
		}
		time.Sleep(time.Duration(sleepTime) * time.Second)
		logrus.Info("try reconnect...")
	}
}
