package config

import (
	"log"
	"os"

	"github.com/RackSec/srslog"
	"github.com/go-playground/validator"
	"github.com/spf13/viper"
)

type Config struct {
	AppName            string `mapstructure:"APP_NAME" validate:"required"`
	GoServicePort      string `mapstructure:"GO_SERVICE_PORT" validate:"required"`
	SysLog             string `mapstructure:"SYSLOG" validate:"required"`
	Https              string `mapstructure:"HTTPS" validate:"required"`
	Logger             *srslog.Writer
	Test               string   `mapstructure:"TEST" validate:"required"`
	KafkaBrokers       []string `mapstructure:"KAFKA_BROKERS" validate:"required"`
	KafkaConsumerGroup string   `mapstructure:"KAFKA_CONSUMER_GROUP" validate:"required"`
	KafkaConsumeTopics []string `mapstructure:"KAFKA_CONSUME_TOPICS" validate:"required"`
	KafkaProduceTopic  string   `mapstructure:"KAFKA_PRODUCE_TOPIC" validate:"required"`
	RedisURL           string   `mapstructure:"REDIS_URL"`
	WsServerURL        string   `mapstructure:"WSSERVER_URL"`
	WsPingPeriod       int      `mapstructure:"WSPING_PERIOD"`
	WsPinMaxError      int      `mapstructure:"WSPING_MAX_ERROR"`
	NatsURL            []string `mapstructure:"NATS_URL" validate:"required"`
	MaxRetry           int      `mapstructure:"MAX_RETRY"`
	MaxWait            int      `mapstructure:"MAX_WAIT"`
}

var config = &Config{}

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath("./internal/config/")
	viper.SetConfigType("json")

	viper.SetDefault("REDIS_PORT", 6379)
	viper.SetDefault("MAX_RETRY", 5)
	viper.SetDefault("MAX_WAIT", 2000)

	log.Println("Reading config...")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Fatal error config: %v", err)
	}

	log.Println("Unmarshalling config...")
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("Unable to decode into struct, %v", err)
	}

	if config.SysLog == "true" {

		config.Logger, err = srslog.Dial("", "", srslog.LOG_INFO, "CEF0")
		if err != nil {
			log.Println("Error setting up syslog:", err)
			os.Exit(1)
		}

	}

	validate := validator.New()
	err = validate.Struct(config)
	if err != nil {
		log.Fatalf("Config validation failed, %v", err)
	}
}

func GetConfig() (*Config, error) {
	return config, nil
}
