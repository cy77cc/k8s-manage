package users

import (
	"github.com/cy77cc/k8s-manage/internal/context"
	"github.com/cy77cc/k8s-manage/internal/svc"
)

type UserService struct {
	SvcCtx *svc.ServiceContext
	Ctx    *context.Context
}
