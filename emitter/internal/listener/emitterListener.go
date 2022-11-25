package listener

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/apache/iotdb-client-go/client"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	cluster "github.com/bsm/sarama-cluster"
	"iotdb-courier/emitter/internal/config"
	"iotdb-courier/emitter/internal/logic"
	"iotdb-courier/emitter/internal/types"
	"time"
)

type Consumer interface {
	Init(*logic.EmitterLogic)
	Close()
}

type EmitterListener struct {
	Config   config.Config
	consumer Consumer
	session  client.Session
}

func NewEmitterListener(c config.Config) *EmitterListener {
	return &EmitterListener{
		Config: c,
	}
}

func (e *EmitterListener) Start() {
	// init iotdb session
	config := &client.ClusterConfig{
		NodeUrls: e.Config.IoTDB.NodeUrl,
		UserName: e.Config.IoTDB.Username,
		Password: e.Config.IoTDB.Password,
	}
	e.session = client.NewClusterSession(config)
	if err := e.session.OpenCluster(false); err != nil {
		panic(err)
	}
	// init consumer
	e.consumer = e.initConsumer()
	e.consumer.Init(logic.NewEmitterLogic(&e.Config, &e.session))
}

func (e *EmitterListener) Stop() {
	if e.consumer != nil {
		e.consumer.Close()
	}
	e.session.Close()
}

type KafkaConsumer struct {
	Config   config.Config
	consumer *cluster.Consumer
}

func (kc *KafkaConsumer) Init(l *logic.EmitterLogic) {
	conf := cluster.NewConfig()
	conf.Group.Return.Notifications = true
	conf.Consumer.Offsets.CommitInterval =
		time.Duration(kc.Config.Kafka.CommitIntervalSec) * time.Second
	conf.Consumer.Offsets.Initial = sarama.OffsetNewest
	con, err := cluster.NewConsumer(kc.Config.Kafka.BootstrapServer,
		kc.Config.Kafka.GroupId, []string{kc.Config.UsingTopic}, conf)
	if err != nil {
		panic(err)
	}
	kc.consumer = con
	go func(c *cluster.Consumer) {
		errors := c.Errors()
		// conf.Group.Return.Notifications = true 时必须及时消费c.Notifications channel
		notifications := c.Notifications()
		for {
			select {
			case err := <-errors:
				fmt.Printf("cluster consumer error: %v", err)
			case <-notifications:
			}
		}
	}(con)
	for msg := range con.Messages() {
		message := types.Message{
			Partition: msg.Partition,
			Offset:    msg.Offset,
			Key:       string(msg.Key),
			Value:     msg.Value,
		}
		if l.Emitter(message) {
			con.MarkOffset(msg, "")
		}
	}
}

func (kc *KafkaConsumer) Close() {
	if kc.consumer != nil {
		kc.consumer.Close()
	}
}

type RocketMQConsumer struct {
	Config   config.Config
	Consumer rocketmq.PushConsumer
	done     chan int
}

func (rc *RocketMQConsumer) Init(l *logic.EmitterLogic) {
	rc.Consumer, _ = rocketmq.NewPushConsumer(
		consumer.WithNameServer(rc.Config.RocketMQ.NameServer), // 接入点地址
		consumer.WithConsumerModel(consumer.Clustering),
		consumer.WithGroupName(rc.Config.RocketMQ.GroupName), // 分组名称
	)
	err := rc.Consumer.Subscribe(rc.Config.UsingTopic, consumer.MessageSelector{},
		func(ctx context.Context, msg ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
			for _, v := range msg {
				message := types.Message{
					Partition: 0,
					Offset:    v.QueueOffset,
					Key:       v.GetKeys(),
					Tag:       v.GetTags(),
					Value:     v.Body,
				}
				l.Emitter(message)
			}
			return consumer.ConsumeSuccess, nil
		})
	if err != nil {
		return
	}
	rc.Consumer.Start()
	<-rc.done
}

func (rc *RocketMQConsumer) Close() {
	rc.Consumer.Shutdown()
	rc.done <- 1
}

func (e *EmitterListener) initConsumer() (con Consumer) {
	if e.Config.UsingQueue == "kafka" {
		con = &KafkaConsumer{
			Config: e.Config,
		}
	} else if e.Config.UsingQueue == "rocketmq" {
		con = &RocketMQConsumer{
			Config: e.Config,
		}
	}
	return
}
