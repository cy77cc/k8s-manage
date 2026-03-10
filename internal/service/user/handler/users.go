package handler

import (
	"strconv"

	"github.com/cy77cc/OpsPilot/internal/httpx"
	"github.com/cy77cc/OpsPilot/internal/model"
	userLogic "github.com/cy77cc/OpsPilot/internal/service/user/logic"
	"github.com/cy77cc/OpsPilot/internal/svc"
	"github.com/cy77cc/OpsPilot/internal/xcode"
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

// GetUserInfo 获取用户信息
func (u *UserHandler) GetUserInfo(c *gin.Context) {
	idStr := c.Param("id")
	var id model.UserID

	if idInt, err := strconv.Atoi(idStr); err != nil {
		httpx.Fail(c, xcode.ParamError, "invalid id")
		return
	} else {
		id = model.UserID(idInt)
	}
	resp, err := userLogic.NewUserLogic(u.svcCtx).GetUser(c.Request.Context(), id)
	if err != nil {
		httpx.Fail(c, xcode.ServerError, err.Error())
		return
	}
	httpx.OK(c, resp)
}
