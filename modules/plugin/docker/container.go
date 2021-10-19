package docker

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	config2 "github.com/zhiting-tech/smartassistant/modules/config"
	logger "github.com/zhiting-tech/smartassistant/pkg/logger"
	"github.com/zhiting-tech/smartassistant/pkg/regex"
)

// ContainerIsRunningByImage 返回是否有镜像对应的容器在运行
func (c *Client) ContainerIsRunningByImage(image string) (isRunning bool, err error) {

	ctx := context.Background()
	var containers []types.Container
	containers, err = c.DockerClient.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return
	}
	for _, con := range containers {
		if con.Image == image {
			return true, nil
		}
	}
	return false, nil
}

// ContainerRunByImage 根据镜像创建容器并运行
func (c *Client) ContainerRunByImage(image Image) (containerID string, err error) {
	ctx := context.Background()
	config := container.Config{
		Image: image.RefStr(),
	}
	// 映射插件目录到宿主机上
	source := filepath.Join(config2.GetConf().SmartAssistant.HostRuntimePath,
		"data", "plugin", image.Name)
	os.MkdirAll(source, os.ModePerm)
	target := "/app/data/"
	logger.Debugf("mount %s to %s", source, target)

	hostConf := container.HostConfig{
		NetworkMode: "host",
		AutoRemove:  true, // TODO 服务挂了的话日志会丢失
		Mounts: []mount.Mount{
			{Type: mount.TypeBind, Source: source, Target: target},
		},
		// 设置容器的logging driver
		LogConfig: container.LogConfig{
			Type: "fluentd",
			Config: map[string]string{
				"fluentd-address": config2.GetConf().SmartAssistant.FluentdAddress,
				"tag":             fmt.Sprintf("smartassistant.plugin.%s", image.RefStr()),
			},
		},
	}
	logger.Info("create container ", image.RefStr())
	r, err := c.DockerClient.ContainerCreate(ctx, &config, &hostConf,
		nil, nil, regex.ToSnakeCase(image.Name))
	if err != nil {
		logger.Error("ContainerCreateErr", err)
		return
	}
	logger.Info("start container ", r.ID)
	containerID = r.ID
	err = c.DockerClient.ContainerStart(ctx, r.ID, types.ContainerStartOptions{})
	if err != nil {
		logger.Error("ContainerStart", err)
		return
	}
	return
}

// ContainerStopByImage 停止并删除容器
func (c *Client) ContainerStopByImage(image string) (err error) {

	ctx := context.Background()
	var containers []types.Container
	containers, err = c.DockerClient.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return
	}

	for _, con := range containers {
		if con.Image == image {
			logger.Debug("stop container", image)
			err = c.DockerClient.ContainerStop(ctx, con.ID, nil)
			if err != nil {
				return
			}
			logger.Debug("container stop", con.ImageID)
		}
	}
	return nil
}

// ContainerRestartByImage 重启容器
func (c *Client) ContainerRestartByImage(image string) (err error) {

	ctx := context.Background()
	var containers []types.Container
	containers, err = c.DockerClient.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return
	}

	for _, con := range containers {
		if con.Image == image {
			logger.Debug("restart container", image)
			err = c.DockerClient.ContainerRestart(ctx, con.ID, nil)
			if err != nil {
				return
			}
			logger.Debug("container restarted", con.ImageID)
		}
	}
	return nil
}

func (c *Client) GetContainerByImage(image string) (id string, err error) {

	ctx := context.Background()
	var containers []types.Container
	containers, err = c.DockerClient.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return
	}

	for _, con := range containers {
		logger.Infof("container %v, %v", con.ID, con.Image)
		if con.Image == image {
			id = con.ID
			return
		}
	}
	return "", errors.New("not found")
}
