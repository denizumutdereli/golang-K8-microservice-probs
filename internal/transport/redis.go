package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/denizumutdereli/golang-K8-microservice-probs/internal/config"
	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type RedisClient struct {
	Client           *redis.Client
	config           *config.Config
	connectionStatus bool
}

func NewRedisClient(redisURL string, cnf *config.Config) (*RedisClient, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %v", err)
	}

	client := redis.NewClient(options)

	_, err = client.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %v", err)
	}

	redisClient := &RedisClient{
		Client:           client,
		connectionStatus: true,
		config:           cnf,
	}

	go redisClient.MonitorConnection()

	return redisClient, nil
}

func (r *RedisClient) IsConnected() bool {
	return r.connectionStatus
}

func (r *RedisClient) hearthbeat() error {
	_, err := r.Client.Ping(context.Background()).Result()
	if err != nil {
		r.connectionStatus = false
		return fmt.Errorf("failed to connect to Redis: %v", err)
	}

	r.connectionStatus = true
	return nil
}

func (r *RedisClient) MonitorConnection() {
	log.Println("Entered Redis MonitorConnection", r.connectionStatus)

	retries := 0
	for {
		waitTime := time.Duration(retries*r.config.MaxWait) * time.Second

		time.Sleep(waitTime)

		err := r.hearthbeat()
		if err != nil {
			log.Printf("Reconnection to Redis failed: %v", err)
			retries++
			if retries >= r.config.MaxRetry {
				log.Fatal("Max retries reached for Redis connection. Exiting.")
				return
			}
		} else {
			if retries > 0 {
				log.Println("Re-connected to Redis successfully")
			}
			retries = 0
		}
	}
}

func (r *RedisClient) SetKeyValue(key string, value interface{}, expiration ...time.Duration) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	expire := time.Duration(0) // Default no expiration
	if len(expiration) > 0 {
		expire = expiration[0]
	}

	err = r.Client.Set(ctx, key, jsonValue, expire).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisClient) GetKeyValue(key string) (interface{}, error) {
	val, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var data interface{}
	err = json.Unmarshal([]byte(val), &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (r *RedisClient) PushList(key string, value interface{}) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	err = r.Client.LPush(ctx, key, jsonValue).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisClient) TrimList(key string, start, stop int64) error {
	err := r.Client.LTrim(ctx, key, start, stop).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisClient) DeleteKey(key string) error {
	err := r.Client.Del(ctx, key).Err()
	if err != nil {
		return err
	}

	return nil
}

func (r *RedisClient) GetList(key string, start, stop int64) ([]interface{}, error) {
	vals, err := r.Client.LRange(ctx, key, start, stop).Result()
	if err != nil {
		return nil, err
	}

	var results []interface{}
	for _, val := range vals {
		var data interface{}
		err = json.Unmarshal([]byte(val), &data)
		if err != nil {
			return nil, err
		}

		results = append(results, data)
	}

	return results, nil
}

func (r *RedisClient) GetAll(prefix string) (map[string]interface{}, error) {
	keys, err := r.Client.Keys(ctx, prefix+"*").Result()
	if err != nil {
		return nil, err
	}

	results := make(map[string]interface{})
	for _, key := range keys {
		val, err := r.Client.Get(ctx, key).Result()
		if err != nil {
			return nil, err
		}

		var data interface{}
		err = json.Unmarshal([]byte(val), &data)
		if err != nil {
			return nil, err
		}

		results[key] = data
	}

	return results, nil
}
