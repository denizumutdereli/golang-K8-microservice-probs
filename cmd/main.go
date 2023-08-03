package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/denizumutdereli/golang-K8-microservice-probs/internal/config"
	"github.com/denizumutdereli/golang-K8-microservice-probs/internal/handler"
	"github.com/denizumutdereli/golang-K8-microservice-probs/internal/service"
	"github.com/denizumutdereli/golang-K8-microservice-probs/internal/transport"
	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Starting the service...")

	config, err := config.GetConfig()
	if err != nil {
		log.Fatalf("Fatal error config: %v", err)
	}

	fmt.Println(config.AppName)

	fmt.Println("---------building packages---------")

	redis, err := transport.NewRedisClient(config.RedisURL, config)

	if err != nil {
		log.Fatalf("Fatal error creating redis config: %v", err)
	}

	kafka, err := transport.NewKafkaClient(config.KafkaBrokers, config.KafkaConsumerGroup, config.MaxRetry, time.Duration(config.MaxRetry))
	if err != nil {
		log.Fatalf("Fatal error creating kafka config: %v", err)
	}

	websocket := transport.NewWSClient(config.WsServerURL, time.Duration(config.WsPingPeriod), config.WsPinMaxError, config.MaxRetry, time.Duration(config.MaxRetry))

	nats, err := transport.NewNatsClient(config.NatsURL)

	if err != nil {
		log.Fatalf("Fatal error creating nats config: %v", err)
	}

	streamService := service.NewStreamService(kafka, redis, websocket, nats, config)
	serviceHandler := handler.NewRestHandler(streamService, config)

	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Welcome to the " + config.AppName})
	})

	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"message": "Method or route not found in: " + config.AppName})
	})

	router.GET("/live", serviceHandler.Live)
	router.GET("/read", serviceHandler.Read)

	// REST server
	srv := &http.Server{
		Addr:    ":" + config.GoServicePort,
		Handler: router,
	}

	go func() {
		log.Printf("REST Server is starting on port:%s\n", config.GoServicePort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error running REST server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive a signal in the quit channel
	<-quit

	log.Println("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
