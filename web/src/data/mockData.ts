import type { Host, Config, Task, TaskLog, K8sCluster, K8sNode, K8sPod, K8sService, K8sIngress, Alert, AlertRule, MonitorMetric, IntegrationTool, Service, ServiceQuota, ConfigApp, ConfigItem, ConfigTemplate, Release, AuditLog } from '../types';

// 任务调度数据部分（新增）
interface Job {
  id: string;
  name: string;
  type: 'shell' | 'http' | 'python' | 'ansible' | 'kubectl';
  command: string;
  schedule: string;
  timeout: number;
  hostGroupId?: string;
  strategy: 'random' | 'round-robin' | 'specify' | 'broadcast';
  retryCount: number;
  retryInterval: number;
  concurrencyPolicy: 'Allow' | 'Forbid' | 'Replace';
  description?: string;
  enabled: boolean;
  createdAt: string;
  updatedAt: string;
}

interface Execution {
  id: string;
  jobId: string;
  hostId?: string;
  startTime: string;
  endTime?: string;
  status: 'pending' | 'running' | 'success' | 'failed' | 'killed';
  exitCode?: number;
  stdout: string;
  stderr: string;
  retryCount: number;
}

interface JobSchedule {
  id: string;
  jobId: string;
  scheduledTime: string;
  actualStartTime?: string;
  actualEndTime?: string;
  status: 'scheduled' | 'executed' | 'missed';
}

interface HostV2 extends Host {
  status: 'online' | 'offline' | 'warning' | 'maintenance';
}

interface JobMockData {
  jobs: Job[];
  executions: Execution[];
  schedules: JobSchedule[];
  hosts: HostV2[];
}

