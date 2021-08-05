package main

import (
	"log"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/examples/plugin-demo/device"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"
)

func main() {

	logrus.SetLevel(logrus.DebugLevel)
	p := server.NewPluginServer("demo")

	go func() {
		time.Sleep(1 * time.Second)
		d1 := device.NewDemo("abcdefg")
		if err := p.Manager.AddDevice(d1); err != nil {
			log.Panicln(err)
		}
		d2 := device.NewDemo("hijklmn")
		if err := p.Manager.AddDevice(d2); err != nil {
			log.Panicln(err)
		}
	}()
	p.HtmlRouter.Static("", "./html")

	err := sdk.Run(p)
	if err != nil {
		log.Panicln(err)
	}

}
