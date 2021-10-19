package config

import "fmt"

type SmartCloud struct {
	Domain       string `json:"domain" yaml:"domain"`
	TLS          bool   `json:"tls" yaml:"tls"`
	GRPCPort     int    `json:"grpc_port" yaml:"grpc_port"`
	DataCenterID int    `json:"data_center_id"`
	WorkerID     int    `json:"worker_id"`
}

func (sc SmartCloud) URL() string {
	scheme := "http"
	if sc.TLS {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s/api", scheme, sc.Domain)
}

func (sc SmartCloud) WebsocketURL() string {
	scheme := "ws"
	if sc.TLS {
		scheme = "wss"
	}
	return fmt.Sprintf("%s://%s/api", scheme, sc.Domain)
}
