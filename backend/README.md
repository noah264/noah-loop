# Noah Loop 智能体平台后端

<div align="center">

![Noah Loop](https://img.shields.io/badge/Noah%20Loop-v1.0.0-blue.svg)
![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8.svg)
![Architecture](https://img.shields.io/badge/Architecture-DDD%20Microservices-green.svg)
![License](https://img.shields.io/badge/License-MIT-yellow.svg)

基于 DDD（领域驱动设计）架构的智能体平台后端系统，提供完整的智能体生命周期管理、大模型集成、工作流编排和知识管理功能。

[快速开始](#快速开始) • [架构设计](#架构设计) • [API 文档](#api-文档) • [部署指南](#部署指南) • [开发指南](#开发指南)

</div>

## ✨ 核心特性

### 🤖 智能体管理
- **多类型智能体**：支持对话型、任务型、反思型、规划型、多模态智能体
- **记忆系统**：短期记忆、长期记忆、学习记忆的完整管理
- **工具集成**：灵活的工具插件系统，支持自定义工具开发
- **状态管理**：完整的智能体状态跟踪和生命周期管理

### 🧠 大语言模型服务
- **多提供商支持**：OpenAI、Anthropic、本地模型无缝集成
- **统一接口**：标准化的模型调用接口，支持切换和负载均衡
- **成本控制**：实时成本监控、预算管理和智能路由
- **缓存优化**：智能缓存策略提升响应速度

### 💬 会话管理
- **上下文管理**：智能的对话上下文保持和相关性计算
- **多轮对话**：支持长期会话状态管理
- **个性化**：基于用户历史的个性化推荐

### ⚡ 工作流编排
- **可视化设计**：拖拽式工作流设计界面
- **多触发器**：支持手动、定时、事件、Webhook 等触发方式
- **并行处理**：智能任务调度和并行执行
- **监控分析**：实时执行监控和性能分析

### 📚 知识管理 (RAG)
- **多格式文档**：支持文本、PDF、Markdown、HTML、Word 等格式
- **智能分块**：自动语义分块和向量化
- **语义搜索**：基于向量相似度的智能搜索
- **知识库管理**：多租户知识库和权限控制

### 📬 通知服务 (Notify)
- **多渠道支持**：邮件(SMTP)、短信(阿里云)、推送(Bark)、Webhook(Server酱)
- **模板系统**：灵活的通知模板管理和变量替换
- **批量发送**：支持批量通知和接收者管理
- **重试机制**：智能重试和失败恢复

### 📨 消息队列 (Kafka)
- **事件驱动**：完整的事件驱动架构支持
- **异步通信**：服务间高性能异步消息传递
- **可靠投递**：消息持久化和重试保障
- **分布式追踪**：消息链路完整可观测性

## 🏗️ 架构设计

### 系统架构图
```
┌─────────────────────────────────────────────────────────────────────┐
│                          Noah Loop 智能体平台                       │
├─────────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐                                                    │
│  │  API Gateway │ ← 统一入口，路由分发，认证鉴权，负载均衡              │
│  │   (8080)     │                                                    │
│  └─────┬────────┘                                                    │
│        │                                                             │
│  ┌─────▼──────────────────────────────────────────────────────────┐ │
│  │                      核心微服务模块                             │ │
│  │  ┌───────┐ ┌───────┐ ┌───────┐ ┌─────────┐ ┌───────┐ ┌───────┐ │ │
│  │  │Agent  │ │ LLM   │ │ MCP   │ │Orchestr.│ │ RAG   │ │Notify │ │ │
│  │  │:8081  │ │:8082  │ │:8083  │ │ :8084   │ │:8085  │ │:8086  │ │ │
│  │  │智能体 │ │大模型 │ │会话   │ │工作流   │ │知识   │ │通知   │ │ │
│  │  └───────┘ └───────┘ └───────┘ └─────────┘ └───────┘ └───────┘ │ │
│  └─────────────────────────────────────────────────────────────────┘ │
│                                    │                                  │
│  ┌─────────────────────────────────▼──────────────────────────────┐  │
│  │                        消息和事件层                            │  │
│  │           ┌─────────────────────────────────────────┐          │  │
│  │           │           Apache Kafka                  │          │  │
│  │           │  ┌─────────┐ ┌─────────┐ ┌─────────────┐ │          │  │
│  │           │  │ Events  │ │ Notify  │ │    Logs     │ │          │  │
│  │           │  └─────────┘ └─────────┘ └─────────────┘ │          │  │
│  │           └─────────────────────────────────────────┘          │  │
│  └──────────────────────────────────────────────────────────────────┘ │
│                                    │                                  │
│  ┌─────────────────────────────────▼──────────────────────────────┐  │
│  │                        共享基础设施层                          │  │
│  │ PostgreSQL │ Redis │ etcd │ Jaeger │ Prometheus │ MinIO │ Milvus │ │
│  │    数据库  │ 缓存  │服务发现│ 链路追踪│   监控     │对象存储│向量库 │ │
│  └──────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
```

### 技术栈

#### 核心技术
- **后端语言**: Go 1.21+ 
- **Web框架**: Gin (HTTP), gRPC (微服务通信)
- **ORM**: GORM (数据库操作)
- **架构模式**: DDD (领域驱动设计), 微服务架构
- **依赖注入**: Google Wire (编译时依赖注入)

#### 数据存储
- **关系数据库**: PostgreSQL (主数据库)
- **向量数据库**: Milvus (向量搜索和相似度计算)
- **缓存系统**: Redis (会话、缓存)
- **对象存储**: MinIO (文件存储)

#### 基础设施
- **消息队列**: Apache Kafka (事件驱动、异步通信)
- **服务发现**: etcd (服务注册、配置管理、密钥存储)
- **链路追踪**: Jaeger + OpenTelemetry (分布式追踪)
- **监控告警**: Prometheus + Grafana (指标监控)
- **容器化**: Docker + Docker Compose

#### 开发工具
- **代码生成**: Wire (DI), Protocol Buffers (gRPC)
- **热重载**: Air (开发环境)
- **代码质量**: golangci-lint, gofmt, goimports
- **测试框架**: Testify, GoConvey

### 模块详细介绍

| 模块 | HTTP | gRPC | 功能描述 | 技术特点 |
|------|------|------|----------|----------|
| **api-gateway** | 8080 | 9090 | API网关，统一入口 | 路由分发、认证鉴权、限流、负载均衡 |
| **agent** | 8081 | 9091 | 智能体核心服务 | 多类型智能体、记忆管理、工具集成 |
| **llm** | 8082 | 9092 | 大语言模型服务 | 多提供商、统一接口、智能路由 |
| **mcp** | 8083 | 9093 | 会话上下文管理 | 上下文管理、会话状态、个性化 |
| **orchestrator** | 8084 | 9094 | 工作流编排引擎 | 流程设计、任务调度、并行执行 |
| **rag** | 8085 | 9095 | 知识检索服务 | 文档管理、向量搜索、语义检索 |
| **notify** | 8086 | 9096 | 多渠道通知服务 | 邮件、短信、推送、模板管理 |
| **shared** | - | - | 共享基础设施 | DDD基础、配置管理、Kafka、etcd |

## 🚀 快速开始

### 环境要求
- **Go**: 1.21 或更高版本
- **Docker**: 20.10 或更高版本
- **Docker Compose**: 2.0 或更高版本
- **内存**: 建议 4GB 以上
- **磁盘**: 建议 10GB 以上空闲空间

### 一键启动（推荐）

```bash
# 克隆项目
git clone <repository-url>
cd noah-loop

# 启动所有基础设施和服务
make up

# 检查服务状态
make status

# 查看日志
make logs
```

### 手动启动

#### 1. 启动基础设施
```bash
cd backend

# 启动完整基础设施（推荐）
docker-compose -f deployments/docker-compose.infrastructure.yml up -d

# 或者启动开发环境
docker-compose -f deployments/docker-compose.yml up -d

# 等待所有服务启动完成
docker-compose ps
```

#### 2. 设置环境变量
```bash
# 复制环境变量模板
cp .env.example .env

# 编辑环境变量
nano .env
```

关键环境变量：
```bash
# OpenAI配置
OPENAI_API_KEY=sk-your-openai-key

# Anthropic配置  
ANTHROPIC_API_KEY=sk-ant-your-anthropic-key

# 数据库配置
DATABASE_URL=postgres://postgres:postgres@localhost:5432/agent_db?sslmode=disable

# Redis配置
REDIS_URL=redis://localhost:6379/0

# etcd配置
ETCD_ENDPOINTS=localhost:2379

# Jaeger配置
JAEGER_ENDPOINT=http://localhost:14268/api/traces

# Kafka配置
KAFKA_BROKERS=localhost:9092

# 通知服务配置（可选）
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password

# 阿里云短信配置（可选）
ALIYUN_ACCESS_KEY=your-access-key
ALIYUN_SECRET_KEY=your-secret-key

# Bark推送配置（可选）
BARK_DEVICE_KEY=your-bark-device-key
```

#### 3. 安装依赖和构建
```bash
# 同步工作空间依赖
go work sync

# 安装开发工具
make install-tools

# 生成代码（Wire依赖注入、Protobuf等）
make generate

# 构建所有模块
make build
```

#### 4. 启动各个服务
```bash
# 方式1：使用Makefile（推荐）
make run-all

# 方式2：分别启动各个模块
make run-gateway      # API网关 (8080)
make run-agent        # 智能体服务 (8081) 
make run-llm          # 大模型服务 (8082)
make run-mcp          # 会话管理 (8083)
make run-orchestrator # 工作流编排 (8084)
make run-rag          # 知识检索 (8085)
make run-notify       # 通知服务 (8086)

# 方式3：直接使用Go命令
cd api-gateway && go run cmd/main.go &
cd modules/agent && go run cmd/main.go &
cd modules/llm && go run cmd/main.go &
cd modules/mcp && go run cmd/main.go &
cd modules/orchestrator && go run cmd/main.go &
cd modules/rag && go run cmd/main.go &
cd modules/notify && go run cmd/main.go &
```

### 验证安装

```bash
# 检查所有服务健康状态
make health-check

# 或者手动检查
curl http://localhost:8080/health  # API Gateway
curl http://localhost:8081/health  # Agent
curl http://localhost:8082/health  # LLM  
curl http://localhost:8083/health  # MCP
curl http://localhost:8084/health  # Orchestrator
curl http://localhost:8085/health  # RAG
curl http://localhost:8086/health  # Notify
```

预期响应：
```json
{
  "status": "ok",
  "timestamp": "2024-01-20T10:00:00Z",
  "version": "1.0.0",
  "dependencies": {
    "database": "connected",
    "redis": "connected",
    "etcd": "connected"
  }
}
```

## 📖 API 文档

### RESTful API 概览

#### 智能体管理
```bash
# 创建智能体
POST /api/v1/agents
{
  "name": "我的智能助手",
  "type": "conversational",
  "description": "通用对话助手",
  "system_prompt": "你是一个有用的AI助手",
  "capabilities": ["text_processing", "question_answering"]
}

# 获取智能体列表
GET /api/v1/agents?page=1&limit=20&type=conversational

# 与智能体对话
POST /api/v1/agents/{id}/chat
{
  "message": "你好，请介绍一下人工智能的发展历史",
  "session_id": "session-uuid"
}
```

#### 大模型服务
```bash
# 聊天补全
POST /api/v1/chat/completions
{
  "model": "gpt-3.5-turbo",
  "messages": [
    {"role": "system", "content": "你是一个专业的AI助手"},
    {"role": "user", "content": "什么是机器学习？"}
  ],
  "temperature": 0.7,
  "max_tokens": 1000
}

# 文本嵌入
POST /api/v1/embeddings
{
  "model": "text-embedding-ada-002",
  "input": ["人工智能", "机器学习", "深度学习"]
}
```

#### 工作流管理
```bash
# 创建工作流
POST /api/v1/workflows
{
  "name": "数据分析流程",
  "description": "自动化数据分析工作流",
  "definition": {
    "steps": [
      {
        "id": "data_process",
        "name": "数据处理", 
        "type": "agent",
        "config": {"agent_id": "data-agent"}
      }
    ]
  }
}

# 执行工作流
POST /api/v1/workflows/{id}/execute
{
  "input": {"data": "分析数据"}
}
```

#### 知识管理
```bash
# 创建知识库
POST /api/v1/knowledge-bases
{
  "name": "技术文档库",
  "description": "存储技术相关文档",
  "settings": {
    "chunk_size": 1000,
    "embedding_model": "text-embedding-ada-002"
  }
}

# 添加文档
POST /api/v1/documents
{
  "title": "API设计指南",
  "content": "本文介绍RESTful API设计最佳实践...",
  "knowledge_base_id": "kb-uuid"
}

# 语义搜索
POST /api/v1/search
{
  "query": "如何设计RESTful API",
  "knowledge_base_id": "kb-uuid",
  "top_k": 5
}
```

#### 通知服务
```bash
# 发送邮件通知
POST /api/v1/notifications
{
  "title": "系统维护通知",
  "content": "系统将在今晚22:00进行维护",
  "type": "system",
  "channel": "email",
  "recipients": [
    {
      "type": "email",
      "identifier": "admin@company.com"
    }
  ]
}

# 创建通知模板
POST /api/v1/templates
{
  "name": "欢迎邮件模板",
  "code": "welcome_email",
  "type": "html",
  "subject": "欢迎加入{{company_name}}",
  "content": "亲爱的{{username}}，欢迎加入我们！",
  "variables": [
    {
      "name": "username",
      "required": true
    },
    {
      "name": "company_name",
      "default_value": "Noah Loop"
    }
  ]
}

# 从模板发送通知
POST /api/v1/notifications/template
{
  "template_id": "welcome_email",
  "channel": "email",
  "variables": {
    "username": "张三",
    "company_name": "我的公司"
  },
  "recipients": [
    {
      "type": "email", 
      "identifier": "user@example.com"
    }
  ]
}
```

### API 认证

系统支持多种认证方式：

```bash
# Bearer Token认证
Authorization: Bearer <your-api-token>

# API Key认证
X-API-Key: <your-api-key>

# Basic Auth认证
Authorization: Basic <base64-encoded-credentials>
```

### 详细API文档
- [智能体 API](./modules/agent/README.md#api-文档)
- [大模型 API](./modules/llm/README.md#api-文档)
- [会话管理 API](./modules/mcp/README.md#api-文档)
- [工作流 API](./modules/orchestrator/README.md#api-文档)  
- [知识检索 API](./modules/rag/README.md#api-文档)
- [通知服务 API](./modules/notify/README.md#api-文档)

### 集成指南
- [Kafka消息队列集成指南](./KAFKA_INTEGRATION_GUIDE.md)
- [分布式追踪指南](./DISTRIBUTED_TRACING_GUIDE.md)
- [服务部署指南](./SERVICES_DEPLOYMENT_GUIDE.md)
- [基础设施集成指南](./INFRASTRUCTURE_INTEGRATION_GUIDE.md)

## 🐳 部署指南

### Docker Compose 部署（生产环境）

```bash
# 克隆项目
git clone <repository-url>
cd noah-loop

# 配置环境变量
cp .env.example .env
# 编辑 .env 文件，设置生产环境配置

# 启动生产环境
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# 检查部署状态
docker-compose ps
```

### Kubernetes 部署

```bash
# 创建命名空间
kubectl create namespace noah-loop

# 应用配置
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secrets.yaml

# 部署基础设施
kubectl apply -f k8s/postgres.yaml
kubectl apply -f k8s/redis.yaml  
kubectl apply -f k8s/etcd.yaml

# 部署应用服务
kubectl apply -f k8s/
```

### 监控面板访问

部署完成后，可以通过以下地址访问监控面板：

| 服务 | 地址 | 账号密码 | 功能说明 |
|------|------|----------|----------|
| **Grafana** | http://localhost:3000 | admin / admin | 可视化监控面板 |
| **Prometheus** | http://localhost:9090 | - | 指标收集和查询 |
| **Jaeger** | http://localhost:16686 | - | 分布式链路追踪 |
| **Kafka UI** | http://localhost:8090 | - | Kafka管理界面 |
| **MinIO** | http://localhost:9001 | minioadmin / minioadmin123 | 对象存储管理 |
| **etcd** | http://localhost:2379 | - | 服务发现和配置 |

## 🛠️ 开发指南

### 项目结构
```
noah-loop/
├── backend/                    # 后端服务
│   ├── api-gateway/           # API网关
│   ├── modules/               # 核心模块
│   │   ├── agent/            # 智能体模块  
│   │   ├── llm/              # 大模型模块
│   │   ├── mcp/              # 会话管理模块
│   │   ├── orchestrator/     # 工作流编排模块
│   │   ├── rag/              # 知识检索模块
│   │   └── notify/           # 通知服务模块
│   ├── shared/               # 共享基础代码
│   ├── configs/              # 配置文件
│   ├── deployments/          # 部署配置
│   └── go.work              # Go工作空间配置
├── frontend/                  # 前端应用（预留）
├── common/                    # 通用配置和脚本
└── release/                  # 发布相关文件
```

### DDD 架构层次

每个模块都遵循DDD架构模式：

```
module/
├── cmd/                      # 启动入口
├── internal/
│   ├── domain/              # 领域层
│   │   ├── entity/         # 实体
│   │   ├── valueobject/    # 值对象  
│   │   ├── service/        # 领域服务
│   │   └── repository/     # 仓储接口
│   ├── application/         # 应用层
│   │   └── service/        # 应用服务
│   ├── infrastructure/      # 基础设施层
│   │   ├── repository/     # 仓储实现
│   │   ├── config/         # 配置
│   │   └── external/       # 外部服务
│   ├── interface/          # 接口层
│   │   ├── http/          # HTTP接口
│   │   └── grpc/          # gRPC接口
│   └── wire/              # 依赖注入
├── go.mod
└── README.md
```

### 开发环境搭建

#### 1. 安装开发工具
```bash
# 安装必要工具
make install-tools

# 包括以下工具：
# - Wire: 依赖注入代码生成
# - Air: 热重载工具
# - golangci-lint: 代码检查
# - protoc: Protocol Buffers编译器
```

#### 2. 配置IDE

**VS Code 推荐插件:**
```json
{
  "recommendations": [
    "golang.go",
    "ms-vscode.vscode-json",
    "redhat.vscode-yaml",  
    "ms-kubernetes-tools.vscode-kubernetes-tools"
  ]
}
```

**GoLand 配置:**
- 启用 Go Modules
- 配置 Wire 作为代码生成器
- 设置 gofmt 和 goimports

#### 3. 代码规范

```bash
# 格式化代码
make fmt

# 代码检查
make lint

# 运行测试
make test

# 生成覆盖率报告
make test-coverage
```

### 代码生成

```bash
# 生成所有代码
make generate

# 单独生成Wire代码
make wire-gen

# 生成Protobuf代码（如需要）
make proto-gen

# 生成API文档
make docs
```

### 热重载开发

```bash
# 启动热重载开发环境
make dev

# 或者单独启动某个模块
cd modules/agent && make dev
```

### 测试

```bash
# 运行所有测试
make test

# 运行单个模块测试
make test-agent
make test-llm
make test-rag
make test-notify

# 集成测试
make test-integration

# E2E测试
make test-e2e

# 性能测试
make benchmark
```

## 📊 监控运维

### 系统监控

#### Prometheus 指标
所有服务都暴露 Prometheus 指标：
```
# HTTP请求指标
http_requests_total
http_request_duration_seconds

# 业务指标  
agent_requests_total
llm_requests_total
workflow_executions_total
notification_sent_total
kafka_messages_produced_total
kafka_messages_consumed_total

# 系统指标
go_memstats_alloc_bytes
go_memstats_gc_duration_seconds
```

#### Grafana 仪表板
预置的监控面板包括：
- **系统概览**: CPU、内存、网络使用情况
- **应用性能**: 请求量、响应时间、错误率
- **业务指标**: 智能体使用情况、工作流执行情况
- **通知监控**: 各渠道发送成功率、失败原因分析
- **消息队列**: Kafka消息生产消费监控、延迟统计
- **数据库监控**: 连接数、查询性能
- **缓存监控**: Redis使用情况

#### 链路追踪
Jaeger 追踪功能：
- HTTP请求链路追踪
- gRPC调用链路追踪
- 数据库查询追踪
- 外部API调用追踪
- Kafka消息链路追踪
- 通知发送链路追踪
- 自定义span标记

### 日志管理

#### 日志级别
```yaml
# 配置示例
log:
  level: "info"        # debug, info, warn, error
  format: "json"       # json, text
  output: "stdout"     # stdout, file
  file_path: "/var/log/noah-loop/"
```

#### 结构化日志
```go
// 使用示例
log.WithFields(log.Fields{
  "user_id": userID,
  "agent_id": agentID,
  "action": "create_agent",
}).Info("Agent created successfully")
```

#### 日志聚合
推荐使用 ELK Stack 或 Loki：
```yaml
# docker-compose.logging.yml
services:
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:7.17.0
  
  logstash:
    image: docker.elastic.co/logstash/logstash:7.17.0
    
  kibana:
    image: docker.elastic.co/kibana/kibana:7.17.0
```

### 性能优化

#### 数据库优化
```sql
-- 关键索引
CREATE INDEX idx_agents_owner_id ON agents(owner_id);
CREATE INDEX idx_agents_type_status ON agents(type, status);
CREATE INDEX idx_sessions_user_id ON sessions(user_id);

-- 分区表（大表优化）
CREATE TABLE agent_logs (
  id UUID PRIMARY KEY,
  created_at TIMESTAMP,
  -- 其他字段
) PARTITION BY RANGE (created_at);
```

#### 缓存策略
```yaml
# Redis缓存配置
cache:
  redis:
    addr: "localhost:6379"
    password: ""
    db: 0
  
  policies:
    agents:
      ttl: 1h
      max_size: 10000
    
    sessions:
      ttl: 24h
      max_size: 50000
```

#### 连接池配置
```yaml
# 数据库连接池
database:
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 1h
  conn_max_idle_time: 10m

# HTTP客户端池
http_client:
  max_idle_conns: 100
  max_idle_conns_per_host: 10
  idle_conn_timeout: 90s
```

## 🔧 故障排除

### 常见问题

#### 1. 服务启动失败
```bash
# 检查端口占用
netstat -tlnp | grep :8080

# 检查日志
docker-compose logs api-gateway

# 检查配置
make check-config
```

#### 2. 数据库连接问题
```bash
# 测试数据库连接
make db-test

# 检查数据库状态
docker-compose exec postgres pg_isready -U postgres

# 查看连接数
docker-compose exec postgres psql -U postgres -c "SELECT count(*) FROM pg_stat_activity;"
```

#### 3. Redis连接问题
```bash
# 测试Redis连接
docker-compose exec redis redis-cli ping

# 检查内存使用
docker-compose exec redis redis-cli info memory
```

#### 4. 模块间通信失败
```bash
# 检查服务发现
curl http://localhost:2379/v2/keys/services

# 检查网络连通性
docker-compose exec agent ping llm

# 检查防火墙设置
sudo ufw status
```

#### 5. OpenAI API 调用失败
```bash
# 测试API密钥
curl -H "Authorization: Bearer $OPENAI_API_KEY" \
  https://api.openai.com/v1/models

# 检查配额
curl -H "Authorization: Bearer $OPENAI_API_KEY" \
  https://api.openai.com/v1/usage
```

### 调试工具

```bash
# 全面健康检查
make health-check

# 检查所有服务状态
make status

# 查看特定服务日志
make logs service=agent

# 进入容器调试
docker-compose exec agent bash

# 数据库调试
make db-shell

# Redis调试  
make redis-shell
```

### 性能分析

```bash
# CPU性能分析
go tool pprof http://localhost:8081/debug/pprof/profile

# 内存分析
go tool pprof http://localhost:8081/debug/pprof/heap

# goroutine分析
go tool pprof http://localhost:8081/debug/pprof/goroutine
```

## 🤝 开发贡献

### 贡献流程

1. **Fork 项目**
   ```bash
   git clone https://github.com/your-username/noah-loop.git
   cd noah-loop
   ```

2. **创建功能分支**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **开发和测试**
   ```bash
   # 开发你的功能
   make dev
   
   # 运行测试确保功能正常
   make test
   make lint
   ```

4. **提交更改**
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   ```

5. **推送分支**
   ```bash
   git push origin feature/your-feature-name
   ```

6. **创建 Pull Request**
   - 在 GitHub 上创建 PR
   - 填写详细的PR描述
   - 确保所有检查通过

### 代码规范

#### 提交消息规范
```
<type>(<scope>): <subject>

<body>

<footer>
```

类型 (type):
- `feat`: 新功能
- `fix`: 修复bug
- `docs`: 文档更新
- `style`: 代码格式化
- `refactor`: 重构
- `test`: 测试相关
- `chore`: 构建过程或辅助工具变动

示例：
```
feat(agent): add memory management for conversational agents

- Implement short-term and long-term memory
- Add memory search and retrieval functionality
- Include memory importance scoring

Closes #123
```

#### Go 代码规范
- 遵循 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- 使用 `gofmt` 格式化代码
- 使用 `golangci-lint` 进行代码检查
- 函数和方法需要有清晰的注释
- 公共API需要有完整的文档注释

### 测试要求

#### 单元测试
```go
func TestAgentService_CreateAgent(t *testing.T) {
    // 测试用例实现
    assert := assert.New(t)
    
    agent, err := service.CreateAgent(ctx, request)
    
    assert.NoError(err)
    assert.NotNil(agent)
    assert.Equal("test-agent", agent.Name)
}
```

#### 集成测试
```go
func TestAgentIntegration(t *testing.T) {
    // 集成测试需要使用测试数据库
    testDB := setupTestDB(t)
    defer teardownTestDB(t, testDB)
    
    // 测试实现
}
```

#### 测试覆盖率要求
- 新代码测试覆盖率需要达到 80% 以上
- 核心业务逻辑需要达到 90% 以上
- 使用 `make test-coverage` 查看覆盖率报告

### 文档要求

#### API 文档
- 所有公共API都需要有详细的文档
- 使用 OpenAPI/Swagger 规范
- 包含请求/响应示例

#### 代码文档
- 所有公共函数/方法需要有注释
- 复杂逻辑需要内联注释说明
- README文档需要及时更新

### Issue 和 PR 模板

#### Bug Report 模板
```markdown
## Bug 描述
简要描述遇到的问题

## 复现步骤
1. 执行操作1
2. 执行操作2  
3. 看到错误

## 预期行为
描述你预期的正确行为

## 环境信息
- OS: [e.g. Ubuntu 20.04]
- Go版本: [e.g. 1.21.0]
- 项目版本: [e.g. v1.0.0]

## 额外信息
添加任何其他相关信息
```

#### Feature Request 模板
```markdown
## 功能描述
详细描述你希望添加的功能

## 使用场景
解释为什么需要这个功能

## 解决方案
如果有具体的实现想法，请描述

## 替代方案
是否考虑过其他解决方案？
```

## 📝 更新日志

### v1.0.0 (2024-01-20)

#### 🎉 核心功能模块
- ✨ **智能体管理系统**: 完整的多类型智能体生命周期管理
- ✨ **大语言模型集成**: 多提供商统一接口和智能路由
- ✨ **工作流编排引擎**: 可视化流程设计和并行执行
- ✨ **知识检索(RAG)系统**: 向量搜索和语义检索
- ✨ **会话上下文管理**: 智能对话状态管理
- ✨ **多渠道通知服务**: 邮件、短信、推送、Webhook支持

#### 🏗️ 基础架构设施  
- 🔧 **DDD微服务架构**: 基于领域驱动设计的清晰分层
- 📨 **Apache Kafka**: 事件驱动的消息队列系统
- 🔐 **etcd服务治理**: 服务发现、配置管理、密钥存储
- 🔍 **OpenTelemetry+Jaeger**: 完整的分布式链路追踪
- 📊 **Prometheus+Grafana**: 全方位监控和可视化
- 🗄️ **多数据存储**: PostgreSQL + Redis + Milvus + MinIO

#### 🚀 开发体验
- ⚡ **Wire依赖注入**: 编译时依赖注入和类型安全  
- 🔥 **热重载开发**: Air支持的快速开发迭代
- 🧪 **完整测试框架**: 单元、集成、E2E测试支持
- 📖 **详细文档**: API文档、部署指南、开发指南
- 🐳 **容器化部署**: Docker Compose一键部署

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](../LICENSE) 文件了解详情。

## 🙏 致谢

感谢以下开源项目对本项目的支持：
- [Go](https://golang.org/) - 高性能编程语言
- [Gin](https://gin-gonic.com/) - HTTP Web框架  
- [GORM](https://gorm.io/) - Go ORM库
- [Wire](https://github.com/google/wire) - 编译时依赖注入工具
- [Apache Kafka](https://kafka.apache.org/) - 高性能分布式消息队列
- [etcd](https://etcd.io/) - 分布式键值存储
- [PostgreSQL](https://www.postgresql.org/) - 强大的关系型数据库
- [Redis](https://redis.io/) - 高性能缓存数据库
- [Milvus](https://milvus.io/) - 向量数据库
- [Jaeger](https://www.jaegertracing.io/) - 分布式链路追踪
- [OpenTelemetry](https://opentelemetry.io/) - 可观测性框架
- [Prometheus](https://prometheus.io/) - 监控和告警系统
- [Grafana](https://grafana.com/) - 可视化监控面板

## 📞 支持与反馈

- 🐛 [报告Bug](https://github.com/noah-loop/issues/new?template=bug_report.md)
- 💡 [功能建议](https://github.com/noah-loop/issues/new?template=feature_request.md)  
- 💬 [讨论交流](https://github.com/noah-loop/discussions)
- 📧 [邮件联系](mailto:support@noah-loop.com)
- 🌟 如果这个项目对你有帮助，请给我们一个 Star！

---

<div align="center">
  
**[⬆ 回到顶部](#noah-loop-智能体平台后端)**

Made with ❤️ by Noah Loop Team

</div>

