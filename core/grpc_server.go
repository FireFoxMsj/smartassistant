package core

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	comet "gitlab.yctc.tech/root/smartassistent.git/core/grpc/proto"
	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
)

// TODO 需要进一步处理错误

type streamChan chan *comet.Request

var StreamBucket *Bucket

func init() {
	StreamBucket = NewBucket()
}

// PushPlugin 推送消息到插件
func PushPlugin(domain string, resp *comet.Request) {
	if cli := StreamBucket.Client(domain); cli != nil {
		select {
		case cli.sendChan <- resp:
		default:
			log.Println("client sendchan buffer is full")
		}
	}
}

type Bucket struct {
	clis sync.Map
}

func NewBucket() *Bucket {
	return &Bucket{clis: sync.Map{}}
}

func (b *Bucket) Put(k string, v interface{}) {
	k = strings.ToLower(k)
	b.clis.Store(k, v)
}

func (b *Bucket) Remove(k string) {
	log.Println("Bucket remove k", k)
	b.clis.Delete(k)
}

func (b *Bucket) Client(k string) *StreamClient {
	if cli, ok := b.clis.Load(k); !ok {
		return nil
	} else {
		return cli.(*StreamClient)
	}
}

// StreamClient 客户端连接的流
type StreamClient struct {
	ID       string
	Stream   comet.Comet_BidStreamServer
	Domain   string
	sendChan streamChan
}

func (cli *StreamClient) Listen() error {
	log.Printf("Listen domain：%s \n", cli.Domain)
	StreamBucket.Put(cli.Domain, cli)
	go cli.send()
	// go cli.heartbeat()
	return cli.recv()
}

// heartbeat 心跳 TODO
func (cli *StreamClient) heartbeat() {

}

// send 发送消息
func (cli *StreamClient) send() {
	defer StreamBucket.Remove(cli.Domain)

	for {
		data, ok := <-cli.sendChan
		if !ok {
			log.Println("sendchan ok break")
			break
		}
		cli.Stream.SendMsg(data)
	}
}

// recv 接收消息
// 支持以两个事件：
// device_discovered
// state_changed
// heartbeat
func (cli *StreamClient) recv() error {
	defer StreamBucket.Remove(cli.Domain)

	for {
		req, err := cli.Stream.Recv()
		if err != nil {
			return err
		}

		m1 := M{}
		// 参数错误，不断开连接
		if err := json.Unmarshal(req.GetBody(), &m1); err != nil {
			log.Println(err)
			continue
		}
		// TODO 此处要处理heartbeat
		Sass.Bus.Fire(req.Event, m1)
	}
}

type Streamer struct{}

func (s *Streamer) BidStream(stream comet.Comet_BidStreamServer) error {
	ctx := stream.Context()
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.DataLoss, "failed to get metadata")
	}

	if t, ok := md["domain"]; ok && len(t[0]) != 0 {
		k := t[0]
		if cli := StreamBucket.Client(k); cli != nil {
			return status.Errorf(codes.AlreadyExists, "domain (%s) already exists", k)
		}
		cli := StreamClient{
			ID:       uuid.New().String(),
			Stream:   stream,
			Domain:   k,
			sendChan: make(streamChan, 4),
		}
		if err := cli.Listen(); err != nil {
			return err
		}
	} else {
		return status.Errorf(codes.InvalidArgument, "metadata need the key: client_id")
	}
	return nil
}

func (s *Streamer) FindDevice(ctx context.Context, req *comet.DeviceReq) (cd *comet.Device, err error) {
	var (
		d  orm.Device
		id int
	)

	id, err = strconv.Atoi(req.Id)
	if err != nil {
		return
	}
	if d, err = orm.GetDeviceByID(id); err != nil {
		return
	}
	cd = &comet.Device{
		Name: d.Name,
	}
	return cd, err
}

func StartRPC() {
	var (
		err error
		lis net.Listener
	)
	server := grpc.NewServer()
	comet.RegisterCometServer(server, &Streamer{})
	if lis, err = net.Listen("tcp", ":9234"); err != nil {
		panic(err)
	}

	if err = server.Serve(lis); err != nil {
		panic(err)
	}

}

func discover(data M) {
	StreamBucket.clis.Range(func(key, value interface{}) bool {
		body, _ := json.Marshal(data)
		req := &comet.Request{Event: EventCallService, Body: body}
		PushPlugin(key.(string), req)
		return true
	})
}
