package main

import (
	"context"
	"encoding/json"
	comet "gitlab.yctc.tech/root/smartassistent.git/core/grpc/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"io"
	"log"
	"time"
)

// ConnetInfo 建立
type ConnetInfo struct {
	PluginName string `json:"plugin_name"`
}

func main() {
	// 创建连接
	conn, err := grpc.Dial("localhost:9234", grpc.WithInsecure())
	if err != nil {
		log.Printf("连接失败: [%v]\n", err)
		return
	}
	defer conn.Close()

	// 声明客户端
	client := comet.NewCometClient(conn)

	// 声明 context
	// md := metadata.Pairs("clientid", "cli123")

	md := metadata.New(map[string]string{"domain": "cli222"})

	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// 创建双向数据流
	stream, err := client.BidStream(ctx)
	if err != nil {
		log.Printf("创建数据流失败: [%v]\n", err)
	}
	ci := ConnetInfo{PluginName: "testPlugin"}
	body, _ := json.Marshal(ci)

	time.AfterFunc(3*time.Second, func() {
		stream.Send(&comet.Request{Body: body})
	})

	for {
		// 接收从 服务端返回的数据流
		recv, err := stream.Recv()
		if err == io.EOF {
			log.Println("io.EOF")
			break // 如果收到结束信号，则退出“接收循环”，结束客户端程序
		}

		if err != nil {
			// TODO: 处理接收错误
			log.Println("接收数据出错:", err)
			return
		}

		// 没有错误的情况下，打印来自服务端的消息
		log.Printf("[客户端收到]: %s", recv.Body)
	}
}
