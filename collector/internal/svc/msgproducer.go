package svc

import (
	"context"
	"errors"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"iotdb-courier/collector/internal/config"
	"time"
)

type MessageProducer struct {
	Type      string
	Topic     string
	RProducer rocketmq.Producer
	KProducer sarama.SyncProducer
}

func NewMessageProducer(c config.Config) (*MessageProducer, error) {
	mq := &MessageProducer{
		Type:  c.UsingQueue,
		Topic: c.UsingTopic,
	}
	if c.UsingQueue == "rocketmq" {
		p, err := rocketmq.NewProducer(
			producer.WithNameServer(c.RocketMQ.NameServer),
			producer.WithRetry(c.RocketMQ.Retry),
			producer.WithGroupName(c.RocketMQ.GroupName),
		)
		if err != nil {
			panic(err)
		}
		p.Start()
		mq.RProducer = p
	} else if c.UsingQueue == "kafka" {
		config := sarama.NewConfig()
		config.Producer.Return.Successes = true
		config.Producer.Timeout = 6 * time.Second
		p, err := sarama.NewSyncProducer(c.Kafka.BootstrapServer, config)
		if err != nil {
			return nil, err
		}
		mq.KProducer = p
	} else {
		return nil, errors.New("unsupported message queue type")
	}
	return mq, nil
}

func (mp *MessageProducer) Close() {
	if mp.RProducer != nil {
		mp.RProducer.Shutdown()
	}
}

func (mp *MessageProducer) Send(series *string, payload []byte) error {
	return mp.SendWithKey(series, nil, payload)
}

func (mp *MessageProducer) SendWithKey(series, key *string, payload []byte) (err error) {
	if mp.Type == "rocketmq" {
		msg := &primitive.Message{
			Topic: mp.Topic,
			Body:  payload,
		}
		msg.WithTag(*series)
		if key != nil && len(*key) > 0 {
			msg.WithKeys([]string{*key})
		}
		sendResult, err := mp.RProducer.SendSync(context.Background(), msg)
		if err == nil && sendResult.Status != primitive.SendOK {
			err = errors.New("send message failed")
		}
	} else if mp.Type == "kafka" {
		var keyEncoder sarama.ByteEncoder
		if key != nil {
			keyEncoder = sarama.ByteEncoder(fmt.Sprintf("%s:%s", *series, *key))
		} else {
			keyEncoder = sarama.ByteEncoder(*series)
		}
		msg := &sarama.ProducerMessage{
			Topic: mp.Topic,
			Key:   keyEncoder,
			Value: sarama.ByteEncoder(payload),
		}
		_, _, err = mp.KProducer.SendMessage(msg)
	}
	return
}
