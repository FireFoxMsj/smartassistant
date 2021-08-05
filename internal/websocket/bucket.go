package websocket

import (
	"github.com/sirupsen/logrus"
)

// bucket 管理websocket连接的桶
type bucket struct {
	clients    map[string]*client
	broadcast  chan []byte
	register   chan *client
	unregister chan *client
}

func newBucket() *bucket {
	return &bucket{
		clients:    make(map[string]*client),
		register:   make(chan *client, 2),
		unregister: make(chan *client, 2),
		broadcast:  make(chan []byte),
	}
}

func (b *bucket) run() {
	for {
		select {
		case client := <-b.register:
			logrus.Debug("register new client", client.key)
			b.put(client)
		case client := <-b.unregister:
			logrus.Debug("del client", client.key)
			b.remove(client)
		case message := <-b.broadcast:
			for _, client := range b.clients {
				logrus.Debug("broadcast clientKey", client.key)
				select {
				case client.send <- message:
				}
			}
		}
	}
}

// 断开客户端，停止
func (b *bucket) stop() {
	for _, client := range b.clients {
		b.remove(client)
	}
}

func (b *bucket) put(cli *client) {
	b.clients[cli.key] = cli
}
func (b *bucket) get(key string) *client {
	return b.clients[key]
}

func (b *bucket) remove(cli *client) {
	if _, ok := b.clients[cli.key]; ok {
		close(cli.send)
		delete(b.clients, cli.key)
		if cli.conn != nil {
			cli.conn.Close()
		}
	}
}
