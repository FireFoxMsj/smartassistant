package main

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/zhiting-tech/smartassistant/modules/supervisor/proto"
)

type Server struct {
	dockerClient *client.Client
	proto.UnimplementedSupervisorServer
}

func newServer() *Server {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv,
		client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}
	return &Server{
		dockerClient: dockerClient,
	}
}

// Restart 重启容器，可能需要切换到新版本镜像
func (s *Server) Restart(_ctx context.Context, req *proto.RestartReq) (*proto.Response, error) {
	ctx := context.Background()
	logrus.Infof("restart %v to %v", req.Image, req.NewImage)
	var (
		resp proto.Response
		err  error
	)
	ID := s.getContainerByImage(req.Image)
	if len(ID) == 0 {
		logrus.Warnf("container not found")
		return &resp, err
	}
	conf, err := s.dockerClient.ContainerInspect(context.Background(), ID)
	if err != nil {
		return &resp, err
	}
	if req.Image == req.NewImage {
		logrus.Infof("restart %v", req.Image)
		err = s.dockerClient.ContainerRestart(ctx, ID, nil)
		return &resp, err
	}
	logrus.Infof("updating %v, %v", req.Image, req.NewImage)
	// 更改镜像版本
	s.dockerClient.ContainerStop(ctx, ID, nil)
	s.dockerClient.ContainerRemove(ctx, ID, types.ContainerRemoveOptions{})
	conf.Config.Image = req.NewImage
	r, err := s.dockerClient.ContainerCreate(ctx,
		conf.Config,
		conf.HostConfig,
		nil,
		nil,
		conf.Name,
	)
	if err != nil {
		logrus.Warnf("create new container error %v", err)
		return &resp, err
	}
	err = s.dockerClient.ContainerStart(ctx, r.ID, types.ContainerStartOptions{})
	if err != nil {
		logrus.Warnf("start new container error %v", err)
	}
	return &resp, err
}

func (s *Server) getContainerByImage(image string) (id string) {
	ctx := context.Background()
	var containers []types.Container
	containers, err := s.dockerClient.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		return
	}
	for _, con := range containers {
		logrus.Infof("found container %v", con.Image)
		if con.Image == image {
			return con.ID
		}
	}
	return
}