export const mockJobData: JobMockData = {
  jobs: [
    {
      id: 'job-001',
      name: '数据库备份任务',
      type: 'shell',
      command: 'mysqldump -u root -p$DB_PASSWORD myapp_db > backup_$(date +%Y%m%d).sql && gzip backup_$(date +%Y%m%d).sql',
      schedule: '0 2 * * *', // 每天凌晨2点执行
      timeout: 3600,
      strategy: 'random',
      retryCount: 2,
      retryInterval: 60,
      concurrencyPolicy: 'Forbid',
      description: '每日定时备份生产数据库',
      enabled: true,
      createdAt: '2026-01-15T08:30:00Z',
      updatedAt: '2026-01-20T14:20:00Z',
    },
    {
      id: 'job-002',
      name: '系统健康检查',
      type: 'python',
      command: 'python3 /opt/scripts/system_health_check.py --env=prod',
      schedule: '*/5 * * * *', // 每5分钟执行一次
      timeout: 300,
      strategy: 'broadcast',
      retryCount: 1,
      retryInterval: 30,
      concurrencyPolicy: 'Allow',
      description: '监测服务器CPU、内存、磁盘使用情况',
      enabled: true,
      createdAt: '2026-02-01T10:15:00Z',
      updatedAt: '2026-02-15T09:45:00Z',
    },
    {
      id: 'job-003',
      name: 'API服务部署',
      type: 'ansible',
      command: 'ansible-playbook -i production deploy.yml',
      schedule: '', // 手动触发
      timeout: 1800,
      strategy: 'specify',
      retryCount: 3,
      retryInterval: 120,
      concurrencyPolicy: 'Replace',
      description: '自动化部署API服务到生产环境',
      enabled: false, // 禁用
      createdAt: '2026-01-28T16:20:00Z',
      updatedAt: '2026-02-20T11:30:00Z',
    },
    {
      id: 'job-004',
      name: '清理临时文件',
      type: 'shell',
      command: 'find /tmp -type f -mtime +7 -delete && find /var/log -name "*.log" -mtime +30 -delete',
      schedule: '0 3 1 * *', // 每月1号凌晨3点执行
      timeout: 600,
      strategy: 'round-robin',
      retryCount: 0,
      retryInterval: 0,
      concurrencyPolicy: 'Forbid',
      description: '定期清理过期的临时文件和日志',
      enabled: true,
      createdAt: '2026-01-10T09:10:00Z',
      updatedAt: '2026-02-10T09:10:00Z',
    },
    {
      id: 'job-005',
      name: 'K8s集群检查',
      type: 'kubectl',
      command: 'kubectl get nodes && kubectl top nodes && kubectl get pods --all-namespaces',
      schedule: '*/10 * * * *', // 每10分钟执行一次
      timeout: 300,
      strategy: 'round-robin',
      retryCount: 1,
      retryInterval: 60,
      concurrencyPolicy: 'Allow',
      description: '监测Kubernetes集群节点和Pod状态',
      enabled: true,
      createdAt: '2026-01-18T14:45:00Z',
      updatedAt: '2026-02-18T14:45:00Z',
    },
    {
      id: 'job-006',
      name: '安全补丁安装',
      type: 'shell',
      command: 'apt-get update && apt-get upgrade -y',
      schedule: '0 4 15 * *', // 每月15号凌晨4点执行
      timeout: 3600,
      strategy: 'broadcast',
      retryCount: 2,
      retryInterval: 180,
      concurrencyPolicy: 'Forbid',
      description: '安装最新的安全补丁',
      enabled: true,
      createdAt: '2026-02-05T13:00:00Z',
      updatedAt: '2026-02-05T13:00:00Z',
    },
  ],
  executions: [
    {
      id: 'exec-001',
      jobId: 'job-001',
      startTime: '2026-02-22T02:00:05Z',
      endTime: '2026-02-22T02:15:22Z',
      status: 'success',
      stdout: 'mysqldump: [Warning] Using a password on the command line interface can be insecure.\nDump completed on 2026-02-22 02:15:22\nCompressed file created successfully',
      stderr: '',
      retryCount: 0,
      exitCode: 0,
    },
    {
      id: 'exec-002',
      jobId: 'job-001',
      startTime: '2026-02-21T02:00:05Z',
      endTime: '2026-02-21T02:16:18Z',
      status: 'success',
      stdout: 'mysqldump: [Warning] Using a password on the command line interface can be insecure.\nDump completed on 2026-02-21 02:16:18\nCompressed file created successfully',
      stderr: '',
      retryCount: 0,
      exitCode: 0,
    },
    {
      id: 'exec-003',
      jobId: 'job-002',
      startTime: '2026-02-23T10:15:00Z',
      status: 'running',
      stdout: 'CPU Usage: 24%\nMemory Usage: 67%\nDisk Usage: 45%\nAll systems healthy',
      stderr: '',
      retryCount: 0,
      exitCode: 0,
    },
    {
      id: 'exec-004',
      jobId: 'job-002',
      startTime: '2026-02-23T10:10:00Z',
      endTime: '2026-02-23T10:10:12Z',
      status: 'success',
      stdout: 'CPU Usage: 19%\nMemory Usage: 62%\nDisk Usage: 43%\nAll systems healthy',
      stderr: '',
      retryCount: 0,
      exitCode: 0,
    },
    {
      id: 'exec-005',
      jobId: 'job-003',
      startTime: '2026-02-20T14:30:00Z',
      endTime: '2026-02-20T14:42:15Z',
      status: 'failed',
      stdout: 'Checking connection to servers...\nDeploying new application version...',
      stderr: 'Error: Could not connect to deployment server\nRetrying... Connect timeout\nDeployment failed after 3 attempts',
      retryCount: 3,
      exitCode: 1,
    },
    {
      id: 'exec-006',
      jobId: 'job-006',
      startTime: '2026-01-15T04:00:05Z',
      endTime: '2026-01-15T05:20:45Z',
      status: 'success',
      stdout: 'Hit:1 http://archive.ubuntu.com/ubuntu bionic InRelease\nGet:2 http://archive.ubuntu.com/ubuntu bionic-updates InRelease [88.7 kB]\nFetched 123 MB in 3s (35.2 MB/s)\nReading changelogs...\nProcessing triggers for libc-bin (2.27-3ubuntu1.2) ...\nPackages upgraded: 15, New: 0, Removed: 0',
      stderr: '',
      retryCount: 0,
      exitCode: 0,
    },
  ],
  schedules: [
    {
      id: 'sched-001',
      jobId: 'job-001',
      scheduledTime: '2026-02-23T02:00:00Z',
      actualStartTime: '2026-02-23T02:00:05Z',
      actualEndTime: '2026-02-23T02:14:55Z',
      status: 'executed',
    },
    {
      id: 'sched-002',
      jobId: 'job-001',
      scheduledTime: '2026-02-22T02:00:00Z',
      actualStartTime: '2026-02-22T02:00:05Z',
      actualEndTime: '2026-02-22T02:15:22Z',
      status: 'executed',
    },
    {
      id: 'sched-003',
      jobId: 'job-001',
      scheduledTime: '2026-02-21T02:00:00Z',
      actualStartTime: '2026-02-21T02:00:05Z',
      actualEndTime: '2026-02-21T02:16:18Z',
      status: 'executed',
    },
    {
      id: 'sched-004',
      jobId: 'job-002',
      scheduledTime: '2026-02-23T10:15:00Z',
      actualStartTime: '2026-02-23T10:15:00Z',
      status: 'scheduled', // 计划执行中（而不是 running）
    },
    {
      id: 'sched-005',
      jobId: 'job-002',
      scheduledTime: '2026-02-23T10:10:00Z',
      actualStartTime: '2026-02-23T10:10:00Z',
      actualEndTime: '2026-02-23T10:10:12Z',
      status: 'executed',
    },
    {
      id: 'sched-006',
      jobId: 'job-004',
      scheduledTime: '2026-02-01T03:00:00Z',
      actualStartTime: '2026-02-01T03:00:05Z',
      actualEndTime: '2026-02-01T03:02:45Z',
      status: 'executed',
    },
    {
      id: 'sched-007',
      jobId: 'job-005',
      scheduledTime: '2026-02-23T10:10:00Z',
      actualStartTime: '2026-02-23T10:10:03Z',
      actualEndTime: '2026-02-23T10:10:08Z',
      status: 'executed',
    },
    {
      id: 'sched-008',
      jobId: 'job-005',
      scheduledTime: '2026-02-23T10:00:00Z',
      actualStartTime: '2026-02-23T10:00:02Z',
      actualEndTime: '2026-02-23T10:00:07Z',
      status: 'executed',
    },
  ],
  hosts: [
    {
      id: 'host-001',
      name: 'api-server-01',
      ip: '192.168.1.10',
      status: 'online',
      cpu: 42,
      memory: 65,
      disk: 30,
      network: 125,
      tags: ['api', 'prod'],
      region: 'us-west-1',
      createdAt: '2025-12-01T08:00:00Z',
      lastActive: '2026-02-23T10:15:30Z',
    },
    {
      id: 'host-002',
      name: 'db-master-01',
      ip: '192.168.1.15',
      status: 'online',
      cpu: 75,
      memory: 82,
      disk: 55,
      network: 450,
      tags: ['db', 'master', 'prod'],
      region: 'us-west-1',
      createdAt: '2025-11-15T10:30:00Z',
      lastActive: '2026-02-23T10:16:45Z',
    },
    {
      id: 'host-003',
      name: 'cache-node-01',
      ip: '192.168.1.20',
      status: 'warning',
      cpu: 89,
      memory: 91,
      disk: 20,
      network: 80,
      tags: ['cache', 'redis', 'prod'],
      region: 'us-west-1',
      createdAt: '2025-12-10T14:20:00Z',
      lastActive: '2026-02-23T10:14:15Z',
    },
    {
      id: 'host-004',
      name: 'worker-node-01',
      ip: '192.168.2.10',
      status: 'online',
      cpu: 25,
      memory: 35,
      disk: 40,
      network: 75,
      tags: ['worker', 'batch'],
      region: 'us-east-1',
      createdAt: '2025-12-20T09:15:00Z',
      lastActive: '2026-02-23T10:12:30Z',
    },
    {
      id: 'host-005',
      name: 'monitoring-01',
      ip: '192.168.1.99',
      status: 'online',
      cpu: 12,
      memory: 23,
      disk: 15,
      network: 25,
      tags: ['monitor', 'alert'],
      region: 'us-west-1',
      createdAt: '2025-11-01T11:00:00Z',
      lastActive: '2026-02-23T10:17:20Z',
    },
  ],
};

