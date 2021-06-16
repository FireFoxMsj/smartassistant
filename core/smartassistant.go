package core

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/gin-gonic/gin"

	comet "gitlab.yctc.tech/root/smartassistent.git/core/grpc/proto"
)

var (
	Sass = NewSmartAssistant()
)

// SmartAssistant 管理整个家居的集成大脑
type SmartAssistant struct {
	Bus       *EventBus
	Services  *ServiceRegistry
	data      map[string]interface{}
	GinEngine *gin.Engine
}

func NewSmartAssistant() *SmartAssistant {
	sa := &SmartAssistant{}
	sa.Services = NewServiceRegistry(sa)
	sa.Bus = NewEventBus()
	sa.GinEngine = gin.Default()
	return sa
}

type Service struct {
	serviceFunc serviceFunc
}

// ServiceRegistry 注册服务
type ServiceRegistry struct {
	_services map[string]map[string]Service
	sass      *SmartAssistant
}

// NewServiceRegistry
func NewServiceRegistry(sass *SmartAssistant) *ServiceRegistry {
	_services := make(map[string]map[string]Service)
	return &ServiceRegistry{sass: sass, _services: _services}
}

func (sr *ServiceRegistry) toLower(args ...string) (lowers []string) {
	for _, arg := range args {
		lowers = append(lowers, strings.ToLower(arg))
	}
	return
}

// Register
func (sr *ServiceRegistry) Register(domain, serviceName string, serviceFunc serviceFunc) {
	res := sr.toLower(domain, serviceName)
	domain, serviceName = res[0], res[1]
	newSer := Service{
		serviceFunc: serviceFunc,
	}
	if ser, ok := sr._services[domain]; ok {
		ser[serviceName] = newSer
	} else {
		service := make(map[string]Service)
		service[serviceName] = newSer
		sr._services[domain] = service
	}

	log.Printf("Register %s %s successfully", domain, serviceName)
}

// Call 调用服务，根据domain找到对应的stream，推送给对应的插件。
func (sr *ServiceRegistry) Call(domain, serviceName string, eventData M) (err error) {
	if serviceName == "discover" {
		discover(eventData)
		return
	}
	res := sr.toLower(domain, serviceName)
	domain, serviceName = res[0], res[1]

	log.Printf("Calling %s %s \n", domain, serviceName)

	// 本地已注册当前服务，不再执行插件定制
	if _, ok := sr._services[domain]; ok {
		if handler, ok := sr._services[domain][serviceName]; ok {
			return handler.serviceFunc(eventData)
		}
	}

	if StreamBucket.Client(domain) == nil {
		err = fmt.Errorf("unregister domain(%s)", domain)
		return
	}
	if body, err := json.Marshal(eventData); err != nil {
		return err
	} else {
		req := &comet.Request{Event: EventCallService, Body: body}
		PushPlugin(domain, req)
	}
	return
}
