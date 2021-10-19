package datatunnel

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	cache2 "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/pkg/datatunnel/proto"
	"github.com/zhiting-tech/smartassistant/pkg/rand"
)

const (
	MonitorSleepTime      = (time.Second * time.Duration(30))
	defaultReadTimeout    = (time.Second * 30)
	cacheKeyLen           = 64
	defaultCleanTime      = (time.Second * 5)
	defaultExpirationTime = defaultReadTimeout + defaultCleanTime
)

// readCount 在timeout时间内获取count长度的字节,超时则出错
func readCount(conn net.Conn, count int, timeout time.Duration) ([]byte, error) {
	if count <= 0 || timeout < 0 {
		return nil, errors.New("count is less than 0")
	}

	result := make([]byte, count)
	conn.SetReadDeadline(time.Now().Add(timeout))
	n, err := io.ReadFull(conn, result)
	if err == nil && n != count {
		text := fmt.Sprintf("Only Read %d bytes", n)
		err = errors.New(text)
	}

	conn.SetReadDeadline(time.Time{})
	return result, err
}

type tunnel struct {
	Server          *TunnelServer
	ControlStopTime time.Time
	Mu              sync.Mutex
	ClientPort      int
}

type PortRange struct {
	Min int
	Max int
}

type DatatunnelControlServer struct {
	proto.UnimplementedDatatunnelControllerServer
	tunnelMap     sync.Map // 通道Map key(通过getMapKey获取) -> *tunnel
	serverMap     sync.Map // 控制通道Map key(通过getMapKey获取) -> grpc_stream_server
	doAuth        func(saID string, key string) bool
	clientPort    PortRange // SA可用端口范围
	clientCurPort int32
	cache         *cache2.Cache
}

type DatatunnelControlServerOption struct {
	DoAuth     func(saID string, key string) bool
	ClientPort PortRange
}

func NewDatatunnelControlServer(option *DatatunnelControlServerOption) *DatatunnelControlServer {
	return &DatatunnelControlServer{
		doAuth:        option.DoAuth,
		clientPort:    option.ClientPort,
		clientCurPort: int32(option.ClientPort.Min) - 1,
		cache:         cache2.New(defaultExpirationTime, defaultCleanTime),
	}
}

func (s *DatatunnelControlServer) Go() {
	// TODO:优化监控SA连接,等待控制通道一段时间无连接后再停止
	// go s.monitor()
}

func (s *DatatunnelControlServer) monitor() {
	for {
		s.tunnelMap.Range(func(key, value interface{}) bool {
			_, load := s.serverMap.Load(key)
			if !load {
				t := value.(*tunnel)
				t.Mu.Lock()
				defer t.Mu.Unlock()
				if time.Since(t.ControlStopTime) > (time.Duration(10) * time.Minute) {
					s.tunnelMap.Delete(key)
				}
			}
			return true
		})

		time.Sleep(MonitorSleepTime)
	}
}

// GetSAConnection 获取SA连接
func (s *DatatunnelControlServer) GetSAConnection(saID string, serviceName string, timeout int) net.Conn {
	key := s.getMapKey(saID, serviceName)
	val, load := s.tunnelMap.Load(key)
	if !load {
		return nil
	}

	t := val.(*tunnel)
	return t.Server.GetClient(timeout)
}

// SAIsConnected SA的控制通道与数据通道是否已经连接成功
func (s *DatatunnelControlServer) SAIsConnected(saID string, serviceName string) bool {
	key := s.getMapKey(saID, serviceName)
	_, load1 := s.serverMap.Load(key)
	_, load2 := s.tunnelMap.Load(key)
	return load1 && load2
}

// SendNewConnection 通知SA创建新连接
func (s *DatatunnelControlServer) SendNewConnection(saID string, serviceName string) error {
	key := s.getMapKey(saID, serviceName)
	val, load := s.serverMap.Load(key)
	if load {
		// 发送ascii字符,对应key的固定长度
		connectionKey := rand.StringK(cacheKeyLen, rand.KindAll)
		bytes, err := json.Marshal(&NewActionData{
			ServiceName:   serviceName,
			ConnectionKey: connectionKey, // SA连接后需要发送的认证信息
		})
		if err != nil {
			return err
		}
		s.cache.SetDefault(connectionKey, struct{}{})
		streamServer := val.(proto.DatatunnelController_ControlStreamServer)
		return streamServer.Send(&proto.ControlStreamData{
			Action:      NewAction,
			ActionValue: string(bytes),
		})
	}

	text := fmt.Sprintf("Cannot find Stream server, id : %s", saID)
	return errors.New(text)
}

// SendCreateAction 认证通过,通知SA创建数据通道客户端
func (s *DatatunnelControlServer) SendCreateAction(server proto.DatatunnelController_ControlStreamServer,
	port int, serviceName string) (err error) {

	var (
		bytes []byte
	)
	bytes, err = json.Marshal(&CreateActionData{
		Port:        port,
		ServiceName: serviceName,
	})
	if err != nil {
		return
	}
	err = server.Send(&proto.ControlStreamData{
		Action:      CreateAction,
		ActionValue: string(bytes),
	})
	return
}

