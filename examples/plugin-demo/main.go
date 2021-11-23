package main

import (
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"
	"plugin-demo/device"
)

func main() {

	logrus.SetLevel(logrus.DebugLevel)
	p := server.NewPluginServer()

	go func() {
		time.Sleep(1 * time.Second)
		d1 := device.NewLightBulbDevice("abcdefg", "lamp9")
		if err := p.Manager.AddDevice(d1); err != nil {
			logger.Panicln(err)
		}
		d2 := device.NewLightBulbDevice("hijklmn", "ceiling17")
		if err := p.Manager.AddDevice(d2); err != nil {
			logger.Panicln(err)
		}
	}()

	err := sdk.Run(p)
	if err != nil {
		logger.Panicln(err)
	}

}
