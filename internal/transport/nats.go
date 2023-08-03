package transport

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

type NatsClient struct {
	nc      *nats.Conn
	wg      sync.WaitGroup
	subject string
}

/*
nm := NewNatsClient([]string{"nats://nats1:4222", "nats://nats2:4222", "nats://nats3:4222"})
*/

func NewNatsClient(servers []string) (*NatsClient, error) {
	opts := []nats.Option{nats.Name("NATS Manager")}
	opts = setupConnOptions(opts)

	serversStr := strings.Join(servers, ",")

	// Connect to NATS
	nc, err := nats.Connect(serversStr, opts...)
	if err != nil {
		return nil, err
	}

	return &NatsClient{nc: nc}, nil
}

func (nm *NatsClient) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	nm.wg.Add(1)

	sub, err := nm.nc.Subscribe(subject, func(m *nats.Msg) {
		handler(m)
		nm.wg.Done()
	})

	if err != nil {
		return nil, err
	}

	return sub, nil
}

func (nm *NatsClient) Publish(subject string, data []byte) error {
	return nm.nc.Publish(subject, data)
}

func (nm *NatsClient) Close() {
	// Gracefully close the connection
	nm.nc.Drain()

	// Wait for all subscriptions to finish
	nm.wg.Wait()

	nm.nc.Close()
}

func setupConnOptions(opts []nats.Option) []nats.Option {
	totalWait := 10 * time.Minute
	reconnectDelay := time.Second

	opts = append(opts, nats.ReconnectWait(reconnectDelay))
	opts = append(opts, nats.MaxReconnects(int(totalWait/reconnectDelay)))
	opts = append(opts, nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
		log.Printf("Got disconnected! Reason: %q\n", err)
	}))
	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		log.Printf("Got reconnected to %v!\n", nc.ConnectedUrl())
	}))
	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		log.Printf("Connection closed. Reason: %q\n", nc.LastError())
	}))
	return opts
}
