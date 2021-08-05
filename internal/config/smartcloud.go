package config

import "fmt"

type SmartCloud struct {
	Domain string `json:"domain" yaml:"domain"`
	TLS    bool   `json:"tls" yaml:"tls"`
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
