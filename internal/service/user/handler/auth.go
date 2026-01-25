package handler

import (
	v1 "github.com/cy77cc/k8s-manage/api/user/v1"
	"github.com/cy77cc/k8s-manage/internal/response"
	userLogic "github.com/cy77cc/k8s-manage/internal/service/user/logic"
	"github.com/cy77cc/k8s-manage/internal/xcode"
	"github.com/gin-gonic/gin"
)

// 登录
func (u *userHandler) Login(c *gin.Context) {
	var req v1.LoginReq
	err := c.ShouldBind(&req)
	if err != nil {
		response.Response(c, nil, xcode.NewErrCode(xcode.ErrInvalidParam))
	}
	resp, err := userLogic.NewuserLogic(u.svcCtx, u.ctx).Login(req)
	if err != nil {
		response.Response(c, nil, xcode.FromError(err))
		return
	}
	response.Response(c, resp, nil)
}

// 注册
func (u *userHandler) Register(c *gin.Context) {
	var req v1.UserCreateReq
	err := c.ShouldBind(&req)
	if err != nil {
		response.Response(c, nil, xcode.NewErrCode(xcode.ErrInvalidParam))
		return
	}
	resp, err := userLogic.NewuserLogic(u.svcCtx, u.ctx).Register(req)
	if err != nil {
		response.Response(c, nil, xcode.FromError(err))
		return
	}
	response.Response(c, resp, nil)
}

// 刷新token
func (u *userHandler) Refresh(c *gin.Context) {

}

// 登出
func (u *userHandler) Logout(c *gin.Context) {

}
