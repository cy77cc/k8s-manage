package logic

import (
	"context"

	v1 "github.com/cy77cc/k8s-manage/api/node/v1"
	"github.com/cy77cc/k8s-manage/internal/dao/node"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/cy77cc/k8s-manage/internal/utils"
)

type NodeLogic struct {
	svcCtx  *svc.ServiceContext
	nodeDao *node.NodeDao
}

func NewNodeLogic(svcCtx *svc.ServiceContext) *NodeLogic {
	return &NodeLogic{
		svcCtx:  svcCtx,
		nodeDao: node.NewNodeDao(svcCtx.DB, svcCtx.Cache, svcCtx.Rdb),
	}
}

func (n *NodeLogic) CreateNode(ctx context.Context, req v1.CreateNodeReq) (v1.NodeResp, error) {

	node := &model.Node{
		Name:    req.Name,
		IP:      req.IP,
		Port:    req.Port,
		SshUser: req.SSHUser,
		Labels:  utils.MapToString(req.Labels, ","),
	}

	if err := n.nodeDao.Create(ctx, node); err != nil {
		return v1.NodeResp{}, err
	}
	return v1.NodeResp{}, nil
}
