package logic

import (
	"context"
	"fmt"
	"iotdb-courier/collector/internal/svc"
	"iotdb-courier/collector/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CollectorLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCollectorLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CollectorLogic {
	return &CollectorLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CollectorLogic) Accept(req *types.AcceptRequest, json *[]byte) (*types.Response, error) {
	err := l.svcCtx.Producer.SendWithKey(&req.Series, &req.Key, *json)
	if err != nil {
		return &types.Response{
			Code:    500,
			Message: fmt.Sprintf("series: [%s], key: [%s], content: [%s]", req.Series, req.Key, *json),
		}, err
	}
	return &types.Response{
		Code: 200,
	}, nil

}
