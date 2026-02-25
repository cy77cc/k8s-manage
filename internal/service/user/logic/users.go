package logic

import (
	"context"
	"fmt"

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

func (l *UserLogic) GetMe(ctx context.Context, uid any) (map[string]any, error) {
	var userID model.UserID
	switch v := uid.(type) {
	case uint:
		userID = model.UserID(v)
	case uint64:
		userID = model.UserID(v)
	case int:
		userID = model.UserID(v)
	case int64:
		userID = model.UserID(v)
	case float64:
		userID = model.UserID(v)
	default:
		return nil, fmt.Errorf("invalid uid type")
	}

	user, err := l.userDAO.FindOneById(ctx, userID)
	if err != nil {
		return nil, err
	}
	roles, permissions, err := l.loadRolesAndPermissions(ctx, uint64(user.ID))
	if err != nil {
		roles = []string{}
		permissions = []string{}
	}
	return map[string]any{
		"id":          user.ID,
		"username":    user.Username,
		"name":        user.Username,
		"email":       user.Email,
		"status":      "active",
		"roles":       roles,
		"permissions": permissions,
	}, nil
}
