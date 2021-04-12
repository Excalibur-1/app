package app_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Excalibur-1/app"
	"github.com/Excalibur-1/configuration"
	"github.com/Excalibur-1/rpc"
	"github.com/Excalibur-1/rpc/testproto"
)

var conf configuration.Configuration

const (
	namespace      = "myconf"
	systemId       = "2000"
	clientSystemId = "8000"
)

func init() {
	conf = configuration.DefaultEngine()
}

type GreeterTest struct {
}

func NewGreeterTest() *GreeterTest {
	return &GreeterTest{}
}

func (t *GreeterTest) SayHello(_ context.Context, req *testproto.HelloRequest) (rep *testproto.HelloReply, err error) {
	rep = new(testproto.HelloReply)
	fmt.Println(req)
	rep.Success = true
	rep.Message = "hello world"
	return
}

func (t *GreeterTest) StreamHello(gss testproto.Greeter_StreamHelloServer) (err error) {
	req, err := gss.Recv()
	if err != nil {
		return
	}
	fmt.Println(req)
	err = gss.Send(&testproto.HelloReply{
		Message: "hello world stream",
		Success: true,
	})
	return
}

// rpc服务端
func TestApp(t *testing.T) {
	go app.App(func(s *rpc.Server) {
		serv := s.Server()
		greeterTest := NewGreeterTest()
		testproto.RegisterGreeterServer(serv, greeterTest)
	}, namespace, systemId, conf)
	time.Sleep(time.Second * 10)
}

// rpc客户端
func TestClient(t *testing.T) {
	conn, _ := rpc.Engine(clientSystemId, conf).ClientConn(systemId)
	defer func() { _ = conn.Close() }()
	greeterClient := testproto.NewGreeterClient(conn)

	// 普通rpc请求（短连接）
	rep, err := greeterClient.SayHello(context.Background(), &testproto.HelloRequest{
		Name: "Tom",
		Age:  18,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(rep)

	// 流式rpc请求（长连接）
	streamHelloClient, err := greeterClient.StreamHello(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	err = streamHelloClient.Send(&testproto.HelloRequest{
		Name: "Alex",
		Age:  19,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	recv, err := streamHelloClient.Recv()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(recv)
}
