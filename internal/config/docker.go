package config

type Docker struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	Server   string `json:"server" yaml:"server"`
}
