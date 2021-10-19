package datatunnel

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type TunnelServer struct {
	clientListener *net.TCPListener
	// userListener   *net.TCPListener
	clientAdd         chan *clientConnection
	clientGet         chan net.Conn
	clientMap         sync.Map
	logger            *logrus.Logger
	ctx               context.Context
	cancelCtx         context.CancelFunc
	onClientConnected func(conn net.Conn) error
	lingerSecond      int // tcp参数SO_LINGER
}

type TunnelServerOption struct {
	OnClientConnected func(conn net.Conn) error
	LingerSecond      int // 单位秒
	LogLevel          string
}

type clientConnection struct {
	Conn       net.Conn
	AcceptTime time.Time
}

func NewTunnelServer(option *TunnelServerOption) *TunnelServer {
	logger := logrus.New()
	logger.SetLevel(getLogrusLevel(option.LogLevel))
	ctx, cancel := context.WithCancel(context.Background())
	if option.LingerSecond == 0 {
		option.LingerSecond = DefaultLingerSecond
	}
	return &TunnelServer{
		clientAdd:         make(chan *clientConnection, 16),
		clientGet:         make(chan net.Conn, 0),
		logger:            logger,
		ctx:               ctx,
		cancelCtx:         cancel,
		onClientConnected: option.OnClientConnected,
		lingerSecond:      option.LingerSecond,
	}
}

func (s *TunnelServer) Go() {
	go s.manager()
	go s.monitor()
	s.logger.Info("Tunnel Server Running!\n")
}

func (s *TunnelServer) monitor() {
MainLoop:
	for {
		s.logger.Tracef("Start Clean connection\n")
		s.clientMap.Range(func(key, value interface{}) bool {
			client := value.(*clientConnection)
			now := time.Now()
			if now.Sub(client.AcceptTime) > (time.Minute * time.Duration(5)) {
				s.clientMap.Delete(key)
				if client.Conn != nil {
					client.Conn.Close()
				}
				s.logger.Debugf("Clean connection %v, accept time %v, clean time %v\n", client.Conn.RemoteAddr(), client.AcceptTime, now)
			}
			return true
		})
		s.logger.Tracef("Finish Clean connection\n")

		select {
		case <-time.After(time.Second * time.Duration(30)):
		case <-s.ctx.Done():
			break MainLoop
		}

	}

	s.logger.Info("Server monitor stop")
}

func (s *TunnelServer) manager() {
	var getConn net.Conn

MainLoop:
	for {
		if getConn == nil {
			getConn = s.get()
			s.logger.Debugf("getConn %v\n", getConn)
		}

		if getConn == nil {
			select {
			case client := <-s.clientAdd:
				s.logger.Debugf("Add client %v\n", client.Conn.RemoteAddr())
				s.clientMap.Store(client.Conn.RemoteAddr().String(), client)
			case <-time.After(time.Second * time.Duration(30)):
			case <-s.ctx.Done():
				break MainLoop
			}
		} else {
			select {
			case client := <-s.clientAdd:
				s.logger.Debugf("Add client %v\n", client.Conn.RemoteAddr())
				s.clientMap.Store(client.Conn.RemoteAddr().String(), client)
			case s.clientGet <- getConn:
				getConn = nil
			case <-time.After(time.Second * time.Duration(30)):
			case <-s.ctx.Done():
				break MainLoop
			}
		}
	}

	s.logger.Info("Server manager stop")
}

func (s *TunnelServer) get() net.Conn {

	var val interface{}
	s.clientMap.Range(func(key, value interface{}) bool {
		var ok bool
		val, ok = s.clientMap.LoadAndDelete(key)
		if ok {
			return false
		} else {
			return true
		}
	})

	if val == nil {
		return nil
	} else {
		client := val.(*clientConnection)
		return client.Conn
	}
}

func (s *TunnelServer) GetClient(timeout int) net.Conn {
	select {
	case conn := <-s.clientGet:
		return conn
	case <-time.After(time.Millisecond * time.Duration(timeout)):
		return nil
	}
}

func (s *TunnelServer) runClientListener(listener *net.TCPListener) {

	s.logger.Info("Start Client Listener")

	defer listener.Close()
	for {
		client, err := listener.AcceptTCP()
		if err != nil {
			s.logger.Warnf("accept client error %v\n", err)
			break
		}
		s.logger.Infof("client connected %v\n", client.RemoteAddr())
		if s.lingerSecond > 0 {
			client.SetLinger(s.lingerSecond)
		}
		client.SetKeepAlive(true)
		client.SetKeepAlivePeriod(DefaultTCPKeepAlivePeriod)

		if s.onClientConnected == nil {
			s.clientAdd <- &clientConnection{
				Conn:       client,
				AcceptTime: time.Now(),
			}
		} else {
			go func() {
				err := s.onClientConnected(client)
				if err != nil {
					s.logger.Debugf("Server onClientConnected error %v", err)
					client.Close()
					return
				}

				s.clientAdd <- &clientConnection{
					Conn:       client,
					AcceptTime: time.Now(),
				}
			}()
		}
	}
}

func (s *TunnelServer) RunClient(addr *net.TCPAddr) {
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		s.logger.Warnf("listen client error %v\n", err)
		return
	}
	s.clientListener = listener
	s.runClientListener(listener)
}

func (s *TunnelServer) GoClient(listener *net.TCPListener) {
	s.clientListener = listener
	go s.runClientListener(listener)
}

func (s *TunnelServer) StopClient() {
	if s.clientListener != nil {
		s.clientListener.Close()
		s.clientListener = nil
	}
}

func (s *TunnelServer) Stop() {
	s.StopClient()
	s.cancelCtx()
	s.clean()
}

func (s *TunnelServer) clean() {
	s.clientMap.Range(func(key, value interface{}) bool {
		val, ok := s.clientMap.LoadAndDelete(key)
		if ok {
			conn := val.(*clientConnection).Conn
			if conn != nil {
				conn.Close()
			}
		}

		return true
	})
}
