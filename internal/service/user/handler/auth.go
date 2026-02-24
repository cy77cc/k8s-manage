package handler

import (
	v1 "github.com/cy77cc/k8s-manage/api/user/v1"
	"github.com/cy77cc/k8s-manage/internal/response"
	userLogic "github.com/cy77cc/k8s-manage/internal/service/user/logic"
	"github.com/cy77cc/k8s-manage/internal/xcode"
	"github.com/gin-gonic/gin"
)

// Login 登录
// @Summary 用户登录
// @Description 用户登录接口
// @BasePath /api/v1
// @Tags Auth
// @Accept json
// @Produce json
// @Param req body v1.LoginReq true "登录请求参数"
// @Success 200 {object} response.Resp "登录成功"
// @Router /auth/login [post]
func (u *UserHandler) Login(c *gin.Context) {
	var req v1.LoginReq

	if err := c.ShouldBind(&req); err != nil {
		response.Response(c, nil, xcode.NewErrCode(xcode.ErrInvalidParam))
		return
	}
	resp, err := userLogic.NewUserLogic(u.svcCtx).Login(c.Request.Context(), req)
	if err != nil {
		response.Response(c, nil, xcode.FromError(err))
		return
	}
	response.Response(c, resp, nil)
}

// Register 注册
// @Summary 用户注册
// @Description 用户注册接口
// @BasePath /api/v1
// @Tags Auth
// @Accept json
// @Produce json
// @Param req body v1.UserCreateReq true "注册请求参数"
// @Success 200 {object} response.Resp "注册成功"
// @Router /auth/register [post]
func (u *UserHandler) Register(c *gin.Context) {
	var req v1.UserCreateReq
	err := c.ShouldBind(&req)
	if err != nil {
		response.Response(c, nil, xcode.NewErrCode(xcode.ErrInvalidParam))
		return
	}
	resp, err := userLogic.NewUserLogic(u.svcCtx).Register(c.Request.Context(), req)
	if err != nil {
		response.Response(c, nil, xcode.FromError(err))
		return
	}
	response.Response(c, resp, nil)
}

// Refresh 刷新token
// @Summary 刷新Token
// @Description 刷新Token接口
// @BasePath /api/v1
// @Tags Auth
// @Accept json
// @Produce json
// @Param req body v1.RefreshReq true "刷新Token请求参数"
// @Success 200 {object} response.Resp "刷新成功"
// @Router /auth/refresh [post]
func (u *UserHandler) Refresh(c *gin.Context) {
	var req v1.RefreshReq
	err := c.ShouldBind(&req)
	if err != nil {
		response.Response(c, nil, xcode.NewErrCode(xcode.ErrInvalidParam))
		return
	}
	resp, err := userLogic.NewUserLogic(u.svcCtx).Refresh(c.Request.Context(), req)
	if err != nil {
		response.Response(c, nil, xcode.FromError(err))
		return
	}
	response.Response(c, resp, nil)
}

// Logout 登出
// @Summary 用户登出
// @Description 用户登出接口
// @BasePath /api/v1
// @Tags Auth
// @Accept json
// @Produce json
// @Param req body v1.LogoutReq true "登出请求参数"
// @Success 200 {object} response.Resp "登出成功"
// @Router /auth/logout [post]
func (u *UserHandler) Logout(c *gin.Context) {
	var req v1.LogoutReq
	err := c.ShouldBind(&req)
	if err != nil {
		response.Response(c, nil, xcode.NewErrCode(xcode.ErrInvalidParam))
		return
	}
	err = userLogic.NewUserLogic(u.svcCtx).Logout(c.Request.Context(), req)
	if err != nil {
		response.Response(c, nil, xcode.FromError(err))
		return
	}
	response.Response(c, nil, nil)
}

// Me 获取当前登录用户信息
func (u *UserHandler) Me(c *gin.Context) {
	uid, ok := c.Get("uid")
	if !ok {
		response.Response(c, nil, xcode.NewErrCode(xcode.Unauthorized))
		return
	}
	resp, err := userLogic.NewUserLogic(u.svcCtx).GetMe(c.Request.Context(), uid)
	if err != nil {
		response.Response(c, nil, xcode.FromError(err))
		return
	}
	response.Response(c, resp, nil)
}
