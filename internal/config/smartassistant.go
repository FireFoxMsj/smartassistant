package config

import (
	"fmt"
)

type SmartAssistant struct {
	ID       string `json:"id" yaml:"id"`
	Key      string `json:"key" yaml:"key"`
	Db       string `json:"db" yaml:"db"`
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	GRPCPort int    `json:"grpc_port" yaml:"grpc_port"`
}

func (sa SmartAssistant) HttpAddress() string {
	return fmt.Sprintf("%s:%d", sa.Host, sa.Port)
}

func (sa SmartAssistant) GRPCAddress() string {
	return fmt.Sprintf("%s:%d", sa.Host, sa.GRPCPort)
}
