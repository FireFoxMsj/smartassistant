package docker

import (
	"context"
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/zhiting-tech/smartassistant/pkg/logger"
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

// ContainerRun 根据镜像创建容器并运行
func (c *Client) ContainerRun(image string, conf container.Config, hostConf container.HostConfig) (containerID string, err error) {
	ctx := context.Background()

	logger.Info("create container ", image)
	r, err := c.DockerClient.ContainerCreate(ctx, &conf, &hostConf,
		nil, nil, regex.ToSnakeCase(image))
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

// ContainerStopByImage 停止并删除容器 TODO 优化
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
			return
		}
	}
	return
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
