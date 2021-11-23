package device

import (
	"context"
	"encoding/json"
	"log"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/proto"
	"github.com/zhiting-tech/smartassistant/pkg/plugin/sdk/server"
	"google.golang.org/grpc"
)

var d1 = NewLightBulbDevice("123", "ceiling17")
var d2 = NewSwitchDevice("456")

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func TestNewPlugin(t *testing.T) {

	p := server.NewPluginServer()

	go func() {
		time.Sleep(1 * time.Second)
		if err := p.Manager.AddDevice(d1); err != nil {
			log.Panicln(err)
		}
		if err := p.Manager.AddDevice(d2); err != nil {
			log.Panicln(err)
		}
	}()

	err := sdk.Run(p)
	if err != nil {
		log.Panicln(err)
	}

}

func TestGetAttributes(t *testing.T) {

	grpcAddr := ":53262"
	time.Sleep(time.Second)
	// subscribe state change
	conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
	if err != nil {
		log.Panicln(err)
	}
	cli := proto.NewPluginClient(conn)
	req := proto.GetAttributesReq{
		Identity: d1.identity,
	}
	resp, err := cli.GetAttributes(context.Background(), &req)
	if err != nil {
		log.Panicln(err)
	}
	log.Println(resp)
}
func TestSwitch(t *testing.T) {
	grpcAddr := ":60657"
	time.Sleep(time.Second)
	// subscribe state change
	conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
	if err != nil {
		log.Panicln(err)
	}
	cli := proto.NewPluginClient(conn)

	ExecCmdPower(cli, d1, "on")
}

func TestListen(t *testing.T) {

	grpcAddr := ":60657"
	// subscribe state change
	conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure())
	if err != nil {
		log.Panicln(err)
	}
	cli := proto.NewPluginClient(conn)
	go func() {
		cc, err := cli.StateChange(context.Background(), &proto.Empty{})
		if err != nil {
			log.Panicln(err)
		}
		for {
			s, err := cc.Recv()
			if err != nil {
				log.Println(err)
				continue
			}
			log.Println("state change:", s, err)
		}
	}()
}

func ExecCmdPower(cli proto.PluginClient, d *LightBulbDevice, onOff string) {

	args := server.SetRequest{
		[]server.SetAttribute{{1, "power", onOff}},
	}
	data, _ := json.Marshal(args)
	_, err := cli.SetAttributes(context.Background(), &proto.SetAttributesReq{
		Identity: d.identity, Data: data,
	})
	if err != nil {
		log.Panicln(err)
	}
}
