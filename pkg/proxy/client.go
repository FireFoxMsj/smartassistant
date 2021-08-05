// Package proxy 数据转发通道
package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/inlets/inlets/pkg/transport"
	"github.com/rancher/remotedialer"
	"github.com/twinj/uuid"
)

// Client for inlets
type Client struct {
	// Remote site for websocket address
	Remote string

	// Map of upstream servers dns.entry=http://ip:port
	UpstreamMap map[string]string

	// Token for authentication
	Token string

	// StrictForwarding
	StrictForwarding bool
}

func makeAllowsAllFilter() func(network, address string) bool {
	return func(network, address string) bool {
		return true
	}
}

func makeFilter(upstreamMap map[string]string) func(network, address string) bool {

	trimmedMap := map[string]bool{}

	for _, v := range upstreamMap {
		u, err := url.Parse(v)
		if err != nil {
			log.Printf("Error parsing: %s, skipping.\n", v)
			continue
		}

		trimmedMap[u.Host] = true
	}

	return func(network, address string) bool {
		if network != "tcp" {
			log.Printf("network not allowed: %q\n", network)

			return false
		}

		if ok, v := trimmedMap[address]; ok && v {
			return true
		}

		return false
	}
}

// Connect connect and serve traffic through websocket
func (c *Client) Connect(ctx context.Context, saID string) error {
	headers := http.Header{}
	headers.Set("SA-ID", saID)
	headers.Set(transport.InletsHeader, uuid.Formatter(uuid.NewV4(), uuid.FormatHex))
	for k, v := range c.UpstreamMap {
		headers.Add(transport.UpstreamHeader, fmt.Sprintf("%s=%s", k, v))
	}
	if c.Token != "" {
		headers.Add("Authorization", "Bearer "+c.Token)
	}

	u := c.Remote
	if !strings.HasPrefix(u, "ws") {
		u = "ws://" + u
	}
	var filter func(network, address string) bool

	if c.StrictForwarding {
		filter = makeFilter(c.UpstreamMap)
	} else {
		filter = makeAllowsAllFilter()
	}

	dialer := &websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		Proxy: http.ProxyFromEnvironment}

	return remotedialer.ClientConnect(ctx, u, headers, dialer, filter, nil)
}