func (s *DatatunnelControlServer) ControlStream(server proto.DatatunnelController_ControlStreamServer) error {
	for {
		data, err := server.Recv()
		if err != nil {
			logrus.Warn("ControlServer Stream error %v\n", err)
			break
		}

		switch data.Action {
		case AuthAction:
			actionData := AuthActionData{}
			if err = GetActionData(data.ActionValue, &actionData); err == nil {
				_, err := s.authAndCreate(server, actionData)
				if err == nil {
					key := s.getMapKey(actionData.SAID, actionData.ServiceName)
					s.serverMap.LoadOrStore(key, server)
					// 控制通道异常停止当前SA的端口监听
					defer s.deleteTunnel(key)
				} else {
					logrus.Warnf("ControlServer AuthAction error %v", err)
					return err
				}
			}
		}
	}

	return nil
}

func (s *DatatunnelControlServer) getMapKey(saID string, serviceName string) string {
	return fmt.Sprintf("%s_%s", saID, serviceName)
}

func (s *DatatunnelControlServer) authAndCreate(server proto.DatatunnelController_ControlStreamServer, data AuthActionData) (*tunnel, error) {
	// 控制通道认证处理
	if s.doAuth == nil || !s.doAuth(data.SAID, data.Key) {
		server.Send(&proto.ControlStreamData{
			Action:      ErrorAction,
			ActionValue: "auth error",
		})
		text := fmt.Sprintf("control client auth error, SA-ID:%s Key:%s", data.SAID, data.Key)
		return nil, errors.New(text)
	}

	// 获取通道
	mapKey := s.getMapKey(data.SAID, data.ServiceName)
	t, err := s.createTunnel(mapKey)
	if err != nil {
		server.Send(&proto.ControlStreamData{
			Action:      ErrorAction,
			ActionValue: err.Error(),
		})
		return nil, err
	} else {
		err := s.SendCreateAction(server, t.ClientPort, data.ServiceName)
		if err != nil {
			s.deleteTunnel(mapKey)
			return nil, err
		}
	}

	return t, nil
}

func (s *DatatunnelControlServer) createTunnel(key string) (*tunnel, error) {
	// 存在则返回旧通道,防止重复创建
	// 同一个SAID使用同一个通道
	val, load := s.tunnelMap.Load(key)
	if load {
		text := fmt.Sprintf("has same key tunnel, Key:%s", key)
		return nil, errors.New(text)
	}

	lisenter := s.getRandomClientListener()
	if lisenter == nil {
		return nil, errors.New("Cannot get random client listener")
	}

	server := NewTunnelServer(&TunnelServerOption{
		// 当客户端连接时,获取规定的Key的长度,检查是否过期,判断SA连接是否合法
		OnClientConnected: func(conn net.Conn) error {
			buf, err := readCount(conn, cacheKeyLen, defaultReadTimeout)
			if err != nil {
				return err
			}

			key := string(buf)
			if _, ok := s.cache.Get(key); !ok {
				return errors.New("Cannot find cache")
			}

			s.cache.Delete(key)
			return nil
		},
	})
	t := &tunnel{
		Server:     server,
		ClientPort: lisenter.Addr().(*net.TCPAddr).Port,
	}

	// 创建数据通道,以Store进缓存的为准
	val, load = s.tunnelMap.LoadOrStore(key, t)
	if load {
		lisenter.Close()
		return val.(*tunnel), nil
	} else {
		t.Server.Go()
		t.Server.GoClient(lisenter)
		return t, nil
	}

}

// deleteTunnel 清理对应的SA
func (s *DatatunnelControlServer) deleteTunnel(key string) {
	s.serverMap.Delete(key)
	val, load := s.tunnelMap.LoadAndDelete(key)
	if load {
		t := val.(*tunnel)
		t.Server.Stop()
	}
}

func (s *DatatunnelControlServer) getRandomClientListener() *net.TCPListener {
	cur := atomic.LoadInt32(&s.clientCurPort)
	port, listener := s.getRandomTcpListener(cur+1, int32(s.clientPort.Min), int32(s.clientPort.Max))
	if listener != nil {
		atomic.CompareAndSwapInt32(&s.clientCurPort, cur, port)
	}

	return listener
}

func (s *DatatunnelControlServer) getRandomTcpListener(cur int32, min int32, max int32) (port int32, listener *net.TCPListener) {
	var (
		err error
	)
	for port = cur; port < max; port++ {
		addr, _ := net.ResolveTCPAddr("tcp", fmt.Sprintf("0.0.0.0:%v", port))
		listener, err = net.ListenTCP("tcp", addr)
		if err == nil {
			return
		}
	}

	for port = min; port < cur; port++ {
		addr, _ := net.ResolveTCPAddr("tcp", fmt.Sprintf("0.0.0.0:%v", port))
		listener, err = net.ListenTCP("tcp", addr)
		if err == nil {
			return
		}
	}

	return
}
