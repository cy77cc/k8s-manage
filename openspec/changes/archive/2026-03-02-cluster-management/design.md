# Design: 集群管理功能

## 1. 系统架构

### 1.1 服务分层

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Frontend (React)                                │
├─────────────────────────────────────────────────────────────────────────────┤
│  ClusterListPage │ ClusterBootstrapWizard │ ClusterImportWizard             │
│  ClusterDetailPage │ NodeManagement │ ResourceViewer                        │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                                      ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              API Gateway                                     │
│                         /api/v1/clusters/*                                  │
│                         /api/v1/deploy/*                                    │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                    ┌─────────────────┼─────────────────┐
                    ▼                 ▼                 ▼
┌───────────────────────┐ ┌───────────────────┐ ┌───────────────────────┐
│   Cluster Service     │ │ Deployment Service │ │    Host Service       │
├───────────────────────┤ ├───────────────────┤ ├───────────────────────┤
│ - Cluster CRUD        │ │ - Target CRUD     │ │ - Host CRUD           │
│ - Bootstrap Logic     │ │ - Release Logic   │ │ - SSH Operations      │
│ - Node Management     │ │ - Credential      │ │ - Probe/Cloud Import  │
│ - Resource Query      │ │ - Bootstrap Job   │ │                       │
└───────────────────────┘ └───────────────────┘ └───────────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              Data Layer                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│  MySQL: clusters │ cluster_nodes │ cluster_credentials │ cluster_bootstrap_tasks │
│  SSH Client: remote command execution                                        │
│  K8s Client: go-client for K8s API                                          │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 1.2 模块职责

| 模块 | 职责 | 依赖 |
|-----|------|-----|
| cluster service | 集群生命周期管理、节点管理、资源查询 | SSH Client, K8s Client, MySQL |
| deployment service | 部署目标管理、发布流程 | cluster service, MySQL |
| host service | 主机管理、SSH 连接 | MySQL, SSH Client |

## 2. 数据模型详细设计

### 2.1 clusters 表扩展

```go
type Cluster struct {
    ID             uint      `gorm:"primaryKey"`
    Name           string    `gorm:"type:varchar(64);not null;uniqueIndex"`
    Description    string    `gorm:"type:varchar(256)"`
    Version        string    `gorm:"type:varchar(64)"`              // K8s 版本
    Status         string    `gorm:"type:varchar(32);not null"`     // active/inactive/error/provisioning
    Type           string    `gorm:"type:varchar(32);not null"`     // kubernetes/openshift
    Source         string    `gorm:"type:varchar(32);default:'platform_managed'"` // platform_managed/external_managed
    Endpoint       string    `gorm:"type:varchar(256)"`
    KubeConfig     string    `gorm:"type:mediumtext"`               // 逐步废弃，迁移到 credential
    CACert         string    `gorm:"type:text"`
    Token          string    `gorm:"type:text"`
    AuthMethod     string    `gorm:"type:varchar(32)"`              // kubeconfig/cert/token
    CredentialID   *uint     `gorm:"column:credential_id"`          // 关联 cluster_credentials
    K8sVersion     string    `gorm:"type:varchar(32)"`              // 1.28.0
    PodCIDR        string    `gorm:"type:varchar(32)"`              // 10.244.0.0/16
    ServiceCIDR    string    `gorm:"type:varchar(32)"`              // 10.96.0.0/12
    ManagementMode string    `gorm:"type:varchar(32);default:'k8s-only'"`
    LastSyncAt     *time.Time
    CreatedBy      string    `gorm:"type:varchar(64)"`
    UpdatedBy      string    `gorm:"type:varchar(64)"`
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

### 2.2 cluster_nodes 表（新增）

```go
type ClusterNode struct {
    ID               uint      `gorm:"primaryKey"`
    ClusterID        uint      `gorm:"not null;index;uniqueIndex:uk_cluster_name,priority:1"`
    HostID           *uint     `gorm:"index"`                         // 关联 nodes 表
    Name             string    `gorm:"type:varchar(64);not null;uniqueIndex:uk_cluster_name,priority:2"`
    IP               string    `gorm:"type:varchar(45);not null"`
    Role             string    `gorm:"type:varchar(32);not null"`     // control-plane/worker/etcd
    Status           string    `gorm:"type:varchar(32);not null"`     // ready/notready/unknown
    KubeletVersion   string    `gorm:"type:varchar(32)"`
    KubeProxyVersion string    `gorm:"type:varchar(32)"`
    ContainerRuntime string    `gorm:"type:varchar(32)"`
    OSImage          string    `gorm:"type:varchar(128)"`
    KernelVersion    string    `gorm:"type:varchar(64)"`
    AllocatableCPU   string    `gorm:"type:varchar(16)"`
    AllocatableMem   string    `gorm:"type:varchar(16)"`
    AllocatablePods  int       `gorm:"default:0"`
    Labels           string    `gorm:"type:json"`
    Taints           string    `gorm:"type:json"`
    Conditions       string    `gorm:"type:json"`                     // [{type, status, reason, message}]
    JoinedAt         *time.Time
    LastSeenAt       *time.Time
    CreatedAt        time.Time
    UpdatedAt        time.Time
}
```

### 2.3 cluster_bootstrap_tasks 表扩展

```go
type ClusterBootstrapTask struct {
    ID             string    `gorm:"type:varchar(64);primaryKey"`
    Name           string    `gorm:"type:varchar(128);not null"`
    ClusterID      *uint     `gorm:"index"`                         // 新增：关联创建的集群
    ControlPlaneID uint      `gorm:"column:control_plane_host_id;index"`
    WorkerIDsJSON  string    `gorm:"type:longtext"`
    K8sVersion     string    `gorm:"type:varchar(32)"`              // 新增
    CNI            string    `gorm:"type:varchar(32);default:'flannel'"`
    PodCIDR        string    `gorm:"type:varchar(32)"`              // 新增
    ServiceCIDR    string    `gorm:"type:varchar(32)"`              // 新增
    StepsJSON      string    `gorm:"type:longtext"`                 // 新增：步骤执行详情
    Status         string    `gorm:"type:varchar(32);index"`        // queued/running/succeeded/failed
    ResultJSON     string    `gorm:"type:longtext"`
    ErrorMessage   string    `gorm:"type:text"`
    CreatedBy      uint64    `gorm:"index"`
    CreatedAt      time.Time
    UpdatedAt      time.Time
}
```

### 2.4 实体关系

```
┌─────────────┐       ┌──────────────────┐       ┌─────────────────┐
│   Node      │       │   ClusterNode    │       │    Cluster      │
│   (主机)    │       │   (集群节点)     │       │    (集群)       │
├─────────────┤       ├──────────────────┤       ├─────────────────┤
│ id          │◄──────│ host_id          │       │ id              │
│ name        │       │ cluster_id       │──────►│ name            │
│ ip          │       │ name             │       │ source          │
│ cluster_id  │───────│ role             │       │ credential_id   │──┐
│ ...         │       │ status           │       │ ...             │  │
└─────────────┘       │ ...              │       └─────────────────┘  │
                      └──────────────────┘                            │
                                                                      │
                      ┌──────────────────┐                            │
                      │ClusterCredential │◄───────────────────────────┘
                      │   (凭证)         │
                      ├──────────────────┤
                      │ id               │
                      │ cluster_id       │
                      │ source           │
                      │ kubeconfig_enc   │
                      │ ...              │
                      └──────────────────┘
```

## 3. 核心流程设计

### 3.1 自建集群流程

```go
// BootstrapWorkflow 定义自建集群的工作流
type BootstrapWorkflow struct {
    steps []BootstrapStep
}

type BootstrapStep struct {
    Name        string
    Hosts       []string  // control-plane, workers, all
    Script      string
    Timeout     time.Duration
    Rollback    string    // 失败时的回滚脚本
    OnFailure   string    // abort/continue/skip
}

// 默认步骤定义
var defaultBootstrapSteps = []BootstrapStep{
    {Name: "preflight", Hosts: []string{"all"}, Script: "common/preflight.sh", Timeout: 60*time.Second},
    {Name: "containerd", Hosts: []string{"all"}, Script: "common/containerd-install.sh", Timeout: 5*time.Minute, Rollback: "common/containerd-uninstall.sh"},
    {Name: "kubeadm-install", Hosts: []string{"all"}, Script: "kubeadm/v1.28/install.sh", Timeout: 3*time.Minute},
    {Name: "control-plane-init", Hosts: []string{"control-plane"}, Script: "kubeadm/v1.28/init.sh", Timeout: 10*time.Minute, Rollback: "kubeadm/v1.28/reset.sh"},
    {Name: "cni-install", Hosts: []string{"control-plane"}, Script: "cni/calico/v3.26/install.sh", Timeout: 3*time.Minute},
    {Name: "worker-join", Hosts: []string{"workers"}, Script: "kubeadm/v1.28/join.sh", Timeout: 5*time.Minute},
    {Name: "fetch-kubeconfig", Hosts: []string{"control-plane"}, Script: "kubeadm/v1.28/fetch-kubeconfig.sh", Timeout: 30*time.Second},
    {Name: "sync-nodes", Hosts: []string{"control-plane"}, Script: "", Timeout: 30*time.Second}, // 通过 API 同步
}

// 执行流程
func (l *Logic) ExecuteBootstrap(ctx context.Context, task *model.ClusterBootstrapTask) error {
    hosts, err := l.loadHosts(ctx, task.ControlPlaneID, task.WorkerIDs)
    if err != nil {
        return err
    }

    var lastErr error
    for _, step := range defaultBootstrapSteps {
        stepRecord := l.createStepRecord(ctx, task.ID, step.Name)

        err := l.executeStep(ctx, step, hosts)
        if err != nil {
            lastErr = err
            l.failStep(ctx, stepRecord, err)

            // 执行回滚
            if step.Rollback != "" {
                l.executeRollback(ctx, step, hosts)
            }

            // 根据策略决定是否继续
            if step.OnFailure == "abort" {
                break
            }
        } else {
            l.succeedStep(ctx, stepRecord)
        }
    }

    // 创建集群记录
    if lastErr == nil {
        cluster, cred, nodes := l.createClusterRecords(ctx, task, hosts)
        task.ClusterID = &cluster.ID
        task.Status = "succeeded"
    } else {
        task.Status = "failed"
        task.ErrorMessage = lastErr.Error()
    }

    return l.svcCtx.DB.Save(task).Error
}
```

### 3.2 导入集群流程

```go
func (l *Logic) ImportCluster(ctx context.Context, req ImportClusterReq) (*Cluster, error) {
    // 1. 解析 kubeconfig
    config, err := clientcmd.Load([]byte(req.Kubeconfig))
    if err != nil {
        return nil, fmt.Errorf("invalid kubeconfig: %w", err)
    }

    // 2. 获取当前 context
    context := config.Contexts[config.CurrentContext]
    if context == nil {
        return nil, fmt.Errorf("no current context in kubeconfig")
    }

    // 3. 获取 cluster 信息
    cluster := config.Clusters[context.Cluster]
    if cluster == nil {
        return nil, fmt.Errorf("cluster not found in kubeconfig")
    }

    // 4. 构建 REST config 并测试连接
    restConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(req.Kubeconfig))
    if err != nil {
        return nil, fmt.Errorf("failed to build rest config: %w", err)
    }

    client, err := kubernetes.NewForConfig(restConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to create k8s client: %w", err)
    }

    // 5. 获取集群版本
    version, err := client.Discovery().ServerVersion()
    if err != nil {
        return nil, fmt.Errorf("failed to connect to cluster: %w", err)
    }

    // 6. 创建 Cluster 记录
    clusterRecord := &model.Cluster{
        Name:       req.Name,
        Source:     "external_managed",
        Endpoint:   cluster.Server,
        Version:    version.GitVersion,
        Status:     "active",
        Type:       "kubernetes",
        AuthMethod: "kubeconfig",
    }
    if err := l.svcCtx.DB.Create(clusterRecord).Error; err != nil {
        return nil, err
    }

    // 7. 创建 Credential 记录
    cred := &model.ClusterCredential{
        Name:         fmt.Sprintf("%s-credential", req.Name),
        ClusterID:    clusterRecord.ID,
        Source:       "external_managed",
        Endpoint:     cluster.Server,
        AuthMethod:   "kubeconfig",
        Status:       "active",
    }
    if err := l.encryptAndSaveCredential(ctx, cred, req.Kubeconfig); err != nil {
        return nil, err
    }

    // 8. 关联 Credential
    clusterRecord.CredentialID = &cred.ID
    l.svcCtx.DB.Save(clusterRecord)

    // 9. 同步节点信息
    l.syncClusterNodes(ctx, clusterRecord.ID, client)

    return clusterRecord, nil
}
```

### 3.3 资源查询流程

```go
// ClusterClient 管理集群连接
type ClusterClient struct {
    cache map[uint]*kubernetes.Clientset
    mu    sync.RWMutex
}

func (c *ClusterClient) GetClient(clusterID uint) (*kubernetes.Clientset, error) {
    c.mu.RLock()
    if client, ok := c.cache[clusterID]; ok {
        c.mu.RUnlock()
        return client, nil
    }
    c.mu.RUnlock()

    // 从数据库加载凭证
    var cred model.ClusterCredential
    if err := db.Where("cluster_id = ?", clusterID).First(&cred).Error; err != nil {
        return nil, err
    }

    // 构建 REST config
    restConfig, err := buildRestConfigFromCredential(&cred)
    if err != nil {
        return nil, err
    }

    // 创建 client
    client, err := kubernetes.NewForConfig(restConfig)
    if err != nil {
        return nil, err
    }

    // 缓存
    c.mu.Lock()
    c.cache[clusterID] = client
    c.mu.Unlock()

    return client, nil
}

// 资源查询示例
func (l *Logic) ListNamespaces(ctx context.Context, clusterID uint) ([]string, error) {
    client, err := l.clusterClient.GetClient(clusterID)
    if err != nil {
        return nil, err
    }

    namespaces, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
    if err != nil {
        return nil, err
    }

    result := make([]string, 0, len(namespaces.Items))
    for _, ns := range namespaces.Items {
        result = append(result, ns.Name)
    }
    return result, nil
}

func (l *Logic) ListDeployments(ctx context.Context, clusterID uint, namespace string) ([]DeploymentInfo, error) {
    client, err := l.clusterClient.GetClient(clusterID)
    if err != nil {
        return nil, err
    }

    deployments, err := client.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
    if err != nil {
        return nil, err
    }

    result := make([]DeploymentInfo, 0, len(deployments.Items))
    for _, d := range deployments.Items {
        result = append(result, DeploymentInfo{
            Name:      d.Name,
            Namespace: d.Namespace,
            Replicas:  *d.Spec.Replicas,
            Ready:     d.Status.ReadyReplicas,
            Updated:   d.Status.UpdatedReplicas,
            Available: d.Status.AvailableReplicas,
        })
    }
    return result, nil
}
```

## 4. 脚本设计

### 4.1 preflight.sh

```bash
#!/usr/bin/env bash
set -euo pipefail

# 预检查脚本
# 检查系统是否满足 K8s 安装要求

ERRORS=()
WARNINGS=()

# 1. 检查操作系统
check_os() {
    if [[ ! -f /etc/os-release ]]; then
        ERRORS+=("无法确定操作系统版本")
        return
    fi
    . /etc/os-release
    echo "操作系统: $PRETTY_NAME"
}

# 2. 检查内存
check_memory() {
    local mem_total=$(grep MemTotal /proc/meminfo | awk '{print $2}')
    local mem_gb=$((mem_total / 1024 / 1024))
    echo "内存: ${mem_gb}GB"

    if [[ $mem_gb -lt 2 ]]; then
        ERRORS+=("内存不足 2GB，当前 ${mem_gb}GB")
    elif [[ $mem_gb -lt 4 ]]; then
        WARNINGS+=("内存建议至少 4GB，当前 ${mem_gb}GB")
    fi
}

# 3. 检查 CPU
check_cpu() {
    local cpu_cores=$(nproc)
    echo "CPU 核心数: $cpu_cores"

    if [[ $cpu_cores -lt 2 ]]; then
        ERRORS+=("CPU 核心数不足 2，当前 $cpu_cores")
    fi
}

# 4. 检查 swap
check_swap() {
    local swap_total=$(grep SwapTotal /proc/meminfo | awk '{print $2}')
    if [[ $swap_total -gt 0 ]]; then
        WARNINGS+=("检测到 swap 已启用，K8s 要求禁用 swap")
        echo "提示: 执行 'swapoff -a' 禁用 swap"
    fi
}

# 5. 检查端口
check_ports() {
    local required_ports=(6443 2379 2380 10250 10251 10252)
    for port in "${required_ports[@]}"; do
        if ss -tuln | grep -q ":${port} "; then
            ERRORS+=("端口 $port 已被占用")
        fi
    done
}

# 6. 检查必要的内核模块
check_kernel_modules() {
    local modules=(br_netfilter overlay)
    for mod in "${modules[@]}"; do
        if ! lsmod | grep -q "^${mod}"; then
            WARNINGS+=("内核模块 $mod 未加载")
        fi
    done
}

# 7. 检查 sysctl 配置
check_sysctl() {
    local bridge_nf_call=$(sysctl -n net.bridge.bridge-nf-call-iptables 2>/dev/null || echo "0")
    local ip_forward=$(sysctl -n net.ipv4.ip_forward)

    if [[ "$bridge_nf_call" != "1" ]]; then
        WARNINGS+=("net.bridge.bridge-nf-call-iptables 未设置为 1")
    fi
    if [[ "$ip_forward" != "1" ]]; then
        WARNINGS+=("net.ipv4.ip_forward 未设置为 1")
    fi
}

# 执行所有检查
echo "=== Kubernetes 预检查 ==="
check_os
check_memory
check_cpu
check_swap
check_ports
check_kernel_modules
check_sysctl

# 输出结果
echo ""
echo "=== 检查结果 ==="
if [[ ${#WARNINGS[@]} -gt 0 ]]; then
    echo "警告:"
    printf '  - %s\n' "${WARNINGS[@]}"
fi

if [[ ${#ERRORS[@]} -gt 0 ]]; then
    echo "错误:"
    printf '  - %s\n' "${ERRORS[@]}"
    exit 1
fi

echo "预检查通过"
exit 0
```

### 4.2 init.sh (kubeadm init)

```bash
#!/usr/bin/env bash
set -euo pipefail

# kubeadm init 脚本
# 参数通过环境变量传入:
#   POD_CIDR: Pod 网络 CIDR (默认 10.244.0.0/16)
#   SERVICE_CIDR: Service 网络 CIDR (默认 10.96.0.0/12)
#   K8S_VERSION: Kubernetes 版本 (默认 1.28.0)
#   CONTROL_PLANE_ENDPOINT: 控制平面端点 (可选)

POD_CIDR="${POD_CIDR:-10.244.0.0/16}"
SERVICE_CIDR="${SERVICE_CIDR:-10.96.0.0/12}"
K8S_VERSION="${K8S_VERSION:-1.28.0}"
CONTROL_PLANE_ENDPOINT="${CONTROL_PLANE_ENDPOINT:-}"
ADVERTISE_ADDRESS="${ADVERTISE_ADDRESS:-}"

# 构建 kubeadm init 命令
INIT_ARGS=(
    "--pod-network-cidr=${POD_CIDR}"
    "--service-cidr=${SERVICE_CIDR}"
    "--kubernetes-version=v${K8S_VERSION}"
    "--ignore-preflight-errors=Swap"
    "--upload-certs"
)

if [[ -n "$CONTROL_PLANE_ENDPOINT" ]]; then
    INIT_ARGS+=("--control-plane-endpoint=${CONTROL_PLANE_ENDPOINT}")
fi

if [[ -n "$ADVERTISE_ADDRESS" ]]; then
    INIT_ARGS+=("--apiserver-advertise-address=${ADVERTISE_ADDRESS}")
fi

echo "执行 kubeadm init..."
echo "参数: ${INIT_ARGS[*]}"

kubeadm init "${INIT_ARGS[@]}"

# 为当前用户配置 kubectl
mkdir -p "$HOME/.kube"
cp -f /etc/kubernetes/admin.conf "$HOME/.kube/config"
chown "$(id -u):$(id -g)" "$HOME/.kube/config"

# 输出 join 命令（供后续使用）
echo ""
echo "=== Join 命令 ==="
kubeadm token create --print-join-command > /tmp/kubeadm-join.sh
chmod +x /tmp/kubeadm-join.sh
cat /tmp/kubeadm-join.sh

echo ""
echo "kubeadm init 完成"
```

### 4.3 join.sh

```bash
#!/usr/bin/env bash
set -euo pipefail

# kubeadm join 脚本
# 参数通过环境变量传入:
#   JOIN_COMMAND: join 命令 (从 control plane 获取)
#   或
#   CONTROL_PLANE_IP: 控制平面 IP
#   TOKEN: join token
#   CA_CERT_HASH: CA 证书 hash

if [[ -n "${JOIN_COMMAND:-}" ]]; then
    echo "执行 join 命令..."
    eval "$JOIN_COMMAND"
elif [[ -n "${CONTROL_PLANE_IP:-}" && -n "${TOKEN:-}" && -n "${CA_CERT_HASH:-}" ]]; then
    echo "执行 kubeadm join..."
    kubeadm join "${CONTROL_PLANE_IP}:6443" \
        --token "${TOKEN}" \
        --discovery-token-ca-cert-hash "sha256:${CA_CERT_HASH}"
else
    echo "错误: 需要提供 JOIN_COMMAND 或 CONTROL_PLANE_IP/TOKEN/CA_CERT_HASH"
    exit 1
fi

echo "节点加入完成"
```

## 5. 前端组件设计

### 5.1 ClusterBootstrapWizard 改进

```tsx
// 步骤定义
const bootstrapSteps = [
  { key: 'basic', title: '基本信息' },
  { key: 'control-plane', title: 'Control Plane' },
  { key: 'workers', title: 'Worker 节点' },
  { key: 'network', title: '网络配置' },
  { key: 'review', title: '确认配置' },
  { key: 'execute', title: '执行安装' },
];

// 网络配置表单
const NetworkConfigForm = () => (
  <Form layout="vertical">
    <Form.Item name="k8sVersion" label="Kubernetes 版本" initialValue="1.28.0">
      <Select options={[
        { label: '1.28.0 (推荐)', value: '1.28.0' },
        { label: '1.27.0', value: '1.27.0' },
        { label: '1.26.0', value: '1.26.0' },
      ]} />
    </Form.Item>
    <Form.Item name="cni" label="CNI 插件" initialValue="calico">
      <Select options={[
        { label: 'Calico (推荐生产环境)', value: 'calico' },
        { label: 'Flannel (简单易用)', value: 'flannel' },
        { label: 'Cilium (高性能，支持 eBPF)', value: 'cilium' },
      ]} />
    </Form.Item>
    <Form.Item name="podCIDR" label="Pod CIDR" initialValue="10.244.0.0/16">
      <Input placeholder="10.244.0.0/16" />
    </Form.Item>
    <Form.Item name="serviceCIDR" label="Service CIDR" initialValue="10.96.0.0/12">
      <Input placeholder="10.96.0.0/12" />
    </Form.Item>
  </Form>
);

// 执行进度展示
const ExecutionProgress = ({ taskId }: { taskId: string }) => {
  const [task, setTask] = useState<BootstrapTask>();

  useEffect(() => {
    const interval = setInterval(async () => {
      const res = await Api.cluster.getBootstrapTask(taskId);
      setTask(res.data);
      if (res.data.status !== 'running') {
        clearInterval(interval);
      }
    }, 2000);
    return () => clearInterval(interval);
  }, [taskId]);

  return (
    <div className="space-y-4">
      <Steps current={task?.currentStep} direction="vertical">
        {task?.steps?.map(step => (
          <Step
            key={step.name}
            title={step.title}
            status={step.status}
            description={step.message}
          />
        ))}
      </Steps>
      {task?.status === 'running' && <Spin tip="安装进行中..." />}
      {task?.status === 'succeeded' && (
        <Result status="success" title="集群创建成功"
          extra={<Button onClick={() => navigate(`/clusters/${task.clusterId}`)}>查看集群</Button>}
        />
      )}
      {task?.status === 'failed' && (
        <Result status="error" title="集群创建失败"
          subTitle={task.errorMessage}
        />
      )}
    </div>
  );
};
```

## 6. 测试策略

### 6.1 单元测试

```go
func TestPreflightCheck(t *testing.T) {
    // 测试预检查脚本输出解析
}

func TestKubeconfigValidation(t *testing.T) {
    // 测试 kubeconfig 解析和验证
}

func TestClusterNodeSync(t *testing.T) {
    // 测试节点信息同步
}
```

### 6.2 集成测试

```go
func TestBootstrapWorkflow(t *testing.T) {
    // 使用 mock SSH 测试完整流程
}

func TestImportCluster(t *testing.T) {
    // 使用 kind 集群测试导入
}
```

### 6.3 E2E 测试

- 使用 kind 创建测试集群
- 测试自建集群流程
- 测试导入集群流程
- 测试资源查询 API
