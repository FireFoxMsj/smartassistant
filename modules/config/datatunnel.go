package config

import (
	"fmt"
	"strconv"
)

type Datatunnel struct {
	ExportServices map[string]string `json:"export_services" yaml:"export_services"`
}

func (t *Datatunnel) GetAddr(serviceName string) (addr string, ok bool) {
	var (
		val string
	)
	val, ok = t.ExportServices[serviceName]
	if !ok {
		return
	}

	port, err := strconv.ParseInt(val, 10, 32)
	if err == nil {
		addr = fmt.Sprintf("127.0.0.1:%d", port)
		return
	}

	addr = val
	return
}
