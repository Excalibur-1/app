package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/Excalibur-1/configuration"
	"github.com/Excalibur-1/rpc"
	"github.com/Excalibur-1/zipkin"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

// Initially the app is not ready and the isStartReady flag is false.
var isStartReady = false

const (
	app   = "base"
	group = "app"
)

func init() {
	fmt.Println("Loading App Engine ver:1.0.0")
}

type serv func(s *rpc.Server)

type Rpc struct {
	server *rpc.Server
}

func App(s serv, namespace, systemId string, conf configuration.Configuration) {
	server := OptApp(s, namespace, systemId, conf)
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	server.Close(func() {})
}

func OptApp(s serv, namespace, systemId string, conf configuration.Configuration) Rpc {
	server, config := rpc.Engine(systemId, conf).Server(namespace, app, group, systemId)
	if len(config.Tag) > 0 {
		zipkin.Init(systemId, conf, config.Tag)
	}
	// Register the health service.
	grpc_health_v1.RegisterHealthServer(server.Server(), &Health{})
	s(server)
	go func() {
		isStartReady = true
		if err := server.Run(config.Addr); err != nil {
			log.Fatal().Err(err).Msg("App Engine Start has error")
		}
	}()
	return Rpc{server: server}
}

func (r Rpc) Close(close func()) {
	close()
	fmt.Println("App Engine Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := r.server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("App Engine Shutdown has error")
	}
	fmt.Println("App Engine exiting")
}

// Health struct
type Health struct {
}

// Check does the health check and changes the status of the server based on weather the app is ready or not.
func (h *Health) Check(context.Context, *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	if isStartReady == true {
		return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING}, nil
	} else if isStartReady == false {
		return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING}, nil
	} else {
		return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_UNKNOWN}, nil
	}
}

// Watch is used by clients to receive updates when the service status changes.
// Watch only dummy implemented just to satisfy the interface.
func (h *Health) Watch(*grpc_health_v1.HealthCheckRequest, grpc_health_v1.Health_WatchServer) error {
	return status.Error(codes.Unimplemented, "Watching is not supported")
}
