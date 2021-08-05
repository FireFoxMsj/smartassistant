package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/internal/api/middleware"
	"github.com/zhiting-tech/smartassistant/internal/config"
	"net/http"
	"time"
)

type HttpServer struct {
	addr      string
	ginEngine *gin.Engine
}

func NewHttpServer(ws gin.HandlerFunc) *HttpServer {
	conf := config.GetConf()
	r := gin.Default()
	r.GET("/ws", middleware.RequireToken, ws)
	loadModules(r)
	r.Static(fmt.Sprintf("static/%s", conf.SmartAssistant.ID), "./static")

	return &HttpServer{
		ginEngine: r,
		addr:      conf.SmartAssistant.HttpAddress(),
	}
}

func (s *HttpServer) Run(ctx context.Context) {
	logrus.Infof("starting http server on %s", s.addr)
	srv := &http.Server{
		Addr:    s.addr,
		Handler: s.ginEngine,
	}

	// 启动http服务
	go func() {
		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			logrus.Infof("http listen: %s\n", err)
		}
	}()
	<-ctx.Done()

	stop, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = srv.Shutdown(stop)
	logrus.Warning("http server stopped")
}
