package svc

import (
	"fmt"
	"iotdb-courier/collector/internal/config"
)

type ServiceContext struct {
	Config   config.Config
	Producer *MessageProducer
}

func NewServiceContext(c config.Config) *ServiceContext {
	mp, err := NewMessageProducer(c)
	if err != nil {
		fmt.Println(err)
	}
	return &ServiceContext{
		Config:   c,
		Producer: mp,
	}
}

func (sc *ServiceContext) Destroy() {
	sc.Producer.Close()
}
