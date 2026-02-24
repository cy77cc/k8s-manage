package handler

import (
	"errors"
	v1 "github.com/cy77cc/k8s-manage/api/node/v1"
	"github.com/cy77cc/k8s-manage/internal/response"
	hostlogic "github.com/cy77cc/k8s-manage/internal/service/host/logic"
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
	createReq := hostlogic.CreateReq{
		Name:        req.Name,
		IP:          req.IP,
		Port:        req.Port,
		Username:    req.SSHUser,
		Password:    req.SSHPassword,
		Description: req.Description,
		Role:        req.Role,
		ClusterID:   uint(req.ClusterID),
		Force:       true,
		Status:      "offline",
	}
	if req.SSHKeyID != 0 {
		keyID := uint64(req.SSHKeyID)
		createReq.SSHKeyID = &keyID
	}
	node, err := hostlogic.NewHostService(n.svcCtx).CreateWithProbe(c.Request.Context(), 0, true, createReq)
	if err != nil {
		response.Response(c, nil, xcode.FromError(err))
		return
	}
	resp := v1.NodeResp{
		ID:          node.ID,
		Name:        node.Name,
		Hostname:    node.Hostname,
		Description: node.Description,
		IP:          node.IP,
		Port:        node.Port,
		SSHUser:     node.SSHUser,
		OS:          node.OS,
		Arch:        node.Arch,
		Kernel:      node.Kernel,
		CPUCores:    node.CpuCores,
		MemoryMB:    node.MemoryMB,
		DiskGB:      node.DiskGB,
		Status:      node.Status,
		Role:        node.Role,
		ClusterID:   int64(node.ClusterID),
		Labels:      node.Labels,
		LastCheckAt: node.LastCheckAt,
		CreatedAt:   node.CreatedAt,
		UpdatedAt:   node.UpdatedAt,
	}
	response.Response(c, resp, nil)
}

func (n *NodeHandler) Get(c *gin.Context) {
	response.Response(c, nil, xcode.FromError(errors.New("not implemented")))
}
