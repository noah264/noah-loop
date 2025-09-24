# Noah Loop 模块系统

## 概述

Noah Loop 后端采用微服务架构，由四个核心模块组成，每个模块都有特定的职责和功能。这些模块通过标准的 API 接口进行通信，共同构建了一个完整的智能代理系统平台。

## 模块架构图

```
┌─────────────────────────────────────────────────────────────┐
│                     Noah Loop 系统                          │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌──────────────┐
│  │   Agent     │  │     LLM     │  │     MCP     │  │ Orchestrator │
│  │    模块     │  │    模块     │  │    模块     │  │     模块     │
│  │   :8081     │  │   :8082     │  │   :8083     │  │    :8084     │
│  └─────────────┘  └─────────────┘  └─────────────┘  └──────────────┘
│         │                │               │                 │
│         └────────────────┼───────────────┼─────────────────┘
│                          │               │
│  ┌─────────────────────────────────────────────────────────────┐
│  │                    共享基础设施                           │
│  │  • 数据库 (PostgreSQL)  • 缓存 (Redis)                    │
│  │  • 消息队列 (RabbitMQ)   • 配置管理                       │
│  └─────────────────────────────────────────────────────────────┘
└─────────────────────────────────────────────────────────────┘
```

## 模块介绍

### 🤖 Agent 模块 (端口: 8081)
**智能代理服务**

负责管理和执行各种类型的智能代理，提供完整的代理生命周期管理。

**核心功能：**
- 多类型代理支持（对话型、任务型、反思型、规划型、多模态）
- 记忆管理系统
- 工具集成和执行
- 学习和适应能力
- 状态管理

**适用场景：**
- 智能客服
- 任务自动化
- 知识问答
- 创意写作

📖 [详细文档](./agent/README.md)

### 🧠 LLM 模块 (端口: 8082)
**大语言模型服务**

提供统一的大语言模型接口，支持多种模型提供商和模型类型。

**核心功能：**
- 多提供商支持（OpenAI、Anthropic、本地模型）
- 统一的模型调用接口
- 智能路由和负载均衡
- 成本控制和监控
- 缓存和优化

**支持的模型类型：**
- 聊天模型（Chat）
- 补全模型（Completion）
- 嵌入模型（Embedding）
- 图像模型（Image）
- 音频模型（Audio）

📖 [详细文档](./llm/README.md)

### 💬 MCP 模块 (端口: 8083)
**会话和上下文管理服务**

管理用户与代理之间的会话状态、上下文信息和对话历史。

**核心功能：**
- 会话生命周期管理
- 智能上下文管理
- 相关性计算和搜索
- 自动清理和优化
- 内存管理

**适用场景：**
- 多轮对话
- 上下文相关查询
- 会话历史管理
- 个性化推荐

📖 [详细文档](./mcp/README.md)

### ⚡ Orchestrator 模块 (端口: 8084)
**工作流编排服务**

设计、管理和执行复杂的工作流程，实现业务流程自动化。

**核心功能：**
- 可视化工作流设计
- 多触发器支持
- 智能调度和执行
- 条件分支和并行处理
- 执行监控和分析

**触发器类型：**
- 手动触发
- 定时触发（Cron）
- 事件触发
- Webhook 触发
- 条件触发

📖 [详细文档](./orchestrator/README.md)

## 模块间通信

### 通信方式
- **HTTP/REST API**: 主要通信方式
- **消息队列**: 异步事件通知
- **共享数据库**: 数据持久化
- **服务发现**: 动态服务注册和发现

### 典型交互流程

#### 1. 用户对话流程
```mermaid
sequenceDiagram
    participant User
    participant MCP
    participant Agent
    participant LLM
    
    User->>MCP: 创建会话
    MCP->>Agent: 获取代理信息
    User->>MCP: 发送消息
    MCP->>Agent: 处理用户消息
    Agent->>LLM: 调用语言模型
    LLM->>Agent: 返回响应
    Agent->>MCP: 更新会话上下文
    MCP->>User: 返回响应
```

#### 2. 工作流执行流程
```mermaid
sequenceDiagram
    participant Trigger
    participant Orchestrator
    participant Agent
    participant LLM
    participant MCP
    
    Trigger->>Orchestrator: 触发工作流
    Orchestrator->>Agent: 执行代理任务
    Agent->>LLM: 调用模型
    Agent->>MCP: 创建临时会话
    Agent->>Orchestrator: 返回执行结果
    Orchestrator->>Trigger: 工作流完成
```

## 统一配置

### 配置文件结构
```yaml
# configs/config.yaml
database:
  url: "postgres://user:password@localhost/noah_loop?sslmode=disable"
  max_open_conns: 100
  max_idle_conns: 10

services:
  agent:
    port: 8081
  llm:
    port: 8082
  mcp:
    port: 8083
  orchestrator:
    port: 8084

cache:
  type: redis
  url: "redis://localhost:6379/0"

logging:
  level: info
  format: json
```

### 环境变量
```bash
# 数据库配置
DATABASE_URL="postgres://user:password@localhost/noah_loop"

# LLM 提供商
OPENAI_API_KEY="sk-xxx"
ANTHROPIC_API_KEY="sk-ant-xxx"

# Redis配置
REDIS_URL="redis://localhost:6379/0"

# 日志级别
LOG_LEVEL="info"
```

## 快速启动

### 使用 Docker Compose

```bash
# 克隆项目
git clone <repository-url>
cd noah-loop

# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 本地开发启动

```bash
# 启动数据库和缓存
docker-compose up -d postgres redis

