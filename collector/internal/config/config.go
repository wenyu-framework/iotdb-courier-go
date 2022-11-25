package config

import "github.com/zeromicro/go-zero/rest"

type RocketMQConfig struct {
	NameServer []string
	Retry      int    `json:", default=2"`
	GroupName  string `json:", default=ProducerGroup-1"`
}

type KafkaConfig struct {
	BootstrapServer []string
}

type Config struct {
	rest.RestConf
	UsingTopic string
	UsingQueue string
	RocketMQ   RocketMQConfig
	Kafka      KafkaConfig
}
