package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/pkg/regex"
)

// ContainerIsRunningByImage 返回是否有镜像对应的容器在运行
func (c *Client) ContainerIsRunningByImage(image string) (isRunning bool, err error) {

	ctx := context.Background()
	var containers []types.Container
	containers, err = c.dockerClient.ContainerList(ctx, types.ContainerListOptions{})
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
	hostConf := container.HostConfig{
		NetworkMode: "host",
		AutoRemove:  true,
	}
	logrus.Info("create container ", image.RefStr())
	r, err := c.dockerClient.ContainerCreate(ctx, &config, &hostConf,
		nil, nil, regex.ToSnakeCase(image.Name))
	if err != nil {
		logrus.Error("ContainerCreateErr", err)
		return
	}
	logrus.Info("start container ", r.ID)
	containerID = r.ID
	err = c.dockerClient.ContainerStart(ctx, r.ID, types.ContainerStartOptions{})
	if err != nil {
		logrus.Error("ContainerStart", err)
		return
	}
	return
}

// ContainerStopByImage 停止并删除容器，同时删除镜像
func (c *Client) ContainerStopByImage(image string) (err error) {

	ctx := context.Background()
	var containers []types.Container
	containers, err = c.dockerClient.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return
	}

	for _, con := range containers {
		if con.Image == image {
			logrus.Debug("stop container and remove image", image)
			err = c.dockerClient.ContainerStop(ctx, con.ID, nil)
			if err != nil {
				return
			}
			logrus.Debug("stop container", con.ImageID)
			err = c.dockerClient.ContainerRemove(ctx, con.ID, types.ContainerRemoveOptions{Force: true})
			if err != nil {
				return
			}
			_, err = c.dockerClient.ImageRemove(ctx, con.ImageID, types.ImageRemoveOptions{})
			if err != nil {
				return
			}
			logrus.Debug("remove image")
		}
	}
	return nil
}
