# Agent 模块

## 概述

Agent 模块是 Noah Loop 系统的智能代理服务，负责管理和执行各种类型的智能代理。该模块提供了完整的代理生命周期管理、工具集成、记忆管理等功能。

## 主要功能

- 🤖 **多类型代理支持**：对话型、任务型、反思型、规划型、多模态代理
- 🧠 **记忆管理**：短期记忆、长期记忆、学习记忆系统
- 🔧 **工具集成**：灵活的工具插件系统
- 📊 **状态管理**：完整的代理状态跟踪
- 🎯 **学习能力**：自适应学习和知识积累

## 代理类型

### 1. 对话型代理 (Conversational)
- 适用于日常对话和问答场景
- 具备上下文理解和连续对话能力

### 2. 任务型代理 (Task)
- 专注于执行特定任务
- 支持工具调用和步骤化执行

### 3. 反思型代理 (Reflective)
- 具备自我反思和学习能力
- 可以从经验中不断优化表现

### 4. 规划型代理 (Planning)
- 擅长制定和执行复杂计划
- 支持多步骤任务分解和执行

### 5. 多模态代理 (MultiModal)
- 支持文本、图片、音频等多种模态
- 提供丰富的交互体验

## 快速开始

### 1. 安装依赖

```bash
cd backend/modules/agent
go mod download
```

### 2. 配置环境

在 `configs/config.yaml` 中配置代理服务：

```yaml
services:
  agent:
    port: 8081
    database:
      url: "postgres://user:password@localhost/noah_loop?sslmode=disable"
```

### 3. 启动服务

```bash
# 开发环境
go run cmd/main.go

# 生产环境
go build -o agent cmd/main.go
./agent
```

### 4. 验证服务

```bash
curl http://localhost:8081/health
```

## API 文档

### 代理管理

#### 创建代理
```http
POST /api/v1/agents
Content-Type: application/json

{
  "name": "我的助手",
  "type": "conversational",
  "description": "智能对话助手",
  "system_prompt": "你是一个有用的AI助手",
  "capabilities": ["text_processing", "question_answering"],
  "owner_id": "uuid-here"
}
```

#### 获取代理列表
```http
GET /api/v1/agents
```

#### 获取特定代理
```http
GET /api/v1/agents/{id}
```

#### 更新代理
```http
PUT /api/v1/agents/{id}
Content-Type: application/json

{
  "name": "更新后的助手",
  "description": "更新后的描述"
}
```

#### 删除代理
```http
DELETE /api/v1/agents/{id}
```

### 代理状态管理

#### 改变代理状态
```http
POST /api/v1/agents/{id}/status
Content-Type: application/json

{
  "status": "busy"
}
```

支持的状态：
- `idle`: 空闲
- `busy`: 忙碌
- `learning`: 学习中
- `sleeping`: 休眠
- `maintenance`: 维护中

### 工具管理

#### 为代理添加工具
```http
POST /api/v1/agents/{id}/tools
Content-Type: application/json

{
  "tool_id": "tool-uuid-here"
}
```

#### 获取代理的工具列表
```http
GET /api/v1/agents/{id}/tools
```

#### 移除代理的工具
```http
DELETE /api/v1/agents/{id}/tools/{tool_id}
```

### 记忆管理

#### 为代理添加记忆
```http
POST /api/v1/agents/{id}/memory
Content-Type: application/json

{
  "content": "重要的知识内容",
  "type": "learned",
  "importance": 0.8
}
```

#### 获取代理记忆
```http
GET /api/v1/agents/{id}/memory
```

### 代理执行

#### 执行对话
```http
POST /api/v1/agents/{id}/chat
Content-Type: application/json

{
  "message": "你好，请帮我分析这个问题",
  "context": {
    "session_id": "session-uuid"
  }
}
```

#### 执行任务
```http
POST /api/v1/agents/{id}/execute
Content-Type: application/json

{
  "task": "执行特定任务",
  "parameters": {
    "param1": "value1"
  }
}
```

## 数据模型

