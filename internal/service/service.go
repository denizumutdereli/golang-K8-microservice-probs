package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/denizumutdereli/golang-K8-microservice-probs/internal/config"
	"github.com/denizumutdereli/golang-K8-microservice-probs/internal/transport"
)

type ServiceStatuses struct {
	redisConnected     bool
	websocketConnected bool
}

type StreamService interface {
	Live(ctx context.Context) (int, bool, error)
	Read(ctx context.Context) (int, bool, error)
}

type streamService struct {
	config    *config.Config
	assets    []string
	kafka     *transport.KafkaClient
	redis     *transport.RedisClient
	webSocket *transport.WSClient
	nats      *transport.NatsClient
	status    *ServiceStatuses
}

func NewStreamService(kf *transport.KafkaClient, rd *transport.RedisClient, ws *transport.WSClient, nc *transport.NatsClient, cnf *config.Config) StreamService {

	ctx := context.Background()

	service := &streamService{kafka: kf, redis: rd, webSocket: ws, nats: nc, config: cnf, status: &ServiceStatuses{}}

	service.ConnectToWebSocket(ctx)
	go service.webSocket.MonitorConnection()

	service.MonitorServices(ctx)

	return service
}

func (s *streamService) ConnectToWebSocket(ctx context.Context) error {
	err := s.webSocket.Connect()
	if err != nil {
		log.Fatalf("Fatal error connecting to websocket server: %v", err)
	}
	fmt.Println("Connected to the WebSocket server")

	return nil
}

func (s *streamService) MonitorServices(ctx context.Context) {
	log.Println("## Service monitoring started")

	go func() {
		for {
			connected := s.webSocket.IsConnected()
			if s.status.websocketConnected != connected {
				s.status.websocketConnected = connected
				fmt.Println("WebSocket connection status:", connected)
			}
			time.Sleep(time.Second)
		}
	}()

	go func() {
		for {
			connected := s.redis.IsConnected()
			if s.status.redisConnected != connected {
				s.status.redisConnected = connected
				fmt.Println("Redis connection status:", connected)
			}
			time.Sleep(time.Second)
		}
	}()
}

func (s *streamService) Live(ctx context.Context) (int, bool, error) {
	fmt.Println("Redis:", s.status.redisConnected, "WebSocket:", s.status.websocketConnected)
	if s.status.redisConnected && s.status.websocketConnected {
		return http.StatusOK, true, nil
	}
	return http.StatusServiceUnavailable, false, fmt.Errorf("services are not fully operational")
}

func (s *streamService) Read(ctx context.Context) (int, bool, error) {
	return s.Live(ctx)
}
