package sdk

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/proto"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"
	addr2 "github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/utils/addr"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/utils/registry"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"
)

func Run(p *server.Server) error {

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

		select {
		case <-sig:
			devices, _ := p.Manager.Devices()
			for _, d := range devices {
				d.Close()
			}
			cancel()
		}
	}()

	if err := runServer(ctx, p); err != nil {
		logrus.Error(err)
	}
	logrus.Info("shutting down.")
	return nil
}

// mixHandler 同时处理http和grpc请求
func mixHandler(mux *http.ServeMux, grpcServer *grpc.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor != 2 {
			mux.ServeHTTP(w, r)
			return
		}
		if strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
			return
		}
		return
	}
}

func runServer(ctx context.Context, p *server.Server) error {

	ln, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return err
	}

	go func() {
		select {
		case <-ctx.Done():
			ln.Close()
		}
	}()

	localIP, err := addr2.LocalIP()
	if err != nil {
		return err
	}
	addr := net.TCPAddr{
		IP:   net.ParseIP(localIP),
		Port: ln.Addr().(*net.TCPAddr).Port,
	}
	// 往etcd注册服务
	if err := registry.RegisterService(ctx, p.Domain, addr.String()); err != nil {
		return err
	}
	defer registry.UnregisterService(p.Domain)

	// grpc服务
	grpcServer := grpc.NewServer()
	proto.RegisterPluginServer(grpcServer, p)

	// http服务
	mux := http.NewServeMux()
	mux.Handle("/", p.Router)

	// h2c实现了不用tls的http/2
	h1s := http.Server{
		Handler: h2c.NewHandler(mixHandler(mux, grpcServer), &http2.Server{}),
	}
	if err = h1s.Serve(ln); err != nil {
		logrus.Error(err)
	}
	return nil
}
