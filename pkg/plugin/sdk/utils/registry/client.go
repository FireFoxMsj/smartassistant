package registry

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
)

const (
	registerTTL = 10

	etcdURL = "http://0.0.0.0:2379"

	managerTarget = "/sa/plugins"
)

func endpointsTarget(service string) string {
	return fmt.Sprintf("%s/%s", managerTarget, service)
}

// RegisterService 注册服务
func RegisterService(ctx context.Context, service, addr string) (err error) {
	logrus.Infoln("register service:", service, addr)
	cli, err := clientv3.NewFromURL(etcdURL)
	if err != nil {
		return
	}
	em, err := endpoints.NewManager(cli, managerTarget)
	if err != nil {
		return
	}

	lease := clientv3.NewLease(cli)
	resp, err := lease.Grant(ctx, registerTTL)
	if err != nil {
		return
	}
	kl, err := lease.KeepAlive(ctx, resp.ID)
	if err != nil {
		return
	}
	go func() {
		for {
			if _, ok := <-kl; !ok {
				return
			}
		}
	}()

	return em.AddEndpoint(ctx,
		endpointsTarget(service),
		endpoints.Endpoint{Addr: addr},
		clientv3.WithLease(resp.ID),
	)
}

// UnregisterService 取消注册服务
func UnregisterService(service string) (err error) {
	logrus.Infoln("unregister service:", service)
	cli, err := clientv3.NewFromURL(etcdURL)
	if err != nil {
		return
	}
	em, err := endpoints.NewManager(cli, managerTarget)
	if err != nil {
		return
	}

	return em.DeleteEndpoint(context.TODO(), endpointsTarget(service))
}
