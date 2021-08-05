package docker

import (
	"context"
	"fmt"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/sirupsen/logrus"
)

// Image 镜像
type Image struct {
	Name     string `json:"name"`
	Tag      string `json:"tag"`      // 镜像标签
	Registry string `json:"registry"` // 仓库地址
}

func (i Image) RefStr() string {
	tag := i.Tag
	if tag == "" {
		tag = "latest"
	}
	return fmt.Sprintf("%s:%s", i.Repository(), tag)
}

func (i Image) Repository() string {
	return fmt.Sprintf("%s/%s", i.Registry, i.Name)
}

// IsImageNewest TODO 镜像是否最新
func (c *Client) IsImageNewest() (isNewest bool, err error) {
	return false, nil
}

// IsImageAdd 镜像是否已经拉取到本地
func (c *Client) IsImageAdd(refStr string) (isAdded bool) {

	ctx := context.Background()
	inspect, b, err := c.dockerClient.ImageInspectWithRaw(ctx, refStr)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(inspect)
	log.Println(string(b))
	return true
}

// Pull 拉取镜像
func (c *Client) Pull(refStr string) (err error) {
	logrus.Info("pulling image: ", refStr)
	ctx := context.Background()
	_, err = c.dockerClient.ImagePull(ctx, refStr, types.ImagePullOptions{RegistryAuth: c.authStr})
	if err != nil {
		logrus.Warning(err)
		return
	}
	logrus.Info("pull success")
	return
}
