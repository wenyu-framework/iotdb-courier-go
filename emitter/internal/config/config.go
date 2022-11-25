package config

type RocketMQConfig struct {
	NameServer []string
	Retry      int    `json:",default=2"`
	GroupName  string `json:",default=ConsumerGroup-1"`
}

type KafkaConfig struct {
	BootstrapServer   []string
	CommitIntervalSec int64  `json:",default=1"`
	GroupId           string `json:",default=ConsumerGroup-1"`
}

type IoTDBConfig struct {
	NodeUrl            []string
	Username           string
	Password           string
	BasePath           string
	TimeParse          string `json:",default=2006-01-02T15:04:05.000"`
	TimeField          string `json:",default=time"`
	TimeUnit           string `json:",default=ms"`
	TimeZone           string `json:",default=Asia/Shanghai"`
	OverwriteDeviceId  string `json:",optional"`
	EnableBatchCommit  bool   `json:",default=false"`
	CommitInterval     int    `json:",default=1"`
	MaxCommitBatchSize int    `json:",default=100"`
}

type Config struct {
	Name       string
	UsingTopic string
	UsingQueue string
	IoTDB      IoTDBConfig
	RocketMQ   RocketMQConfig
	Kafka      KafkaConfig
}
