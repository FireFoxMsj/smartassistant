package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/proto"
)

type Server struct {
	Manager    *Manager
	Domain     string
	Router     *gin.Engine
	ApiRouter  *gin.RouterGroup
	HtmlRouter *gin.RouterGroup
}

func (p Server) Discover(context context.Context, request *proto.Empty, server proto.Plugin_DiscoverStream) error {
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

func (p Server) GetAttributes(context context.Context, request *proto.GetAttributesReq, resp *proto.GetAttributesResp) error {
	log.Println("GetAttributes:", request)

	instances, err := p.Manager.GetAttributes(request.Identity)
	if err != nil {
		return err
	}

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
	return nil
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

func (p Server) SetAttributes(context context.Context, request *proto.SetAttributesReq, resp *proto.SetAttributesResp) error {
	log.Println("SetAttribute:", request)
	var req SetRequest
	err := json.Unmarshal(request.Data, &req)
	if err != nil {
		return err
	}
	for _, attr := range req.Attributes {
		logrus.Debugf("set %s %d attr %s %v:\n", request.Identity, attr.InstanceID, attr.Attribute, attr.Val)
		err := p.Manager.SetAttribute(request.Identity, attr.InstanceID, attr.Attribute, attr.Val)
		if err != nil {
			return err
		}
	}
	resp.Success = true
	return nil
}
func (p Server) StateChange(context context.Context, request *proto.Empty, server proto.Plugin_StateChangeStream) error {
	log.Println("stateChange requesting...")

	nc := make(chan Notify, 20)

	p.Manager.Subscribe(nc)
	defer p.Manager.Subscribe(nc)

	for {
		select {
		case <-context.Done():
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

func NewPluginServer(domain string) *Server {
	m := NewManager()
	m.init()

	route := gin.Default()
	path := fmt.Sprintf("/plugin/%s", domain)
	pluginGroup := route.Group(path)

	return &Server{
		Manager:    m,
		Domain:     domain,
		Router:     route,
		ApiRouter:  pluginGroup.Group("api"),
		HtmlRouter: pluginGroup.Group("html"),
	}
}
