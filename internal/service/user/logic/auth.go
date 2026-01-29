package logic

import (
	"context"
	"errors"
	"fmt"
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

	// 2. Verify password (mock implementation)
	// In production, use bcrypt.CompareHashAndPassword
	if user.PasswordHash != req.Password {
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

	// 4. Update last login time
	user.LastLoginTime = time.Now().Unix()
	if err := l.userDAO.Update(ctx, user); err != nil {
		// Log error but don't fail login? Or fail?
		// For strict consistency, we might fail or just log.
		// Here we'll just log/ignore for now as we don't have logger injected yet
	}

	return v1.TokenResp{
		AccessToken:  token,
		RefreshToken: refreshToken,
		Expires:      time.Now().Add(time.Hour * 24).Unix(), // Should match config
		Uid:          uint64(user.ID),
		Roles:        []string{}, // TODO: Fetch roles
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
	// In production, use bcrypt.GenerateFromPassword

	encryptPwd, err := utils.EncryptPassword(req.Password)
	if err != nil {
		return v1.TokenResp{}, err
	}

	newUser := &model.User{
		Username:     req.Username,
		PasswordHash: encryptPwd, // Plaintext for demo, should be hashed
		Email:        req.Email,
		CreateTime:   time.Now().Unix(),
		UpdateTime:   time.Now().Unix(),
	}

	if err := l.userDAO.Create(ctx, newUser); err != nil {
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

	return v1.TokenResp{
		AccessToken:  token,
		RefreshToken: refreshToken,
		Expires:      time.Now().Add(time.Hour * 24).Unix(),
		Uid:          uint64(newUser.ID),
		Roles:        []string{},
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

	return v1.TokenResp{
		AccessToken:  newToken,
		RefreshToken: newRefreshToken,
		Expires:      time.Now().Add(time.Hour * 24).Unix(),
		Uid:          uint64(claims.Uid),
		Roles:        []string{},
	}, nil
}

// 登出
func (l *UserLogic) Logout(ctx context.Context, req v1.LogoutReq) error {
	// In a stateless JWT system, logout usually means blacklisting the token.
	// For now, we just return success as we haven't implemented blacklist.
	// 用白名单机制
	return xcode.FromError(l.whiteListDao.DeleteToken(ctx, req.RefreshToken))
}