// 之前的数据...

// Host Data
export const hosts: Host[] = [
  {
    id: 'host-1',
    name: 'Web Server 1',
    ip: '192.168.1.10',
    status: 'online',
    cpu: 72,
    memory: 58,
    disk: 45,
    network: 82,
    tags: ['web', 'production', 'us-west'],
    region: 'us-west-1',
    createdAt: '2023-01-15T08:30:00Z',
    lastActive: '2023-11-20T14:20:00Z',
  },
  {
    id: 'host-2',
    name: 'Database Server 1',
    ip: '192.168.1.15',
    status: 'online',
    cpu: 65,
    memory: 73,
    disk: 60,
    network: 42,
    tags: ['database', 'production', 'us-west'],
    region: 'us-west-1',
    createdAt: '2023-02-10T09:15:00Z',
    lastActive: '2023-11-20T14:15:00Z',
  },
  {
    id: 'host-3',
    name: 'Cache Server 1',
    ip: '192.168.1.20',
    status: 'warning',
    cpu: 85,
    memory: 82,
    disk: 38,
    network: 65,
    tags: ['cache', 'redis', 'staging'],
    region: 'us-east-1',
    createdAt: '2023-02-20T11:20:00Z',
    lastActive: '2023-11-20T14:10:00Z',
  },
  {
    id: 'host-4',
    name: 'API Server 1',
    ip: '192.168.1.25',
    status: 'offline',
    cpu: 0,
    memory: 0,
    disk: 30,
    network: 0,
    tags: ['api', 'testing', 'us-west'],
    region: 'us-west-2',
    createdAt: '2023-03-01T12:00:00Z',
    lastActive: '2023-10-15T08:30:00Z',
  },
];

// Task Data
export const tasks: Task[] = [
  {
    id: 'task-1',
    name: 'Data Backup',
    type: 'scheduled',
    status: 'success',
    schedule: '0 2 * * *',
    lastRun: '2023-11-19T02:00:00Z',
    nextRun: '2023-11-20T02:00:00Z',
    duration: 125,
    createdAt: '2023-01-15T08:30:00Z',
  },
  {
    id: 'task-2',
    name: 'Security Scan',
    type: 'dependency',
    status: 'running',
    lastRun: '2023-11-20T10:00:00Z',
    nextRun: '2023-11-21T10:00:00Z',
    createdAt: '2023-02-20T11:20:00Z',
  },
  {
    id: 'task-3',
    name: 'Log Cleanup',
    type: 'scheduled',
    status: 'failed',
    schedule: '0 1 * * 0',
    lastRun: '2023-11-12T01:00:00Z',
    nextRun: '2023-11-26T01:00:00Z',
    duration: 30,
    createdAt: '2023-02-10T09:15:00Z',
  },
  {
    id: 'task-4',
    name: 'Performance Test',
    type: 'parallel',
    status: 'pending',
    nextRun: '2023-11-20T16:00:00Z',
    createdAt: '2023-03-01T12:00:00Z',
  },
];

// Task Log Data
export const taskLogs: TaskLog[] = [
  {
    id: 'log-1',
    taskId: 'task-1',
    timestamp: '2023-11-19T02:00:00Z',
    level: 'info',
    message: 'Backup started',
  },
  {
    id: 'log-2',
    taskId: 'task-1',
    timestamp: '2023-11-19T02:02:05Z',
    level: 'info',
    message: 'Backup completed successfully',
  },
  {
    id: 'log-3',
    taskId: 'task-2',
    timestamp: '2023-11-20T10:00:00Z',
    level: 'info',
    message: 'Scan initiated',
  },
  {
    id: 'log-4',
    taskId: 'task-2',
    timestamp: '2023-11-20T10:15:30Z',
    level: 'warn',
    message: 'Medium severity vulnerability detected',
  },
];

// Kubernetes Cluster Data
export const k8sClusters: K8sCluster[] = [
  {
    id: 'cluster-1',
    name: 'Production Cluster',
    version: 'v1.25.3',
    status: 'healthy',
    nodes: 10,
    pods: 98,
    cpu: 75,
    memory: 65,
    createdAt: '2023-01-15T08:30:00Z',
  },
  {
    id: 'cluster-2',
    name: 'Staging Cluster',
    version: 'v1.24.8',
    status: 'warning',
    nodes: 3,
    pods: 24,
    cpu: 45,
    memory: 52,
    createdAt: '2023-02-20T11:20:00Z',
  },
];

// Kubernetes Node Data
export const k8sNodes: K8sNode[] = [
  {
    id: 'node-1',
    name: 'prod-worker-1',
    role: 'worker',
    status: 'ready',
    cpu: 72,
    memory: 64,
    pods: 15,
    labels: {
      zone: 'us-west-1a',
      'node-type': 'general-purpose',
    },
  },
  {
    id: 'node-2',
    name: 'prod-worker-2',
    role: 'worker',
    status: 'ready',
    cpu: 80,
    memory: 72,
    pods: 18,
    labels: {
      zone: 'us-west-1b',
      'node-type': 'general-purpose',
    },
  },
  {
    id: 'node-3',
    name: 'staging-worker-1',
    role: 'worker',
    status: 'ready',
    cpu: 45,
    memory: 32,
    pods: 8,
    labels: {
      zone: 'us-east-1a',
      'node-type': 'general-purpose',
    },
  },
];

