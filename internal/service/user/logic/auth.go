package logic

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	v1 "github.com/cy77cc/k8s-manage/api/user/v1"
	"github.com/cy77cc/k8s-manage/internal/config"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/utils"
	"github.com/cy77cc/k8s-manage/internal/xcode"
	"gorm.io/gorm"
)

// 登录
func (l *UserLogic) Login(ctx context.Context, req v1.LoginReq) (v1.TokenResp, error) {
	// 1. Check user existence
	user, err := l.userDAO.FindOneByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return v1.TokenResp{}, xcode.NewErrCode(xcode.UserNotExist)
		}
		return v1.TokenResp{}, fmt.Errorf("failed to query user: %w", err)
	}

	// 2. Verify password
	if !utils.VerifyPassword(req.Password, user.PasswordHash) {
		return v1.TokenResp{}, xcode.NewErrCode(xcode.PasswordError)
	}

	// 3. Generate Token
	token, err := utils.GenToken(uint(user.ID), false)
	if err != nil {
		return v1.TokenResp{}, fmt.Errorf("failed to generate token: %w", err)
	}

	refreshToken, err := utils.GenToken(uint(user.ID), true)
	if err != nil {
		return v1.TokenResp{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	if err := l.whiteListDao.AddToWhitelist(ctx, refreshToken, time.Now().Add(config.CFG.JWT.RefreshExpire)); err != nil {
		return v1.TokenResp{}, xcode.NewErrCode(xcode.CacheError)
	}

	// 4. Update last login time
	user.LastLoginTime = time.Now().Unix()
	if err := l.userDAO.Update(ctx, user); err != nil {
		// Log error but don't fail login? Or fail?
		// For strict consistency, we might fail or just log.
		// Here we'll just log/ignore for now as we don't have logger injected yet
	}

	roles, permissions, _ := l.loadRolesAndPermissions(ctx, uint64(user.ID))
	return v1.TokenResp{
		AccessToken:  token,
		RefreshToken: refreshToken,
		Expires:      time.Now().Add(config.CFG.JWT.Expire).Unix(), // Should match config
		Uid:          uint64(user.ID),
		Roles:        roles,
		User: &v1.AuthUser{
			Id:          uint64(user.ID),
			Username:    user.Username,
			Name:        user.Username,
			Email:       user.Email,
			Status:      "active",
			Roles:       roles,
			Permissions: permissions,
		},
		Permissions: permissions,
	}, nil
}

// 注册
func (l *UserLogic) Register(ctx context.Context, req v1.UserCreateReq) (v1.TokenResp, error) {
	// 1. Check if user exists
	_, err := l.userDAO.FindOneByUsername(ctx, req.Username)
	if err == nil {
		return v1.TokenResp{}, xcode.NewErrCode(xcode.UserAlreadyExist)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return v1.TokenResp{}, fmt.Errorf("database error: %w", err)
	}

	// 2. Create User
	encryptPwd, err := utils.HashPassword(req.Password)
	if err != nil {
		return v1.TokenResp{}, fmt.Errorf("failed to hash password: %w", err)
	}

	newUser := &model.User{
		Username:     req.Username,
		PasswordHash: encryptPwd, // Plaintext for demo, should be hashed
		Email:        req.Email,
		CreateTime:   time.Now().Unix(),
		UpdateTime:   time.Now().Unix(),
	}

	if err := l.svcCtx.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(newUser).Error; err != nil {
			return err
		}
		var viewerRole model.Role
		if err := tx.Where("LOWER(code) = ?", "viewer").First(&viewerRole).Error; err == nil {
			if err := tx.Create(&model.UserRole{UserID: int64(newUser.ID), RoleID: int64(viewerRole.ID)}).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return v1.TokenResp{}, fmt.Errorf("failed to create user: %w", err)
	}

	// 3. Generate Token
	token, err := utils.GenToken(uint(newUser.ID), false)
	if err != nil {
		return v1.TokenResp{}, fmt.Errorf("failed to generate token: %w", err)
	}

	refreshToken, err := utils.GenToken(uint(newUser.ID), true)
	if err != nil {
		return v1.TokenResp{}, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	if err := l.whiteListDao.AddToWhitelist(ctx, refreshToken, time.Now().Add(config.CFG.JWT.RefreshExpire)); err != nil {
		return v1.TokenResp{}, xcode.NewErrCode(xcode.CacheError)
	}

	roles, permissions, _ := l.loadRolesAndPermissions(ctx, uint64(newUser.ID))
	return v1.TokenResp{
		AccessToken:  token,
		RefreshToken: refreshToken,
		Expires:      time.Now().Add(config.CFG.JWT.Expire).Unix(),
		Uid:          uint64(newUser.ID),
		Roles:        roles,
		User: &v1.AuthUser{
			Id:          uint64(newUser.ID),
			Username:    newUser.Username,
			Name:        newUser.Username,
			Email:       newUser.Email,
			Status:      "active",
			Roles:       roles,
			Permissions: permissions,
		},
		Permissions: permissions,
	}, nil
}

// 刷新token
func (l *UserLogic) Refresh(ctx context.Context, req v1.RefreshReq) (v1.TokenResp, error) {
	// Verify refresh token (simplified)
	// 第一步判断rtoken在不在白名单
	ok, err := l.whiteListDao.IsWhitelisted(ctx, req.RefreshToken)
	if err != nil || !ok {
		return v1.TokenResp{}, xcode.NewErrCode(xcode.TokenInvalid)
	}
	// 解析token，判断过期时间
	claims, err := utils.ParseToken(req.RefreshToken)
	if err != nil {
		return v1.TokenResp{}, xcode.NewErrCode(xcode.TokenExpired)
	}

	// 生成新的atoken和rtoken
	newToken, err := utils.GenToken(claims.Uid, false)
	if err != nil {
		return v1.TokenResp{}, err
	}

	newRefreshToken, err := utils.GenToken(claims.Uid, true)
	if err != nil {
		return v1.TokenResp{}, err
	}

	// 从缓存中删除旧的rtoken，添加新的rtoken
	if err := l.whiteListDao.DeleteToken(ctx, req.RefreshToken); err != nil {
		return v1.TokenResp{}, xcode.NewErrCode(xcode.CacheError)
	}

	if err := l.whiteListDao.AddToWhitelist(ctx, newRefreshToken, time.Now().Add(config.CFG.JWT.RefreshExpire)); err != nil {
		return v1.TokenResp{}, xcode.NewErrCode(xcode.CacheError)
	}

	roles, permissions, _ := l.loadRolesAndPermissions(ctx, uint64(claims.Uid))
	return v1.TokenResp{
		AccessToken:  newToken,
		RefreshToken: newRefreshToken,
		Expires:      time.Now().Add(config.CFG.JWT.Expire).Unix(),
		Uid:          uint64(claims.Uid),
		Roles:        roles,
		Permissions:  permissions,
	}, nil
}

// 登出
func (l *UserLogic) Logout(ctx context.Context, req v1.LogoutReq) error {
	if strings.TrimSpace(req.RefreshToken) == "" {
		return nil
	}
	if err := l.whiteListDao.DeleteToken(ctx, req.RefreshToken); err != nil {
		return xcode.FromError(err)
	}
	return nil
}

func (l *UserLogic) loadRolesAndPermissions(ctx context.Context, userID uint64) ([]string, []string, error) {
	roleRows := make([]struct {
		Code string `gorm:"column:code"`
	}, 0)
	if err := l.svcCtx.DB.WithContext(ctx).
		Table("roles").
		Select("roles.code").
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Scan(&roleRows).Error; err != nil {
		return nil, nil, err
	}
	roles := make([]string, 0, len(roleRows))
	roleSet := make(map[string]struct{}, len(roleRows))
	for _, row := range roleRows {
		code := strings.TrimSpace(row.Code)
		if code == "" {
			continue
		}
		if _, ok := roleSet[code]; ok {
			continue
		}
		roleSet[code] = struct{}{}
		roles = append(roles, code)
	}

	permRows := make([]struct {
		Code string `gorm:"column:code"`
	}, 0)
	if err := l.svcCtx.DB.WithContext(ctx).
		Table("permissions").
		Select("permissions.code").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", userID).
		Scan(&permRows).Error; err != nil {
		return roles, nil, err
	}
	permissions := make([]string, 0, len(permRows)+1)
	permSet := make(map[string]struct{}, len(permRows)+1)
	for _, row := range permRows {
		code := strings.TrimSpace(row.Code)
		if code == "" {
			continue
		}
		if _, ok := permSet[code]; ok {
			continue
		}
		permSet[code] = struct{}{}
		permissions = append(permissions, code)
	}
	for _, roleCode := range roles {
		if strings.EqualFold(roleCode, "admin") {
			if _, ok := permSet["*:*"]; !ok {
				permissions = append(permissions, "*:*")
			}
			break
		}
	}

	return roles, permissions, nil
}
