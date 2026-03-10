package deployment

import (
	"github.com/cy77cc/OpsPilot/internal/svc"
)

type Logic struct {
	svcCtx *svc.ServiceContext
}

func NewLogic(svcCtx *svc.ServiceContext) *Logic { return &Logic{svcCtx: svcCtx} }