// Kubernetes Pod Data
export const k8sPods: K8sPod[] = [
  {
    id: 'pod-1',
    name: 'web-app-1',
    namespace: 'frontend',
    status: 'running',
    phase: 'Running',
    node: 'prod-worker-1',
    cpu: 25,
    memory: 15,
    restarts: 0,
    age: '2d 4h',
    labels: { app: 'web-app', tier: 'frontend' },
    containers: [
      { name: 'main', image: 'nginx:1.21', status: 'running', cpu: 20, memory: 12, restarts: 0 }
    ],
    qosClass: 'Burstable',
    createdBy: 'devops-team',
    startTime: '2024-02-20T08:30:00Z',
  },
  {
    id: 'pod-2',
    name: 'db-pod-1',
    namespace: 'backend',
    status: 'running',
    phase: 'Running',
    node: 'prod-worker-2',
    cpu: 45,
    memory: 60,
    restarts: 1,
    age: '5d 2h',
    labels: { app: 'database', tier: 'backend' },
    containers: [
      { name: 'mysql', image: 'mysql:8.0', status: 'running', cpu: 40, memory: 55, restarts: 1 }
    ],
    qosClass: 'Guaranteed',
    createdBy: 'dba-team',
    startTime: '2024-02-20T08:45:00Z',
  },
  {
    id: 'pod-3',
    name: 'cache-pod-1',
    namespace: 'infrastructure',
    status: 'running',
    phase: 'Running',
    node: 'prod-worker-3',
    cpu: 35,
    memory: 40,
    restarts: 0,
    age: '3d 7h',
    labels: { app: 'redis', component: 'cache' },
    containers: [
      { name: 'redis-main', image: 'redis:7.0', status: 'running', cpu: 30, memory: 35, restarts: 0 }
    ],
    qosClass: 'Burstable',
    createdBy: 'devops-team',
    startTime: '2024-02-19T14:20:00Z',
  },
  {
    id: 'pod-4',
    name: 'monitor-pod-1',
    namespace: 'monitoring',
    status: 'running',
    phase: 'Running',
    node: 'monitor-worker-1',
    cpu: 28,
    memory: 32,
    restarts: 2,
    age: '10d 1h',
    labels: { app: 'prometheus', tier: 'monitoring' },
    containers: [
      { name: 'prometheus', image: 'prom/prometheus:v2.37', status: 'running', cpu: 25, memory: 30, restarts: 2 }
    ],
    qosClass: 'Burstable',
    createdBy: 'platform-team',
    startTime: '2024-02-15T09:15:00Z',
  },
  {
    id: 'pod-5',
    name: 'ingress-pod-1',
    namespace: 'kube-system',
    status: 'pending',
    phase: 'Pending',
    node: 'control-plane',
    cpu: 0,
    memory: 0,
    restarts: 0,
    age: '1h 5m',
    labels: { app: 'nginx-ingress', tier: 'ingress' },
    containers: [],
    qosClass: 'BestEffort',
    createdBy: 'admin',
    startTime: '2024-02-25T10:25:00Z',
  },
  {
    id: 'pod-6',
    name: 'api-gateway-1',
    namespace: 'edge',
    status: 'running',
    phase: 'Running',
    node: 'ingress-worker-1',
    cpu: 20,
    memory: 25,
    restarts: 1,
    age: '7d 12h',
    labels: { app: 'api-gateway', tier: 'edge' },
    containers: [
      { name: 'nginx', image: 'nginx:1.23', status: 'running', cpu: 15, memory: 20, restarts: 1 }
    ],
    qosClass: 'Burstable',
    createdBy: 'microservices-team',
    startTime: '2024-02-18T16:45:00Z',
  },
  {
    id: 'pod-7',
    name: 'log-agent-1',
    namespace: 'logging',
    status: 'running',
    phase: 'Running',
    node: 'prod-worker-1',
    cpu: 12,
    memory: 8,
    restarts: 0,
    age: '15d 8h',
    labels: { app: 'fluentd', component: 'logging-agent' },
    containers: [
      { name: 'fluentd', image: 'fluent/fluentd:v1.14', status: 'running', cpu: 10, memory: 6, restarts: 0 }
    ],
    qosClass: 'BestEffort',
    createdBy: 'devops-team',
    startTime: '2024-02-10T07:30:00Z',
  },
];

// Kubernetes Service Data
export const k8sServices: K8sService[] = [
  {
    id: 'svc-1',
    name: 'web-svc',
    namespace: 'frontend',
    type: 'LoadBalancer',
    clusterIP: '10.96.5.10',
    externalIP: '203.0.113.25',
    ports: [
      {
        port: 80,
        targetPort: 8080,
        protocol: 'TCP',
      },
    ],
    selector: {
      app: 'web-app',
    },
    age: '2d 10h',
  },
  {
    id: 'svc-2',
    name: 'db-svc',
    namespace: 'backend',
    type: 'ClusterIP',
    clusterIP: '10.96.5.20',
    ports: [
      {
        port: 5432,
        targetPort: 5432,
        protocol: 'TCP',
      },
    ],
    selector: {
      app: 'database',
    },
    age: '5d 2h',
  },
];

// Kubernetes Ingress Data
export const k8sIngresses: K8sIngress[] = [
  {
    id: 'ing-1',
    name: 'web-ingress',
    namespace: 'frontend',
    host: 'example.com',
    path: '/',
    service: 'web-svc',
    port: 80,
    tls: true,
  },
  {
    id: 'ing-2',
    name: 'api-ingress',
    namespace: 'backend',
    host: 'api.example.com',
    path: '/',
    service: 'api-svc',
    port: 80,
    tls: true,
  },
];

