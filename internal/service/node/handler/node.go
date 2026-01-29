package handler

import (
	v1 "github.com/cy77cc/k8s-manage/api/node/v1"
	"github.com/cy77cc/k8s-manage/internal/response"
	nodeLogic "github.com/cy77cc/k8s-manage/internal/service/node/logic"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/cy77cc/k8s-manage/internal/xcode"
	"github.com/gin-gonic/gin"
)

type NodeHandler struct {
	svcCtx *svc.ServiceContext
}

func NewNodeHandler(svcCtx *svc.ServiceContext) *NodeHandler {
	return &NodeHandler{
		svcCtx: svcCtx,
	}
}

// Add 添加一个节点
func (n *NodeHandler) Add(c *gin.Context) {
	var req v1.CreateNodeReq
	if err := c.BindJSON(&req); err != nil {
		response.Response(c, nil, xcode.NewErrCode(xcode.ErrInvalidParam))
		return
	}
	resp, err := nodeLogic.NewNodeLogic(n.svcCtx).CreateNode(c.Request.Context(), req)
	if err != nil {
		response.Response(c, nil, xcode.FromError(err))
	}
	response.Response(c, resp, nil)
}
