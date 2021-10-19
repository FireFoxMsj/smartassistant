package config

type Options struct {
	Debug          bool           `json:"debug" yaml:"debug"`
	SmartCloud     SmartCloud     `json:"smartcloud" yaml:"smartcloud"`
	SmartAssistant SmartAssistant `json:"smartassistant" yaml:"smartassistant"`
	Docker         Docker         `json:"docker" yaml:"docker"`
	Datatunnel     Datatunnel     `json:"datatunnel" yaml:"datatunnel"`
}
