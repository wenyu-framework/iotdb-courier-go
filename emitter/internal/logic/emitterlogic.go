package logic

import (
	"context"
	"fmt"
	"github.com/apache/iotdb-client-go/client"
	"github.com/zeromicro/go-zero/core/logx"
	"iotdb-courier/emitter/internal/config"
	"iotdb-courier/emitter/internal/types"
	"time"
)

type EmitterLogic struct {
	config *config.Config
	logx.Logger
	session   *client.Session
	recordSet *types.RecordSet
}

func NewEmitterLogic(c *config.Config, session *client.Session) *EmitterLogic {
	rcSet := &types.RecordSet{}
	// 定时定批插入
	if c.IoTDB.EnableBatchCommit {
		go func(rs *types.RecordSet) {
			tick := time.NewTicker(time.Duration(c.IoTDB.CommitInterval) * time.Second)
			for {
				select {
				case <-tick.C:
					rcSet.Insert(session)
				}
			}
		}(rcSet)
	}
	return &EmitterLogic{
		Logger:    logx.WithContext(context.Background()),
		config:    c,
		session:   session,
		recordSet: rcSet,
	}
}

func (l *EmitterLogic) Emitter(message types.Message) (acked bool) {
	err := l.recordSet.AddRecord(&message, l.config)
	if err != nil {
		return false
	}
	if !l.config.IoTDB.EnableBatchCommit ||
		l.recordSet.Len() >= l.config.IoTDB.MaxCommitBatchSize {
		fmt.Println("Max commit batch size exceed.")
		l.recordSet.Insert(l.session)
	}
	return true
}
