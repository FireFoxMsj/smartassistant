package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	http2 "gitlab.yctc.tech/root/smartassistent.git/core/http"
	"gitlab.yctc.tech/root/smartassistent.git/core/http/middleware"
	"gitlab.yctc.tech/root/smartassistent.git/core/orm"
	"gitlab.yctc.tech/root/smartassistent.git/core/plugin"
	"gitlab.yctc.tech/root/smartassistent.git/core/smq"
	"gitlab.yctc.tech/root/smartassistent.git/utils/errors"
	"gitlab.yctc.tech/root/smartassistent.git/utils/session"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"gitlab.yctc.tech/root/smartassistent.git/core"
)

type Client struct {
	Key    string
	Conn   *websocket.Conn
	send   chan []byte
	bucket *Bucket
}

type Bucket struct {
	clients map[string]*Client

	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

var bucket = NewBucket()

func NewBucket() *Bucket {
	return &Bucket{
		clients:    make(map[string]*Client),
		register:   make(chan *Client, 2),
		unregister: make(chan *Client, 2),
		broadcast:  make(chan []byte),
	}
}

func (b *Bucket) run() {
	for {
		select {
		case client := <-b.register:
			log.Println("register new client", client.Key)
			b.put(client)
		case client := <-b.unregister:
			log.Println("del client", client.Key)
			b.remove(client)
		case message := <-b.broadcast:
			for _, client := range b.clients {
				log.Println("broadcast clientKey", client.Key)
				select {
				case client.send <- message:
				}
			}
		}
	}
}

func (b *Bucket) put(cli *Client) {
	b.clients[cli.Key] = cli
}
func (b *Bucket) get(key string) *Client {
	return b.clients[key]
}

func (b *Bucket) remove(cli *Client) {
	if _, ok := b.clients[cli.Key]; ok {
		close(cli.send)
		delete(b.clients, cli.Key)
		if cli.Conn != nil {
			cli.Conn.Close()
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var callServicePool sync.Pool

func init() {
	callServicePool.New = func() interface{} {
		return &core.CallService{}
	}
}

// readWS
func (cli *Client) readWS(user session.User) {
	defer func() { cli.bucket.unregister <- cli }()

	for {
		_, data, err := cli.Conn.ReadMessage()
		if err != nil {
			return
		}
		go func() {
			if err = cli.handleWSMessage(data, user); err != nil {
				log.Printf("handleWSMessageErr: %+v\n", err)
				return
			}
		}()
	}
}

func (cli *Client) handleWSMessage(data []byte, user session.User) (err error) {
	log.Println("readMessage：", string(data))

	cs := callServicePool.Get().(*core.CallService)
	defer callServicePool.Put(cs)
	cs.Reset(cli.Key)

	if err = json.Unmarshal(data, cs); err != nil {
		log.Printf("readWS json.UnMarshal err %v \n", err)
		return err
	}
	cs.ServiceData["call_service_id"] = cs.ID
	cs.ServiceData["service_name"] = cs.Service

	deviceID := cs.ServiceData.ValInt("device_id")
	if deviceID != 0 { // 控制设备的命令

		device, err := orm.GetDeviceByID(deviceID)
		if err != nil {
			return err
		}
		cs.ServiceData.Put("id", device.Identity)
		cs.ServiceData.Put("user_id", user.UserID)

		// 根据插件配置判断用户是否具有权限
		if !orm.IsCmdPermit(user.UserID, device, cs.Service) {
			return errors.New(errors.Deny)
		}
	}
	go func() {
		if err := core.Sass.Services.Call(cs.Domain, cs.Service, cs.ServiceData); err != nil {
			log.Printf("readWS Call service(%s.%s) err %v \n", cs.Domain, cs.Service, err)
		}
	}()

	return
}

// TODO writeWS
func (cli *Client) writeWS() {
	defer func() { cli.bucket.unregister <- cli }()

	for {
		select {
		case msg, ok := <-cli.send:
			if !ok {
				cli.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			cli.Conn.WriteMessage(websocket.TextMessage, msg)
		}
	}
}

func BroadcastToClient(event core.Event) error {
	// 将硬件id转换成设备id
	device, err := orm.GetDeviceByIdentity(event.Data.ValString("id"))
	if err != nil {
		return err
	}
	event.Data.Put("device_id", device.ID)

	if b, err := json.Marshal(event); err != nil {
		return err
	} else {
		bucket.broadcast <- b
	}
	return nil
}

// ExecuteSceneTask 设备状态变更时触发场景
func ExecuteSceneTask(event core.Event) error {
	device, err := orm.GetDeviceByIdentity(event.Data.ValString("id"))
	if err != nil {
		return err
	}

	state := event.Data.Get("state").(map[string]interface{})
	smq.DeviceStateChange(device.ID, state)
	return nil
}

func SingleCast(event core.Event) error {
	key := event.Data.ValString("client_key")
	if key == "" {
		return nil
	}
	cli := bucket.get(key)
	msg, _ := json.Marshal(event.Data.Get("data"))
	cli.send <- msg
	return nil
}

// TODO 单元测试
func main() {
	r := core.Sass.GinEngine

	core.Sass.Bus.Listen(core.EventSingleCast, SingleCast)
	core.Sass.Bus.Listen(core.EventStateChanged, BroadcastToClient, ExecuteSceneTask)

	plugin.Load()
	go core.StartRPC()

	go bucket.run()

	r.GET("/ws", middleware.RequireToken, func(c *gin.Context) {
		acceptWs(c, bucket)
	})

	http2.LoadModules(r)

	r.Static("html", "./plugins")

	if err := r.Run(":8088"); err != nil {
		panic(err)
	}
}

// acceptWs
// TODO
//  1）完善websocket 的心跳
// 	2）完善读写协议
func acceptWs(c *gin.Context, bucket *Bucket) {
	log.Println("acceptWs")
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(500, err)
		return
	}

	var (
		lAddr = conn.LocalAddr().String()
		rAddr = conn.RemoteAddr().String()
	)

	i := fmt.Sprintf("start websocket serve \"%s\" with \"%s\"", lAddr, rAddr)
	log.Println(i)

	cli := &Client{
		Key:    uuid.New().String(),
		Conn:   conn,
		bucket: bucket,
		send:   make(chan []byte, 4),
	}

	bucket.register <- cli
	log.Println("new client Key：", cli.Key)

	user := session.Get(c)
	go cli.readWS(*user)
	go cli.writeWS()
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