// Alert Data
export const alerts: Alert[] = [
  {
    id: 'alert-1',
    title: 'High CPU Usage on Web Server',
    severity: 'warning',
    source: 'Prometheus',
    status: 'firing',
    createdAt: '2023-11-20T09:45:00Z',
  },
  {
    id: 'alert-2',
    title: 'Server Unreachable',
    severity: 'critical',
    source: 'Ping Prober',
    status: 'firing',
    createdAt: '2023-11-20T10:15:00Z',
    resolvedAt: '2023-11-20T11:30:00Z',
  },
  {
    id: 'alert-3',
    title: 'Disk Space Low',
    severity: 'warning',
    source: 'Agent Monitoring',
    status: 'resolved',
    createdAt: '2023-11-19T14:20:00Z',
    resolvedAt: '2023-11-19T15:45:00Z',
  },
];

// Alert Rule Data
export const alertRules: AlertRule[] = [
  {
    id: 'rule-1',
    name: 'CPU Threshold',
    condition: 'cpu_usage > 80%',
    severity: 'warning',
    enabled: true,
    channels: ['email', 'slack'],
    createdAt: '2023-01-15T08:30:00Z',
  },
  {
    id: 'rule-2',
    name: 'Memory Threshold',
    condition: 'memory_usage > 85%',
    severity: 'critical',
    enabled: true,
    channels: ['email', 'sms'],
    createdAt: '2023-02-10T09:15:00Z',
  },
  {
    id: 'rule-3',
    name: 'Disk Threshold',
    condition: 'disk_usage > 90%',
    severity: 'critical',
    enabled: true,
    channels: ['email', 'webhook'],
    createdAt: '2023-02-20T11:20:00Z',
  },
];

// Monitor Metric Data
export const monitorMetrics: MonitorMetric[] = [
  {
    timestamp: '2023-11-20T00:00:00Z',
    value: 72.1,
  },
  {
    timestamp: '2023-11-20T01:00:00Z',
    value: 74.3,
  },
  {
    timestamp: '2023-11-20T02:00:00Z',
    value: 68.9,
  },
  {
    timestamp: '2023-11-20T03:00:00Z',
    value: 71.2,
  },
  {
    timestamp: '2023-11-20T04:00:00Z',
    value: 79.5,
  },
  {
    timestamp: '2023-11-20T05:00:00Z',
    value: 82.4,
  },
  {
    timestamp: '2023-11-20T06:00:00Z',
    value: 78.1,
  },
  {
    timestamp: '2023-11-20T07:00:00Z',
    value: 69.8,
  },
  {
    timestamp: '2023-11-20T08:00:00Z',
    value: 81.2,
  },
  {
    timestamp: '2023-11-20T09:00:00Z',
    value: 84.5,
  },
  {
    timestamp: '2023-11-20T10:00:00Z',
    value: 90.1,
  },
];

// Integration Tool Data
export const integrationTools: IntegrationTool[] = [
  {
    id: 'tool-1',
    name: 'Grafana',
    icon: '📊',
    url: 'https://grafana.example.com',
    status: 'connected',
    description: 'Visualization and analytics platform',
  },
  {
    id: 'tool-2',
    name: 'Jenkins',
    icon: '⚙️',
    url: 'https://jenkins.example.com',
    status: 'connected',
    description: 'CI/CD automation server',
  },
  {
    id: 'tool-3',
    name: 'Slack',
    icon: '💬',
    url: 'https://slack.example.com',
    status: 'connected',
    description: 'Communication platform',
  }
];

// ============================================================================
// Configuration Center Mock Data (Backward Compatibility)
// ============================================================================

// Old Config interface data (maintains compatibility with legacy ConfigPage)
export const configs: Config[] = [
  { 
    id: '1', 
    name: '数据库连接池配置',
    key: 'db.pool.size', 
    value: '{"min": 5, "max": 50}', 
    version: 3, 
    status: 'active', 
    type: 'json', 
    env: 'production', 
    updatedAt: '2024-02-15 14:30:00', 
    updatedBy: 'admin' 
  },
  { 
    id: '2', 
    name: 'Redis缓存策略',
    key: 'redis.cache.policy', 
    value: '{"ttl": 3600, "maxmemory": "2gb"}', 
    version: 2, 
    status: 'active', 
    type: 'json', 
    env: 'production', 
    updatedAt: '2024-02-14 10:20:00', 
    updatedBy: 'devops' 
  },
  { 
    id: '3', 
    name: 'API限流配置',
    key: 'api.rate.limit', 
    value: '{"requests": 1000, "window": "60s"}', 
    version: 1, 
    status: 'active', 
    type: 'json', 
    env: 'production', 
    updatedAt: '2024-02-10 09:00:00', 
    updatedBy: 'admin' 
  },
  { 
    id: '4', 
    name: '日志级别配置',
    key: 'app.log.level', 
    value: 'info', 
    version: 4, 
    status: 'active', 
    type: 'string', 
    env: 'production', 
    updatedAt: '2024-02-18 16:45:00', 
    updatedBy: 'devops' 
  },
  { 
    id: '5', 
    name: '邮件通知配置',
    key: 'notification.email.smtp', 
    value: '{"host": "smtp.company.com", "port": 587}', 
    version: 2, 
    status: 'draft', 
    type: 'json', 
    env: 'staging', 
    updatedAt: '2024-02-19 11:30:00', 
    updatedBy: 'admin' 
  },
  { 
    id: '6', 
    name: '文件上传限制',
    key: 'upload.max.size', 
    value: '10485760', 
    version: 1, 
    status: 'active', 
    type: 'number', 
    env: 'production', 
    updatedAt: '2024-02-12 08:15:00', 
    updatedBy: 'devops' 
  },
  { 
    id: '7', 
    name: 'JWT密钥配置',
    key: 'auth.jwt.secret', 
    value: '******', 
    version: 5, 
    status: 'active', 
    type: 'secret', 
    env: 'production', 
    updatedAt: '2024-02-17 13:00:00', 
    updatedBy: 'admin' 
  },
  { 
    id: '8', 
    name: 'CDN加速域名',
    key: 'cdn.domains', 
    value: '["cdn1.example.com", "cdn2.example.com"]', 
    version: 2, 
    status: 'deprecated', 
    type: 'json', 
    env: 'production', 
    updatedAt: '2024-02-08 15:20:00', 
    updatedBy: 'devops' 
  }
];

