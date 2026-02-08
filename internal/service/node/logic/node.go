package logic

import (
	"context"
	"strconv"
	"strings"

	v1 "github.com/cy77cc/k8s-manage/api/node/v1"
	client "github.com/cy77cc/k8s-manage/internal/client/ssh"
	"github.com/cy77cc/k8s-manage/internal/dao/node"
	"github.com/cy77cc/k8s-manage/internal/model"
	"github.com/cy77cc/k8s-manage/internal/svc"
	"github.com/cy77cc/k8s-manage/internal/utils"
	"golang.org/x/crypto/ssh"
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

type nodeProbeResult struct {
	Hostname string
	OS       string
	Arch     string
	Kernel   string
	CpuCores int
	MemoryMB int
	DiskGB   int
	Status   string
}

func probeNode(cli *ssh.Client) (*nodeProbeResult, error) {
	cmd := `
echo "hostname=$(hostname)"
echo "os=$(cat /etc/os-release | grep PRETTY_NAME | cut -d= -f2 | tr -d '"')"
echo "arch=$(uname -m)"
echo "kernel=$(uname -r)"
echo "cpu=$(nproc)"
echo "mem=$(free -m | awk '/Mem:/ {print $2}')"
echo "disk=$(df -BG / | tail -1 | awk '{print $2}' | tr -d G)"
`

	out, err := client.RunCommand(cli, cmd)
	if err != nil {
		return nil, err
	}

	result := &nodeProbeResult{Status: "Ready"}

	for _, line := range strings.Split(out, "\n") {
		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "hostname":
			result.Hostname = kv[1]
		case "os":
			result.OS = kv[1]
		case "arch":
			result.Arch = kv[1]
		case "kernel":
			result.Kernel = kv[1]
		case "cpu":
			result.CpuCores, _ = strconv.Atoi(kv[1])
		case "mem":
			result.MemoryMB, _ = strconv.Atoi(kv[1])
		case "disk":
			result.DiskGB, _ = strconv.Atoi(kv[1])
		}
	}

	return result, nil
}

func (l *NodeLogic) CreateNode(ctx context.Context, req v1.CreateNodeReq) (v1.NodeResp, error) {

	var sshKeyID *model.NodeID
	if req.SSHKeyID != 0 {
		sshKeyID = &req.SSHKeyID
	}

	node := &model.Node{
		Name:        req.Name,
		IP:          req.IP,
		Port:        req.Port,
		SSHUser:     req.SSHUser,
		SSHPassword: req.SSHPassword,
		SSHKeyID:    sshKeyID,
		Labels:      utils.MapToString(req.Labels, ","),
	}

	var privateKey string
	if node.SSHKeyID != nil {
		SSHKey, err := l.nodeDao.FindSSHKeyByID(ctx, *node.SSHKeyID)
		if err != nil {
			return v1.NodeResp{}, err
		}
		privateKey = SSHKey.PrivateKey
	}

	cli, err := client.NewSSHClient(node.SSHUser, node.SSHPassword, node.IP, node.Port, privateKey)
	if err != nil {
		return v1.NodeResp{}, err
	}

	res, err := probeNode(cli)
	if err != nil {
		return v1.NodeResp{}, err
	}

	node.Hostname = res.Hostname
	node.OS = res.OS
	node.Arch = res.Arch
	node.Kernel = res.Kernel
	node.CpuCores = res.CpuCores
	node.MemoryMB = res.MemoryMB
	node.DiskGB = res.DiskGB
	node.Status = res.Status

	if err := l.nodeDao.Create(ctx, node); err != nil {
		return v1.NodeResp{}, err
	}
	return v1.NodeResp{
		ID:          node.ID,
		Name:        node.Name,
		Hostname:    node.Hostname,
		Description: node.Description,
		IP:          node.IP,
		Port:        node.Port,
		SSHUser:     node.SSHUser,
		Arch:        node.Arch,
		Kernel:      node.Kernel,
		CPUCores:    node.CpuCores,
		MemoryMB:    node.MemoryMB,
		DiskGB:      node.DiskGB,
		Status:      node.Status,
		Labels:      node.Labels,
	}, nil
}

// 获取节点详情
func (l *NodeLogic) GetNode(id model.NodeID) (v1.NodeResp, error) {
	return v1.NodeResp{}, nil
}

// 更新节点
func (l *NodeLogic) UpdateNode(v1.UpdateNodeReq) (v1.NodeResp, error) {
	return v1.NodeResp{}, nil
}

// 删除节点
func (l *NodeLogic) DeleteNode(id model.NodeID) error {
	return nil
}

// 获取节点列表
func (l *NodeLogic) ListNodes(v1.ListNodeReq) (v1.ListNodeResp, error) {
	return v1.ListNodeResp{}, nil
}

// 同步节点信息 (从物理机采集信息)
func (l *NodeLogic) SyncNodeInfo(id model.NodeID) error {
	return nil
}

// 检查节点健康状态
func (l *NodeLogic) CheckNodeHealth(id model.NodeID) error {
	return nil
}

// 连接终端
func (l *NodeLogic) OpenTerm(id model.NodeID) {

}
