package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/internal/entity"
	"github.com/zhiting-tech/smartassistant/internal/task"
	"github.com/zhiting-tech/smartassistant/internal/utils/session"
	plugin2 "github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"
)

var (
	ErrClientNotFound = errors.New("client not found")
)

const attributeChange = "attribute_change"

// Server WebSocket服务端
type Server struct {
	bucket *bucket
}

func NewWebSocketServer() *Server {
	return &Server{
		bucket: newBucket(),
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *Server) AcceptWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logrus.Error(err)
		return
	}
	var (
		lAddr = conn.LocalAddr().String()
		rAddr = conn.RemoteAddr().String()
	)
	logrus.Debugf("start websocket serve \"%s\" with \"%s\"", lAddr, rAddr)
	cli := &client{
		key:    uuid.New().String(),
		conn:   conn,
		bucket: s.bucket,
		send:   make(chan []byte, 4),
	}

	s.bucket.register <- cli
	logrus.Debug("new client Key：", cli.key)
	user := session.Get(c)
	go cli.readWS(user)
	go cli.writeWS()
}

// SingleCast 发送单播消息
func (s *Server) SingleCast(cliID string, data []byte) error {
	cli := s.bucket.get(cliID)
	if cli == nil {
		return ErrClientNotFound
	}
	cli.send <- data
	return nil
}

func (s *Server) Broadcast(data []byte) {
	s.bucket.broadcast <- data
}

func (s *Server) Run(ctx context.Context) {
	logrus.Info("starting websocket server")
	go s.bucket.run()
	<-ctx.Done()
	s.bucket.stop()
	logrus.Warning("websocket server stopped")
}

// OnDeviceStateChange 设备状态改变回调，会广播给所有客户端，并且触发场景
func (s *Server) OnDeviceStateChange(identity string, instanceID int, dsNew plugin2.Attribute) {
	resp := Event{
		EventType: attributeChange,
		Data: map[string]interface{}{
			"identity":    identity,
			"instance_id": instanceID,
			"attr":        dsNew,
		},
	}
	data, _ := json.Marshal(resp)
	s.Broadcast(data)
	logrus.Debug("broadcast state change:", string(data))
	task.GetManager().DeviceStateChange(identity, entity.Attribute{Attribute: dsNew, InstanceID: instanceID})
}