// Helper function to get config versions (for backwards compatibility)
export const getConfigVersions = (configId: string) => {
  // Return a simple list based on current config version (for backwards compatibility)
  const config = configs.find(c => c.id === configId);
  if (!config) return [];
  
  // Generate dummy versions for backwards compatibility
  const versions = [];
  for (let i = 1; i <= config.version; i++) {
    versions.push({
      id: `${config.id}-v${i}`,
      configId: config.id,
      version: i,
      value: config.value, // Using same value for demo
      createdAt: config.updatedAt,
      createdBy: config.updatedBy,
      comment: `Version ${i}`
    });
  }
  return versions;
};

// Service Management Data
export const services: Service[] = [
  {
    id: 'svc-app-1',
    name: 'Frontend App',
    status: 'running',
    owner: 'frontteam',
    environment: 'production',
    tags: ['web', 'ui', 'http'],
    cpu: 60,
    memory: 2048,
    replicas: 3,
    lastDeployTime: '2023-11-18T14:30:00Z',
    createdAt: '2023-01-15T08:30:00Z',
    k8sResources: {
      pods: [
        {
          id: 'fe-pod-1',
          name: 'fe-web-1',
          namespace: 'frontend',
          status: 'running',
          cpu: 15,
          memory: 512,
          restarts: 0,
          age: '4d',
          node: 'prod-worker-1',
        },
        {
          id: 'fe-pod-2',
          name: 'fe-web-2',
          namespace: 'frontend',
          status: 'running',
          cpu: 18,
          memory: 512,
          restarts: 0,
          age: '4d',
          node: 'prod-worker-2',
        },
      ],
      services: [],
      ingresses: [],
    },
    config: '{\n  "port": 80,\n  "healthCheck": "/health"\n}',
    metrics: {
      cpu: [60, 65, 70, 72, 65],
      memory: [2048, 2100, 2200, 2090, 2110],
    },
  },
  {
    id: 'svc-api-1',
    name: 'API Gateway',
    status: 'syncing',
    owner: 'apiteam',
    environment: 'production',
    tags: ['api', 'gateway', 'http'],
    cpu: 85,
    memory: 4096,
    replicas: 2,
    lastDeployTime: '2023-11-19T09:15:00Z',
    createdAt: '2023-02-10T09:15:00Z',
    k8sResources: {
      pods: [
        {
          id: 'api-pod-1',
          name: 'api-gw-1',
          namespace: 'backend',
          status: 'running',
          cpu: 35,
          memory: 1024,
          restarts: 0,
          age: '3d',
          node: 'prod-worker-1',
        },
      ],
      services: [],
      ingresses: [],
    },
    config: '{\n  "port": 8080,\n  "rateLimit": 1000\n}',
    metrics: {
      cpu: [85, 88, 92, 87, 85],
      memory: [4096, 4196, 4250, 4180, 4096],
    },
  },
  {
    id: 'svc-db-1',
    name: 'Database Service',
    status: 'deploying',
    owner: 'dbsquad',
    environment: 'staging',
    tags: ['database', 'postgres', 'sql'],
    cpu: 45,
    memory: 8192,
    replicas: 1,
    lastDeployTime: '2023-11-20T10:45:00Z',
    createdAt: '2023-02-20T11:20:00Z',
    k8sResources: {
      pods: [
        {
          id: 'db-pod-1',
          name: 'db-main-1',
          namespace: 'database',
          status: 'running',
          cpu: 45,
          memory: 2048,
          restarts: 0,
          age: '2d 6h',
          node: 'prod-worker-2',
        },
      ],
      services: [],
      ingresses: [],
    },
    config: '{\n  "engine": "PostgreSQl",\n  "version": "14"\n}',
    metrics: {
      cpu: [45, 50, 40, 52, 48],
      memory: [8192, 8200, 8250, 8180, 8300],
    },
  },
];

export const serviceQuotas: ServiceQuota = {
  cpuLimit: 100,
  memoryLimit: 16384,
  cpuUsed: 190,
  memoryUsed: 14336,
};

// Config Center Mock Data

// Config Apps
export const configApps: ConfigApp[] = [
  {
    id: 'app-1',
    name: 'user-service',
    serviceId: 'svc-app-1',
    description: '用户服务配置',
    namespaces: ['development', 'staging', 'production'],
    createdAt: '2026-01-10T09:15:00Z',
    updatedAt: '2026-01-20T09:15:00Z',
  },
  {
    id: 'app-2',
    name: 'order-service',
    serviceId: 'svc-app-2',
    description: '订单服务配置',
    namespaces: ['development', 'staging', 'production'],
    createdAt: '2026-01-15T11:00:00Z',
    updatedAt: '2026-01-25T14:30:00Z',
  },
  {
    id: 'app-3',
    name: 'payment-service',
    serviceId: 'svc-app-3',
    description: '支付服务配置',
    namespaces: ['development', 'staging', 'production'],
    createdAt: '2026-01-20T15:20:00Z',
    updatedAt: '2026-01-30T10:15:00Z',
  },
];

