package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/pkg/archive"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/proto"
	"log"
	"os"
)

type Server struct {
	Manager      *Manager
	Domain       string
	Router       *gin.Engine
	ApiRouter    *gin.RouterGroup
	pluginRouter *gin.RouterGroup
	configFile   string
	staticDir    string
}

func (p Server) HealthCheck(context context.Context, req *proto.HealthCheckReq) (resp *proto.HealthCheckResp, err error) {
	logrus.Debugf("%s HealthCheck", req.Identity)

	resp = &proto.HealthCheckResp{
		Identity: req.Identity,
		Online:   p.Manager.HealthCheck(req.Identity),
	}
	return
}

func (p Server) Discover(request *proto.Empty, server proto.Plugin_DiscoverServer) error {
	devices, _ := p.Manager.Devices()
	for _, device := range devices {
		d := proto.Device{
			Identity:     device.Identity(),
			Model:        device.Info().Model,
			Manufacturer: device.Info().Manufacturer,
		}
		server.Send(&d)
	}
	return nil
}

func (p Server) GetAttributes(context context.Context, request *proto.GetAttributesReq) (resp *proto.GetAttributesResp, err error) {
	logrus.Debugf("%s GetAttributes", request.Identity)

	instances, err := p.Manager.GetAttributes(request.Identity)
	if err != nil {
		return
	}

	resp = new(proto.GetAttributesResp)
	resp.Success = true
	for _, instance := range instances {

		data, _ := json.Marshal(instance.Attributes)
		ins := proto.Instance{
			Type:       instance.Type,
			Identity:   request.Identity,
			InstanceId: int32(instance.InstanceId),
			Attributes: data,
		}
		resp.Instances = append(resp.Instances, &ins)
	}
	log.Println("instances resp:", resp)
	return
}

type Attribute struct {
	ID        int         `json:"id"`
	Attribute string      `json:"attribute"`
	Val       interface{} `json:"val"`
	ValType   string      `json:"val_type"`
	Min       *int        `json:"min,omitempty"`
	Max       *int        `json:"max,omitempty"`
}

type Instance struct {
	Type       string      `json:"type"`
	InstanceId int         `json:"instance_id"`
	Attributes []Attribute `json:"attributes"`
}

type SetAttribute struct {
	InstanceID int         `json:"instance_id"`
	Attribute  string      `json:"attribute"`
	Val        interface{} `json:"val"`
}

type SetRequest struct {
	Attributes []SetAttribute `json:"attributes"`
}

func (p Server) SetAttributes(context context.Context, request *proto.SetAttributesReq) (resp *proto.SetAttributesResp, err error) {
	logrus.Debugf("%s SetAttribute", request.Identity)

	var req SetRequest
	err = json.Unmarshal(request.Data, &req)
	if err != nil {
		return
	}
	for _, attr := range req.Attributes {
		logrus.Debugf("set %s %d attr %s %v:\n", request.Identity, attr.InstanceID, attr.Attribute, attr.Val)
		err = p.Manager.SetAttribute(request.Identity, attr.InstanceID, attr.Attribute, attr.Val)
		if err != nil {
			return
		}
	}
	resp = new(proto.SetAttributesResp)
	resp.Success = true
	return
}
func (p Server) StateChange(request *proto.Empty, server proto.Plugin_StateChangeServer) error {
	log.Println("stateChange requesting...")

	nc := make(chan Notify, 20)

	p.Manager.Subscribe(nc)
	defer p.Manager.Unsubscribe(nc)

	for {
		select {
		case <-server.Context().Done():
			return nil
		case n := <-nc:
			var s proto.State
			s.Identity = n.Identity
			s.InstanceId = int32(n.InstanceID)
			s.Attributes, _ = json.Marshal(n.Attribute)
			log.Printf("notification:%#v\n", s)
			server.Send(&s)
		}
	}
}

func (p *Server) Init() {
	p.pluginRouter.Group("html").Static("", p.staticDir)

	// 压缩静态文件，返回压缩包
	fileName := fmt.Sprintf("%s.zip", p.Domain)

	if !Exist(fileName) {
		if err := archive.Zip(fileName, p.staticDir, p.configFile); err != nil {
			logrus.Errorf("archive file %s err: %s", p.staticDir, err.Error())
			return
		}
	}
	archiveAPI := fmt.Sprintf("resources/archive/%s", fileName)
	p.pluginRouter.StaticFile(archiveAPI, fileName)
}

func Exist(name string) bool {
	_, err := os.Stat(name)
	if err == nil {
		return true
	}
	//if errors.Is(err, os.ErrNotExist) {
	//	return false, nil
	//}
	return false
}

type OptionFunc func(s *Server)

func WithStatic(staticDir string) OptionFunc {
	return func(s *Server) {
		s.staticDir = staticDir
	}
}
func WithConfigFile(configFile string) OptionFunc {
	return func(s *Server) {
		s.configFile = configFile
	}
}

func NewPluginServer(domain string, opts ...OptionFunc) *Server {
	m := NewManager()
	m.Init()

	route := gin.Default()
	path := fmt.Sprintf("api/plugin/%s", domain)
	pluginGroup := route.Group(path)

	s := Server{
		Manager:      m,
		Domain:       domain, // TODO 校验domain格式
		Router:       route,
		pluginRouter: pluginGroup,
		ApiRouter:    pluginGroup.Group("api"),
		staticDir:    "./html",
		configFile:   "./config.json",
	}
	for _, opt := range opts {
		opt(&s)
	}
	s.Init()
	return &s
}
