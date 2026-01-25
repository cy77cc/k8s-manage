package logic

import (
	"github.com/cy77cc/k8s-manage/internal/context"
	"github.com/cy77cc/k8s-manage/internal/svc"
)

type userLogic struct {
	svcCtx *svc.ServiceContext
	ctx    *context.Context
}

func NewuserLogic(svcCtx *svc.ServiceContext, ctx *context.Context) *userLogic {
	return &userLogic{
		svcCtx: svcCtx,
		ctx:    ctx,
	}
}