// Config Items
export const configItems: ConfigItem[] = [
  {
    id: 'config-1',
    appId: 'app-1',
    namespace: 'production',
    env: 'prod',
    key: 'database.url',
    value: 'jdbc:mysql://prod-db.example.com:3306/users',
    format: 'text',
    isSecret: false,
    createdAt: '2026-01-10T10:00:00Z',
    updatedAt: '2026-01-10T10:00:00Z',
    updatedBy: 'admin',
    versions: [
      {
        version: 1,
        value: 'jdbc:mysql://test-db.example.com:3306/users',
        createdBy: 'admin',
        createdAt: '2026-01-10T10:00:00Z',
        comment: '测试环境URL',
      },
      {
        version: 2,
        value: 'jdbc:mysql://prod-db.example.com:3306/users',
        createdBy: 'admin',
        createdAt: '2026-01-15T14:00:00Z',
        comment: '生产环境URL',
      },
    ],
  },
  {
    id: 'config-2',
    appId: 'app-2',
    namespace: 'production',
    env: 'prod',
    key: 'redis.endpoint',
    value: 'redis-prod-cluster.abcd.0001.usw2.cache.amazonaws.com',
    format: 'text',
    isSecret: false,
    createdAt: '2026-01-15T12:00:00Z',
    updatedAt: '2026-01-15T12:00:00Z',
    updatedBy: 'admin',
    versions: [
      {
        version: 1,
        value: 'redis-test-cluster.efgh.0001.usw2.cache.amazonaws.com',
        createdBy: 'admin',
        createdAt: '2026-01-15T12:00:00Z',
        comment: '初始配置',
      },
    ],
  },
  {
    id: 'config-3',
    appId: 'app-1',
    namespace: 'production',
    env: 'prod',
    key: 'app.features',
    value: '{"feature-x": true, "feature-y": false}',
    format: 'json',
    isSecret: false,
    createdAt: '2026-01-25T16:30:00Z',
    updatedAt: '2026-01-25T16:30:00Z',
    updatedBy: 'developer',
    versions: [
      {
        version: 1,
        value: '{"feature-x": false, "feature-y": false}',
        createdBy: 'qa-engineer',
        createdAt: '2026-01-25T16:30:00Z',
        comment: '关闭功能X/Y',
      },
      {
        version: 2,
        value: '{"feature-x": true, "feature-y": false}',
        createdBy: 'admin',
        createdAt: '2026-02-01T09:00:00Z',
        comment: '开启功能X',
      },
    ],
  },
];

// Audit Logs  
export const auditLogs: AuditLog[] = [
  { id: 'log-1', appId: 'app-1', appName: 'user-service', namespace: 'database', key: 'db.pool.size', action: 'update', operator: 'admin', timestamp: '2024-02-15 14:30:00', details: '更新配置值: max: 30 -> max: 50', oldValue: JSON.stringify({ min: 5, max: 30 }), newValue: JSON.stringify({ min: 5, max: 50 }), status: 'success' },
  { id: 'log-2', appId: 'app-1', appName: 'user-service', namespace: 'database', key: 'db.pool.size', action: 'release', operator: 'admin', timestamp: '2024-02-15 14:30:00', details: '发布配置 v3 到生产环境', status: 'success' },
  { id: 'log-3', appId: 'app-1', appName: 'user-service', namespace: 'redis', key: 'redis.cache.ttl', action: 'update', operator: 'devops', timestamp: '2024-02-14 10:20:00', details: '更新配置值: 1800 -> 3600', oldValue: '1800', newValue: '3600', status: 'success' },
  { id: 'log-4', appId: 'app-1', appName: 'user-service', namespace: 'redis', key: 'redis.cache.ttl', action: 'release', operator: 'devops', timestamp: '2024-02-14 10:20:00', details: '发布配置 v2 到生产环境', status: 'success' },
  { id: 'log-5', appId: 'app-1', appName: 'user-service', namespace: 'api', key: 'api.rate.limit', action: 'create', operator: 'admin', timestamp: '2024-02-10 09:00:00', details: '创建新配置', newValue: JSON.stringify({ requests: 1000, window: '60s' }), status: 'success' },
  { id: 'log-6', appId: 'app-1', appName: 'user-service', namespace: 'api', key: 'api.rate.limit', action: 'release', operator: 'admin', timestamp: '2024-02-10 09:00:00', details: '发布配置 v1 到生产环境', status: 'success' },
  { id: 'log-7', appId: 'app-5', appName: 'gateway-service', namespace: 'rate-limit', key: 'gateway.rate.limit.rules', action: 'update', operator: 'admin', timestamp: '2024-02-19 10:00:00', details: '更新限流规则', oldValue: 'global:\n  requests: 5000\n  window: 60s', newValue: 'global:\n  requests: 10000\n  window: 60s\nendpoints:\n  /api/users:\n    requests: 1000\n  /api/orders:\n    requests: 500', status: 'success' },
  { id: 'log-8', appId: 'app-5', appName: 'gateway-service', namespace: 'rate-limit', key: 'gateway.rate.limit.rules', action: 'release', operator: 'admin', timestamp: '2024-02-19 10:00:00', details: '发布配置 v2 到生产环境', status: 'success' },
  { id: 'log-9', appId: 'app-2', appName: 'order-service', namespace: 'payment', key: 'payment.gateway.config', action: 'create', operator: 'wangwu', timestamp: '2024-02-17 14:00:00', details: '创建支付网关配置', newValue: JSON.stringify({ provider: 'stripe', timeout: 30000 }), status: 'success' },
  { id: 'log-10', appId: 'app-2', appName: 'order-service', namespace: 'payment', key: 'payment.gateway.config', action: 'release', operator: 'wangwu', timestamp: '2024-02-17 14:00:00', details: '发布配置 v1 到生产环境', status: 'success' },
  { id: 'log-11', appId: 'app-1', appName: 'user-service', namespace: 'common', key: 'app.log.level', action: 'release', operator: 'devops', timestamp: '2024-02-18 16:45:00', details: '发布配置 v3 到生产环境', status: 'failed' },
  { id: 'log-12', appId: 'app-5', appName: 'gateway-service', namespace: 'auth', key: 'jwt.secret', action: 'update', operator: 'admin', timestamp: '2024-02-18 16:00:00', details: '轮换JWT密钥', status: 'success' },
  { id: 'log-13', appId: 'app-5', appName: 'gateway-service', namespace: 'auth', key: 'jwt.secret', action: 'release', operator: 'admin', timestamp: '2024-02-18 16:00:00', details: '发布配置 v2 到生产环境', status: 'success' },
];

// Config Audit Logs (alias to auditLogs for compatibility) 
export const configAuditLogs: AuditLog[] = auditLogs;

