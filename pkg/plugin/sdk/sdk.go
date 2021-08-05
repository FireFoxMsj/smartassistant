package sdk

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
	"github.com/micro/go-micro/v2/web"
	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/proto"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"
)

const registerTTL = time.Second * 10
const registerInterval = time.Second * 5

func Run(p *server.Server) error {

	r := etcd.NewRegistry(
		registry.Addrs("0.0.0.0:2379"),
		registry.Timeout(10*time.Second),
	)

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		RunGrpcServer(ctx, r, p.Domain, p)
		wg.Done()
	}()
	go func() {
		RunWebServer(ctx, r, p.Domain, p)
		wg.Done()
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sig:
		// Exit by user
		devices, _ := p.Manager.Devices()
		for _, d := range devices {
			d.Close()
		}
		cancel()
		wg.Wait()
	}
	logrus.Info("shutting down.")
	return nil
}

func RunGrpcServer(ctx context.Context, r registry.Registry, domain string, p *server.Server) {

	s := micro.NewService(micro.Context(ctx), micro.Registry(r), micro.Name(domain),
		micro.RegisterTTL(registerTTL), micro.RegisterInterval(registerInterval))
	proto.RegisterPluginHandler(s.Server(), p)
	if err := s.Run(); err != nil {
		log.Println(err)
	}
}

func RunWebServer(ctx context.Context, r registry.Registry, domain string, p *server.Server) {

	ws := web.NewService(web.Context(ctx), web.Registry(r), web.Name(domain+".http"),
		web.RegisterTTL(registerTTL), web.RegisterInterval(registerInterval))
	ws.Handle("/", p.Router)

	if err := ws.Run(); err != nil {
		log.Println(err)
	}
}
