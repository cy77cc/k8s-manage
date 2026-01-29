package node

import (
	v1 "github.com/cy77cc/k8s-manage/api/node/v1"
	"github.com/cy77cc/k8s-manage/internal/model"
)

// Node 节点管理接口
type Node interface {
	// 创建节点
	CreateNode(v1.CreateNodeReq) (v1.NodeResp, error)
	// 获取节点详情
	GetNode(id model.NodeID) (v1.NodeResp, error)
	// 更新节点
	UpdateNode(v1.UpdateNodeReq) (v1.NodeResp, error)
	// 删除节点
	DeleteNode(id model.NodeID) error
	// 获取节点列表
	ListNodes(v1.ListNodeReq) (v1.ListNodeResp, error)
	
	// 同步节点信息 (从物理机采集信息)
	SyncNodeInfo(id model.NodeID) error
	// 检查节点健康状态
	CheckNodeHealth(id model.NodeID) error

	// 连接终端
	OpenTerm(id model.NodeID)
}

// SSHKey SSH密钥管理接口
type SSHKey interface {
	CreateSSHKey(v1.CreateSSHKeyReq) (v1.SSHKeyResp, error)
	ListSSHKeys(v1.ListSSHKeyReq) (v1.ListSSHKeyResp, error)
	DeleteSSHKey(id model.NodeID) error
}

// NodeOperation 节点运维接口
type NodeOperation interface {
	// 批量操作
	BatchOperation(v1.NodeBatchOpReq) error
	// 获取节点Web Shell连接地址
	GetShellURL(v1.NodeShellReq) (string, error)
}
