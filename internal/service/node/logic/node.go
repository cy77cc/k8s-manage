package logic

import (
	"context"

	v1 "github.com/cy77cc/k8s-manage/api/node/v1"
	"github.com/cy77cc/k8s-manage/internal/dao/node"
	"github.com/cy77cc/k8s-manage/internal/svc"
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
	return v1.NodeResp{}, nil
}