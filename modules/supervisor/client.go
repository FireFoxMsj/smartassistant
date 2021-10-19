package supervisor

import (
	"context"
	"log"
	"sync"

	"github.com/zhiting-tech/smartassistant/modules/supervisor/proto"
	"google.golang.org/grpc"
)

var (
	client proto.SupervisorClient
	_conce sync.Once
)

func getClient() proto.SupervisorClient {
	_conce.Do(func() {
		conn, err := grpc.Dial("supervisor:9090", grpc.WithInsecure())
		if err != nil {
			log.Fatalln(err)
		}
		client = proto.NewSupervisorClient(conn)
	})
	return client
}

func Restart(name string) error {
	if len(name) == 0 {
		name = saImage.RefStr()
	}
	req := &proto.RestartReq{
		Image:    saImage.RefStr(),
		NewImage: name,
	}
	_, err := getClient().Restart(context.Background(), req)
	return err
}
