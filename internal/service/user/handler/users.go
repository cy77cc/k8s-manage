package handler

import (
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/response"
	userLogic "github.com/cy77cc/k8s-manage/internal/service/user/logic"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/cy77cc/k8s-manage/internal/xcode"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	svcCtx *svc.ServiceContext
}

func NewUserHandler(svcCtx *svc.ServiceContext) *UserHandler {
	return &UserHandler{
		svcCtx: svcCtx,
	}
}

// Get 获取用户信息
// @Summary 获取用户信息
// @Description 获取用户信息
// @BasePath /api/v1
// @Tags User
// @Accept json
// @Produce json
// @Success 200 {object} response.Resp "登录成功"
// @Router /user/:id [get]
func (u *UserHandler) GetUserInfo(c *gin.Context) {
	var id model.UserID
	if err := c.ShouldBindUri(&id); err != nil {
		response.Response(c, nil, xcode.NewErrCode(xcode.ErrInvalidParam))
		return
	}
	resp, err := userLogic.NewUserLogic(u.svcCtx).GetUser(c.Request.Context(), id)
	if err != nil {
		response.Response(c, nil, xcode.FromError(err))
		return
	}

	response.Response(c, resp, nil)
}
