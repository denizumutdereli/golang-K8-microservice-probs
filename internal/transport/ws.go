package transport

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type WSClient struct {
	URL          *url.URL
	Connection   *websocket.Conn
	PingPeriod   time.Duration
	MaxPingError int
	MaxRetry     int
	RetryWait    time.Duration
	connected    bool
}

func NewWSClient(u string, pingPeriod time.Duration, maxPingError int, maxRetry int, retryWait time.Duration) *WSClient {
	wsurl, err := url.Parse(u)
	if err != nil {
		log.Fatalf("Failed to parse WebSocket server URL: %v", err)
	}

	return &WSClient{
		URL:          wsurl,
		PingPeriod:   pingPeriod,
		MaxPingError: maxPingError,
		MaxRetry:     maxRetry,
		RetryWait:    retryWait,
		connected:    false,
	}
}

func (w *WSClient) Connect() error {
	var err error

	for i := 0; i < w.MaxRetry; i++ {
		w.Connection, _, err = websocket.DefaultDialer.Dial(w.URL.String(), nil)
		if err != nil {
			w.connected = false
			waitTime := time.Duration(i) * w.RetryWait * time.Second
			fmt.Printf("Failed to connect wsserver, retry %d/%d, waiting %v before retrying\n", i+1, w.MaxRetry, waitTime)
			time.Sleep(waitTime)
		} else {
			w.connected = true
			break
		}
	}

	if err != nil {
		return fmt.Errorf("failed to connect wsserver after %d retries: %v", w.MaxRetry, err)
	}

	return nil
}

func (w *WSClient) IsConnected() bool {
	return w.connected
}

func (w *WSClient) MonitorConnection() {
	ticker := time.NewTicker(w.PingPeriod)
	defer ticker.Stop()

	pingFailures := 0
	retries := 0

	for {
		select {
		case <-ticker.C:
			err := w.Connection.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				log.Println("Failed to send ping:", err)
				pingFailures++
				if pingFailures >= w.MaxPingError {
					w.Connection.Close()
					log.Println("Max ping failures reached, reconnecting...")
					err = w.Connect()
					if err != nil {
						retries++
						if retries >= w.MaxRetry {
							log.Println("Failed to reconnect after max retries")
							return
						}
					} else {
						log.Println("Re-connected to WebSocket successfully")
						pingFailures = 0
						retries = 0
					}
				}
			} else {
				pingFailures = 0
			}
		}
	}
}