# 启动各个模块
cd backend/modules/agent && go run cmd/main.go &
cd backend/modules/llm && go run cmd/main.go &
cd backend/modules/mcp && go run cmd/main.go &
cd backend/modules/orchestrator && go run cmd/main.go &
```

### 健康检查

```bash
# 检查所有服务状态
curl http://localhost:8081/health  # Agent
curl http://localhost:8082/health  # LLM
curl http://localhost:8083/health  # MCP
curl http://localhost:8084/health  # Orchestrator
```

## 使用示例

### 1. 创建智能代理
```bash
# 创建一个对话型代理
curl -X POST http://localhost:8081/api/v1/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "智能助手",
    "type": "conversational",
    "description": "通用智能助手",
    "system_prompt": "你是一个有用的AI助手",
    "owner_id": "user-uuid"
  }'
```

### 2. 配置语言模型
```bash
# 注册 GPT-3.5 模型
curl -X POST http://localhost:8082/api/v1/models \
  -H "Content-Type: application/json" \
  -d '{
    "name": "gpt-3.5-turbo",
    "provider": "openai",
    "type": "chat",
    "max_tokens": 4096,
    "price_per_k": 0.002
  }'
```

### 3. 创建对话会话
```bash
# 创建会话
curl -X POST http://localhost:8083/api/v1/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-uuid",
    "agent_id": "agent-uuid",
    "title": "日常对话"
  }'
```

### 4. 设计工作流
```bash
# 创建简单的数据处理工作流
curl -X POST http://localhost:8084/api/v1/workflows \
  -H "Content-Type: application/json" \
  -d '{
    "name": "数据分析工作流",
    "description": "自动化数据分析流程",
    "definition": {
      "steps": [
        {
          "id": "analyze",
          "name": "数据分析",
          "type": "agent",
          "config": {
            "agent_id": "analyst-agent"
          }
        }
      ]
    },
    "owner_id": "user-uuid"
  }'
```

## 监控和运维

### 指标收集
所有模块都提供 Prometheus 指标：
- `http://localhost:808x/metrics`

### 日志聚合
推荐使用 ELK Stack 或类似方案：
```yaml
logging:
  outputs:
    - type: elasticsearch
      hosts: ["http://elasticsearch:9200"]
    - type: file
      path: "/var/log/noah-loop/"
```

### 性能监控
关键指标：
- **API 响应时间**
- **数据库连接数**
- **缓存命中率**
- **内存和CPU使用量**
- **错误率**

## 扩展开发

### 添加新的步骤类型
```go
// 在 Orchestrator 模块中
type CustomStepExecutor struct{}

func (e *CustomStepExecutor) GetType() StepType {
    return "custom"
}

func (e *CustomStepExecutor) Execute(ctx context.Context, step *Step, input map[string]interface{}) (*StepExecutionResult, error) {
    // 实现自定义逻辑
    return result, nil
}
```

### 添加新的LLM提供商
```go
// 在 LLM 模块中
type CustomProvider struct{}

func (p *CustomProvider) GetName() string {
    return "custom"
}

func (p *CustomProvider) CreateChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
    // 实现提供商接口
    return response, nil
}
```

## 故障排除

### 常见问题

1. **模块启动失败**
   - 检查端口占用：`netstat -tlnp | grep :808x`
   - 检查数据库连接：确认 PostgreSQL 服务状态
   - 查看日志：`docker-compose logs module-name`

2. **模块间通信失败**
   - 检查网络连通性
   - 验证服务发现配置
   - 检查防火墙设置

3. **数据库连接问题**
   - 检查连接字符串
   - 验证数据库权限
   - 检查连接池配置

4. **内存使用过高**
   - 检查缓存配置
   - 监控垃圾回收
   - 调整内存限制

### 调试工具

```bash
# 检查所有服务状态
make health-check

# 查看详细日志
make logs service=agent

# 数据库连接测试
make db-test

# 性能测试
make benchmark
```

## 开发环境设置

### 依赖安装
```bash
# Go 依赖
go mod download

# 开发工具
make install-tools

# 数据库迁移
make migrate
```

### 代码生成
```bash
# Wire 依赖注入
go generate ./...

# API 文档
make docs

# Protocol Buffers (如果使用)
make proto
```

### 测试
```bash
# 单元测试
make test

# 集成测试
make test-integration

# E2E 测试
make test-e2e
```

## 部署指南

### 生产环境配置
```yaml
# docker-compose.prod.yml
version: '3.8'
services:
  agent:
    image: noah-loop/agent:latest
    deploy:
      replicas: 3
      resources:
        limits:
          cpus: '1.0'
          memory: 1G
  
  # ... 其他服务配置
```

### Kubernetes 部署
```bash
# 部署到 k8s
kubectl apply -f k8s/

# 检查状态
kubectl get pods -n noah-loop

# 查看日志
kubectl logs -f deployment/agent -n noah-loop
```

### 负载均衡
推荐使用 Nginx 或 Traefik：
```nginx
upstream agent_backend {
    server agent1:8081;
    server agent2:8081;
    server agent3:8081;
}

server {
    listen 80;
    location /api/v1/agents {
        proxy_pass http://agent_backend;
    }
}
```

## 版本兼容性

| 模块版本 | 兼容性 | 说明 |
|---------|--------|------|
| v1.0.x  | ✅ | 稳定版本 |
| v1.1.x  | ✅ | 向后兼容 |
| v1.2.x  | ⚠️  | API 变更 |
| v2.0.x  | ❌ | 重大更新 |

## 社区和支持

- **问题反馈**: [GitHub Issues](https://github.com/noah-loop/issues)
- **功能请求**: [GitHub Discussions](https://github.com/noah-loop/discussions)
- **文档贡献**: [文档仓库](https://github.com/noah-loop/docs)
- **社区论坛**: [论坛链接](https://community.noah-loop.com)

## 许可证

MIT License - 查看 [LICENSE](../LICENSE) 文件了解详情。
