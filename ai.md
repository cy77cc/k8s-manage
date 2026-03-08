这份 `specs.md` 规范文档是基于 **Eino** 框架（Go 语言生态）设计的 AI-PaaS 平台核心架构指南。它重点解决了 AI 决策的**确定性**、**安全性**以及与**云原生基础设施**的深度集成。

---

# AI-PaaS Platform Architecture Specs (Eino-Based)

## 1. 概述 (Overview)

本规范定义了基于 Eino 框架构建的 AI 驱动 PaaS 平台的架构标准。核心目标是将 LLM 的推理能力转化为稳定、可预测的集群操作指令。

---

## 2. 核心组件规范 (Core Components)

### 2.1 智能路由层 (Intent Router)

* **组件类型**: `eino.Router`
* **职责**: 解析用户输入的自然语言（NLP）或系统告警（Alerts），将其分发至特定的业务图（Graph）。
* **技术要求**:
* 必须支持 **多级路由**：一级按功能（部署、诊断、优化），二级按资源类型（Pod, Service, Ingress）。
* **Fallback 机制**: 无法识别意图时，必须路由至 `General_Assistance` 节点。



### 2.2 知识检索增强 (RAG Engine)

* **组件类型**: `eino.Indexer` & `eino.Retriever`
* **数据源**:
1. **静态**: 平台 API 文档、Helm Chart 规范。
2. **动态**: K8s 集群实时拓扑、历史日志、Prometheus 指标。


* **实现要点**:
* 使用 **Eino Filter** 注入多租户隔离标识（Namespace ID）。
* 检索结果必须包含 `Source_Metadata` 以供追溯。



### 2.3 确定性工作流 (Action Graph)

* **组件类型**: `eino.Graph`
* **结构规范**:
* **Pre-Processing**: 对 Prompt 进行脱敏处理，移除用户敏感凭据。
* **Reasoning**: LLM 生成逻辑计划（JSON 格式）。
* **Validation (Lambda)**: 强类型的 Go 结构体校验，确保生成的 YAML 符合集群 Open API 规范。
* **Execution (Tool)**: 封装 `client-go` 接口进行实际的资源变更。



---

## 3. 技术细节与 Eino 特性集成

### 3.1 安全切面 (Security Aspects)

* **规范**: 所有涉及 `WRITE` 操作（如 `Create`, `Update`, `Delete`）的 Tool 调用，必须绑定 Eino **Aspect**。
* **功能**:
* **Dry-run 拦截**: 在实际提交 K8s 前，先调用 `server-side dry-run`。
* **二次确认 (Human-in-the-loop)**: 对于高风险操作（如删除整个 Namespace），Aspect 必须挂起当前 Graph 并等待人工审批信号。



### 3.2 内存管理 (Memory Management)

* **短期记忆**: 存储在 `eino.Memory` 中，生命周期为一个会话（Session）。
* **长期记忆**: 存储在向量数据库（如 Milvus），用于记录特定应用的“亚健康”历史模式。

### 3.3 监控与追踪 (Observability)

* **规范**: 利用 Eino 的 `Callout` 机制。
* **要求**: 每一层推理的 Input/Output 必须带有 `TraceID` 并上报至 OpenTelemetry。
* **诊断日志**: AI 的每一步逻辑推导路径必须可视化。

---

## 4. 数据交互协议 (Data Contract)

AI 模块的输出必须严格遵循以下 JSON 模式：

```json
{
  "action_type": "K8S_RESOURCE_PATCH",
  "resource": "Deployment",
  "metadata": {
    "name": "api-server",
    "namespace": "prod"
  },
  "payload": {
    "replicas": 5
  },
  "reasoning": "Detected high CPU usage (85%) in last 10 mins. Scaling up to ensure stability."
}

```

---

## 5. 开发与部署规范

* **开发语言**: Go 1.21+
* **AI 驱动引擎**: 支持多种 Provider（OpenAI, Ollama, DeepSeek），通过 `eino.ChatModel` 接口实现解耦。
* **本地开发**: 建议使用 **Ollama** 进行离线调试，通过 Eino 的环境变量配置快速切换模型端点。

---