package handler

import (
	"github.com/cy77cc/k8s-manage/internal/svc"
)

type UserHandler struct {
	svcCtx *svc.ServiceContext
}

func NewUserHandler(svcCtx *svc.ServiceContext) *UserHandler {
	return &UserHandler{
		svcCtx: svcCtx,
	}
}
