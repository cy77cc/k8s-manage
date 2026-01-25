package handler

import (
	"github.com/cy77cc/k8s-manage/internal/context"
	"github.com/cy77cc/k8s-manage/internal/svc"
)

type userHandler struct {
	svcCtx *svc.ServiceContext
	ctx    *context.Context
}

func NewuserHandler(svcCtx *svc.ServiceContext) *userHandler {
	return &userHandler{
		svcCtx: svcCtx,
		ctx:    context.NewContext(),
	}
}
