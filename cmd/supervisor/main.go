package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/modules/supervisor/proto"
	"google.golang.org/grpc"
)

var (
	Version = "latest"
)

func main() {
	logrus.Infof("starting supervisor %v", Version)
	server := newServer()
	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		logrus.Errorf("failed to listen: %v", err)
		os.Exit(-1)
	}
	grpcServer := grpc.NewServer()
	proto.RegisterSupervisorServer(grpcServer, server)
	result := make(chan struct{}, 1)
	go func() {
		err = grpcServer.Serve(lis)
		if err != nil {
			logrus.Error(err)
			result <- struct{}{}
			return
		}
	}()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sig:
		// Exit by user
		grpcServer.Stop()
		time.Sleep(100 * time.Microsecond)
	case <-result:
	}
	logrus.Info("shutting down.")
}
