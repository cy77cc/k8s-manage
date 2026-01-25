package logic

import v1 "github.com/cy77cc/k8s-manage/api/user/v1"

// 登录
func (l *userLogic) Login(v1.LoginReq) (v1.TokenResp, error) {
	return v1.TokenResp{}, nil
}

// 注册
func (l *userLogic) Register(v1.UserCreateReq) (v1.TokenResp, error) {
	return v1.TokenResp{}, nil
}

// 刷新token
func (l *userLogic) Refresh(v1.RefreshReq) (v1.TokenResp, error) {
	return v1.TokenResp{}, nil
}

// 登出
func (l *userLogic) Logout(v1.LogoutReq) error {
	return nil
}
