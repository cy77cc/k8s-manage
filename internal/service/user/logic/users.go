package logic

import (
	"context"

	v1 "github.com/cy77cc/k8s-manage/api/user/v1"
	dao "github.com/cy77cc/k8s-manage/internal/dao/user"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
)

type UserLogic struct {
	svcCtx       *svc.ServiceContext
	userDAO      *dao.UserDAO
	whiteListDao *dao.WhiteListDao
}

func NewUserLogic(svcCtx *svc.ServiceContext) *UserLogic {
	return &UserLogic{
		svcCtx:       svcCtx,
		userDAO:      dao.NewUserDAO(svcCtx.DB, svcCtx.Cache, svcCtx.Rdb),
		whiteListDao: dao.NewWhiteListDao(svcCtx.DB, svcCtx.Cache, svcCtx.Rdb),
	}
}

func (l *UserLogic) GetUser(ctx context.Context, id model.UserID) (v1.UserResp, error) {
	user, err := l.userDAO.FindOneById(ctx, id)
	if err != nil {
		return v1.UserResp{}, err
	}
	return v1.UserResp{
		Id:            uint64(user.ID),
		Username:      user.Username,
		Email:         user.Email,
		Phone:         user.Phone,
		Avatar:        user.Avatar,
		Status:        int32(user.Status),
		CreateTime:    user.CreateTime,
		UpdateTime:    user.UpdateTime,
		LastLoginTime: user.LastLoginTime,
	}, nil
}
