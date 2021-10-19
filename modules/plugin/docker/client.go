package docker

import (
	"encoding/base64"
	"encoding/json"
	"sync"

	"github.com/zhiting-tech/smartassistant/modules/config"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var (
	defaultClient *Client
	clientOnce    sync.Once
)

type Client struct {
	DockerClient *client.Client
	authStr      string
}

func GetClient() *Client {
	clientOnce.Do(func() {
		conf := config.GetConf()
		dockerClient, err := client.NewClientWithOpts(client.FromEnv,
			client.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}
		authConfig := types.AuthConfig{
			Username:      conf.Docker.Username,
			Password:      conf.Docker.Password,
			ServerAddress: conf.Docker.Server,
		}
		data, err := json.Marshal(authConfig)
		if err != nil {
			panic(err)
		}
		authStr := base64.URLEncoding.EncodeToString(data)
		defaultClient = &Client{
			DockerClient: dockerClient,
			authStr:      authStr,
		}
	})
	return defaultClient
}
