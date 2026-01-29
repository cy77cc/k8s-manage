package logic

import (
	dao "github.com/cy77cc/k8s-manage/internal/dao/user"
	"github.com/cy77cc/k8s-manage/internal/svc"
)

type UserLogic struct {
	svcCtx  *svc.ServiceContext
	userDAO *dao.UserDAO
	whiteListDao *dao.WhiteListDao
}

func NewUserLogic(svcCtx *svc.ServiceContext) *UserLogic {
	return &UserLogic{
		svcCtx:  svcCtx,
		userDAO: dao.NewUserDAO(svcCtx.DB, svcCtx.Cache, svcCtx.Rdb),
		whiteListDao: dao.NewWhiteListDao(svcCtx.DB, svcCtx.Cache, svcCtx.Rdb),
	}
}