### 代理实体 (Agent)
```go
type Agent struct {
    ID               uuid.UUID
    Name             string
    Type             AgentType
    Status           AgentStatus
    Description      string
    SystemPrompt     string
    Config           map[string]interface{}
    Capabilities     []string
    OwnerID          uuid.UUID
    IsActive         bool
    LastActiveAt     time.Time
    LearningRate     float64
    MemoryCapacity   int
    ContextWindow    int
}
```

### 记忆实体 (Memory)
```go
type Memory struct {
    ID           uuid.UUID
    Content      string
    Type         MemoryType
    Importance   float64
    AccessCount  int
    LastAccessed time.Time
}
```

### 工具实体 (Tool)
```go
type Tool struct {
    ID          uuid.UUID
    Name        string
    Type        ToolType
    Description string
    Config      map[string]interface{}
    IsEnabled   bool
}
```

## 配置说明

### 代理配置项

| 配置项 | 类型 | 默认值 | 说明 |
|-------|------|--------|------|
| learning_rate | float64 | 0.1 | 学习速率 |
| memory_capacity | int | 1000 | 记忆容量 |
| context_window | int | 4096 | 上下文窗口大小 |
| max_tools | int | 10 | 最大工具数量 |

### 工具配置

每个工具都有特定的配置项，常见的包括：

- **计算器工具**
  - `precision`: 精度设置
  - `max_operations`: 最大操作数

- **文件工具**
  - `allowed_paths`: 允许访问的路径
  - `max_file_size`: 最大文件大小

## 扩展开发

### 自定义工具

1. 实现 `ToolExecutor` 接口：

```go
type CustomExecutor struct{}

func (e *CustomExecutor) Execute(ctx context.Context, input map[string]interface{}) (*ToolExecutionResult, error) {
    // 实现自定义逻辑
    return &ToolExecutionResult{
        Success: true,
        Output:  "执行结果",
    }, nil
}

func (e *CustomExecutor) GetName() string {
    return "custom_tool"
}
```

2. 注册工具执行器：

```go
// 在初始化时注册
app.ToolService.RegisterExecutor("custom_tool", &CustomExecutor{})
```

### 自定义代理类型

1. 扩展 `AgentType` 枚举
2. 实现对应的业务逻辑
3. 更新代理服务

## 监控和日志

### 健康检查
```http
GET /health
```

### 指标端点
```http
GET /metrics
```

### 日志配置

在配置文件中设置日志级别：

```yaml
logging:
  level: info
  format: json
  output: stdout
```

## 性能优化

### 连接池配置
```yaml
database:
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 1h
```

### 缓存配置
```yaml
cache:
  type: redis
  url: redis://localhost:6379
  ttl: 1h
```

## 故障排除

### 常见问题

1. **代理创建失败**
   - 检查数据库连接
   - 验证 owner_id 格式

2. **工具执行失败**
   - 检查工具配置
   - 验证权限设置

3. **记忆容量不足**
   - 调整 memory_capacity 配置
   - 清理旧的记忆数据

### 日志分析

关键日志关键字：
- `agent.created`: 代理创建
- `agent.status.changed`: 状态变更
- `tool.executed`: 工具执行
- `memory.added`: 记忆添加

## 开发指南

### 目录结构
```
agent/
├── cmd/              # 启动入口
├── internal/
│   ├── domain/       # 领域模型
│   ├── application/  # 应用服务
│   ├── infrastructure/ # 基础设施
│   └── interface/    # 接口层
└── go.mod
```

### 依赖注入

使用 Google Wire 进行依赖注入：

```go
//go:build wireinject
// +build wireinject

func InitializeAgentApp() (*AgentApp, func(), error) {
    wire.Build(
        // 依赖提供者
    )
    return &AgentApp{}, nil, nil
}
```

## 版本历史

- **v1.0.0**: 基础功能实现
- **v1.1.0**: 添加记忆管理
- **v1.2.0**: 支持工具集成
- **v1.3.0**: 增强学习能力

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 创建 Pull Request

## 许可证

MIT License