// Releases
export const releases: Release[] = [
  {
    id: 'release-1',
    appId: 'app-1',
    namespace: 'production',
    key: 'database.url',
    env: 'prod',
    fromVersion: 1,
    toVersion: 2,
    releasedBy: 'admin',
    releasedAt: '2026-01-15T14:30:00Z',
    status: 'success',
    comment: '切换到生产数据库',
  },
  {
    id: 'release-2',
    appId: 'app-1',
    namespace: 'production',
    key: 'app.features',
    env: 'prod',
    fromVersion: 1,
    toVersion: 2,
    releasedBy: 'admin',
    releasedAt: '2026-02-01T09:15:00Z',
    status: 'success',
    comment: '发布Feature X给最终用户',
  },
];

// Audit Logs
// export const auditLogs: AuditLog[] = [
//   {
//     id: 'audit-1',
//     appId: 'app-1',
//     appName: 'user-service',
//     namespace: 'production',
//     key: 'database.url',
//     action: 'update',
//     operator: 'admin',
//     timestamp: '2026-01-15T14:00:00Z',
//     details: 'Update database.url from version 1 to 2',
//     status: 'success',
//     oldValue: 'jdbc:mysql://test-db.example.com:3306/users',
//     newValue: 'jdbc:mysql://prod-db.example.com:3306/users',
//   },
//   {
//     id: 'audit-2',
//     appId: 'app-1',
//     appName: 'user-service',
//     namespace: 'production',
//     key: 'app.features',
//     action: 'update',
//     operator: 'admin',
//     timestamp: '2026-02-01T09:00:00Z',
//     details: 'Update app.features from version 1 to 2',
//     status: 'success',
//     oldValue: '{"feature-x": false, "feature-y": false}',
//     newValue: '{"feature-x": true, "feature-y": false}',
//   },
//   {
//     id: 'audit-3',
//     appId: 'app-1',
//     appName: 'user-service',
//     namespace: 'production',
//     key: 'app.features',
//     action: 'release',
//     operator: 'admin',
//     timestamp: '2026-02-01T09:15:00Z',
//     details: 'Release version 2 of app.features',
//     status: 'success',
//   },
// ];

// Templates
export const templates: ConfigTemplate[] = [
  {
    id: 'tmpl-1',
    name: 'Database Configuration Template',
    description: '标准数据库配置模板',
    format: 'json',
    content: `{
  "database": {
    "url": "{{ DB_HOST }}:{{ DB_PORT }}/{{ DB_NAME }}",
    "username": "{{ DB_USER }}",
    "password": "{{ DB_PASSWORD }}",
    "options": {
      "poolSize": {{ DB_POOL_SIZE }}
    }
  }
}`,
    category: 'infrastructure',
  },
  {
    id: 'tmpl-2',
    name: 'Redis Configuration Template',
    description: '标准Redis缓存配置模板',
    format: 'text',
    content: `redis.host = {{ REDIS_HOST }}
redis.port = {{ REDIS_PORT }}
redis.password = {{ REDIS_PASSWORD }}
redis.cluster_enabled = {{ REDIS_CLUSTER_MODE }}
redis.connection_timeout = {{ REDIS_TIMEOUT_MS }}`,
    category: 'infrastructure',
  },
];

export const mockData = {
  hosts,
  tasks,
  taskLogs,
  k8sClusters,
  k8sNodes,
  k8sPods,
  k8sServices,
  k8sIngresses,
  alerts,
  alertRules,
  monitorMetrics,
  integrationTools,
  configs,
  services,
  serviceQuotas,
  configApps,
  configItems,
  configAuditLogs,
  releases,
  auditLogs,
  templates,
};

// Dashboard related
export interface DashboardSummary {
  hostStats: {
    total: number;
    physical: number;
    virtual: number;
    container: number;
    offline: number;
    trend: number;
  };
  statusOverview: {
    running: number;
    maintenance: number;
    warning: number;
    offline: number;
  };
  resourceUtilization: {
    cpuAvg: number;
    cpuTrend: number;
    memoryAvg: number;
    memoryTrend: number;
  };
}

export interface Pipeline {
  id: string;
  name: string;
  status: 'pending' | 'running' | 'success' | 'failed';
  startTime: string;
  duration: number;
  trigger: string;
}

export const mockDashboardSummary: DashboardSummary = {
  hostStats: {
    total: 128,
    physical: 46,
    virtual: 65,
    container: 17,
    offline: 5,
    trend: 12,
  },
  statusOverview: {
    running: 115,
    maintenance: 8,
    warning: 0,
    offline: 5,
  },
  resourceUtilization: {
    cpuAvg: 52.3,
    cpuTrend: 2.1,
    memoryAvg: 68.7,
    memoryTrend: 3.5,
  }
};

export const mockRecentPipelines: Pipeline[] = [
  { id: 'pl-1', name: 'Web Deploy', status: 'success', startTime: '2024-02-20 02:15:00', duration: 285, trigger: 'daily-schedule' },
  { id: 'pl-2', name: 'API Build', status: 'success', startTime: '2024-02-20 02:10:00', duration: 142, trigger: 'ci-commit' },
  { id: 'pl-3', name: 'DB Migrate', status: 'success', startTime: '2024-02-20 02:05:00', duration: 356, trigger: 'manual-trigger' },
  { id: 'pl-4', name: 'Monitor Update', status: 'success', startTime: '2024-02-20 01:45:00', duration: 98, trigger: 'weekly-schedule' },
  { id: 'pl-5', name: 'Notification Service', status: 'failed', startTime: '2024-02-20 01:30:00', duration: 213, trigger: 'ci-commit' },
];

export const generateResourceTrend = (baseValue: number, points: number, variance: number): number[] => {
  let values: number[] = [];
  let currentValue = baseValue;
  for (let i = 0; i < points; i++) {
    currentValue += (Math.random() - 0.5) * variance;
    currentValue = Math.max(0, Math.min(100, currentValue)); // Keep between 0 and 100
    values.push(parseFloat(currentValue.toFixed(1)));
  }
  return values;
};
