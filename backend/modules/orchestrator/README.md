# Orchestrator 模块

## 概述

Orchestrator 模块是 Noah Loop 系统的工作流编排服务，负责设计、管理和执行复杂的工作流程。该模块提供可视化的工作流设计、智能调度、步骤编排、触发器管理等功能，是系统自动化和流程管理的核心组件。

## 主要功能

- 🔄 **工作流管理**：创建、编辑、版本控制工作流
- ⚡ **多触发器支持**：手动、定时、事件、Webhook、条件触发
- 🎯 **智能调度**：基于优先级和资源的智能任务调度
- 📊 **执行监控**：实时监控工作流执行状态和性能
- 🔀 **条件分支**：支持复杂的条件逻辑和并行执行
- 📈 **统计分析**：执行历史、成功率、性能分析
- 🛠️ **步骤模板**：可复用的步骤模板和组件库

## 核心概念

### 工作流 (Workflow)
工作流是一系列有序步骤的集合，定义了完整的业务流程：
- 包含多个执行步骤
- 支持条件分支和循环
- 具有触发器和调度规则
- 可以设置变量和参数

### 步骤 (Step)
步骤是工作流的基本执行单元：
- 具有特定的执行类型（API调用、脚本执行、代理任务等）
- 支持输入输出参数映射
- 可以设置错误处理和重试策略
- 支持条件执行和并行处理

### 触发器 (Trigger)
触发器定义工作流的启动条件：
- **手动触发**：用户手动启动
- **定时触发**：基于 Cron 表达式的定时执行
- **事件触发**：响应系统或外部事件
- **Webhook触发**：接收外部HTTP请求
- **条件触发**：基于特定条件自动触发

### 执行 (Execution)
执行是工作流的一次运行实例：
- 记录执行过程和状态
- 保存输入输出数据
- 支持暂停、恢复、取消操作
- 提供详细的执行日志

## 工作流状态

- **Draft**: 草稿状态，正在编辑中
- **Active**: 活跃状态，可以被触发执行
- **Paused**: 暂停状态，不会被触发
- **Completed**: 完成状态，已完成所有执行
- **Failed**: 失败状态，执行出现错误
- **Cancelled**: 取消状态，被用户取消

## 步骤类型

- **HTTP**: HTTP请求调用
- **Agent**: 代理任务执行
- **Script**: 脚本执行
- **Database**: 数据库操作
- **Email**: 邮件发送
- **Webhook**: Webhook调用
- **Condition**: 条件判断
- **Loop**: 循环执行
- **Parallel**: 并行执行

## 快速开始

### 1. 安装依赖

```bash
cd backend/modules/orchestrator
go mod download
```

### 2. 配置环境

在 `configs/config.yaml` 中配置编排服务：

```yaml
services:
  orchestrator:
    port: 8084
    database:
      url: "postgres://user:password@localhost/noah_loop?sslmode=disable"
    
    # 调度器配置
    scheduler:
      enabled: true
      workers: 10
      max_concurrent: 100
      retry_attempts: 3
      retry_delay: 30s
    
    # 执行器配置
    executor:
      timeout: 30m
      max_memory: "1GB"
      log_retention: "7d"
```

### 3. 启动服务

```bash
# 开发环境
go run cmd/main.go

# 生产环境
go build -o orchestrator cmd/main.go
./orchestrator
```

### 4. 验证服务

```bash
curl http://localhost:8084/health
```

## API 文档

### 工作流管理

#### 创建工作流
```http
POST /api/v1/workflows
Content-Type: application/json

{
  "name": "数据处理工作流",
  "description": "自动化数据处理和分析流程",
  "definition": {
    "version": "1.0",
    "variables": {
      "input_file": "",
      "output_dir": "/tmp/output"
    },
    "steps": [
      {
        "id": "step1",
        "name": "数据验证",
        "type": "script",
        "config": {
          "script": "validate_data.py",
          "timeout": "5m"
        }
      }
    ]
  },
  "tags": ["数据处理", "自动化"],
  "owner_id": "user-uuid"
}
```

#### 获取工作流列表
```http
GET /api/v1/workflows
GET /api/v1/workflows?owner_id=uuid
GET /api/v1/workflows?status=active
GET /api/v1/workflows?tags=数据处理
```

#### 获取特定工作流
```http
GET /api/v1/workflows/{id}
```

#### 更新工作流
```http
PUT /api/v1/workflows/{id}
Content-Type: application/json

{
  "name": "更新后的工作流",
  "description": "更新描述",
  "definition": {
    "version": "1.1",
    "steps": [...]
  }
}
```

#### 工作流状态管理
```http
POST /api/v1/workflows/{id}/activate
POST /api/v1/workflows/{id}/pause
POST /api/v1/workflows/{id}/resume
DELETE /api/v1/workflows/{id}
```

### 触发器管理

#### 添加触发器
```http
POST /api/v1/workflows/{workflow_id}/triggers
Content-Type: application/json

{
  "type": "schedule",
  "name": "每日数据处理",
  "config": {
    "cron": "0 2 * * *",
    "timezone": "Asia/Shanghai"
  },
  "is_enabled": true
}
```

#### 获取触发器列表
```http
GET /api/v1/workflows/{workflow_id}/triggers
```

#### 更新触发器
```http
PUT /api/v1/triggers/{id}
Content-Type: application/json

{
  "config": {
    "cron": "0 3 * * *"
  },
  "is_enabled": false
}
```

#### 删除触发器
```http
DELETE /api/v1/triggers/{id}
```

### 工作流执行

#### 手动执行工作流
```http
POST /api/v1/workflows/{id}/execute
Content-Type: application/json

{
  "input": {
    "input_file": "data.csv",
    "parameters": {
      "batch_size": 1000
    }
  },
  "priority": 5
}
```

#### 获取执行列表
```http
GET /api/v1/executions
GET /api/v1/executions?workflow_id=uuid
GET /api/v1/executions?status=running
```

#### 获取执行详情
```http
GET /api/v1/executions/{id}
```

#### 执行控制
```http
POST /api/v1/executions/{id}/pause
POST /api/v1/executions/{id}/resume
POST /api/v1/executions/{id}/cancel
POST /api/v1/executions/{id}/retry
```

#### 获取执行日志
```http
GET /api/v1/executions/{id}/logs
GET /api/v1/executions/{id}/logs?step_id=step1
```

### 步骤管理

#### 获取步骤执行详情
```http
GET /api/v1/step-executions/{id}
```

#### 获取步骤输出
```http
GET /api/v1/step-executions/{id}/output
```

#### 重试步骤
```http
POST /api/v1/step-executions/{id}/retry
```

## 数据模型

### 工作流实体 (Workflow)
```go
type Workflow struct {
    ID             uuid.UUID
    Name           string
    Description    string
    Status         WorkflowStatus
    Definition     map[string]interface{}
    Variables      map[string]interface{}
    Tags           []string
    OwnerID        uuid.UUID
    IsTemplate     bool
    ExecutionCount int
    LastExecuted   time.Time
    SuccessRate    float64
    Triggers       []*Trigger
    Steps          []*Step
}
```

### 步骤实体 (Step)
```go
type Step struct {
    ID          uuid.UUID
    WorkflowID  uuid.UUID
    Name        string
    Type        StepType
    Config      map[string]interface{}
    InputMapping map[string]string
    OutputMapping map[string]string
    Order       int
    IsParallel  bool
    Dependencies []uuid.UUID
    RetryPolicy *RetryPolicy
}
```

### 触发器实体 (Trigger)
```go
type Trigger struct {
    ID         uuid.UUID
    WorkflowID uuid.UUID
    Type       TriggerType
    Name       string
    Config     map[string]interface{}
    IsEnabled  bool
    LastTriggered time.Time
    TriggerCount  int
}
```

### 执行实体 (Execution)
```go
type Execution struct {
    ID           uuid.UUID
    WorkflowID   uuid.UUID
    Status       ExecutionStatus
    Input        map[string]interface{}
    Output       map[string]interface{}
    StartTime    time.Time
    EndTime      *time.Time
    Duration     time.Duration
    TriggerID    *uuid.UUID
    Priority     int
    ErrorMessage string
    StepExecutions []*StepExecution
}
```

## 工作流定义格式

### YAML 格式示例
```yaml
version: "1.0"
name: "数据处理工作流"
description: "自动化数据处理流程"

variables:
  input_file: ""
  output_dir: "/tmp/output"
  batch_size: 1000

steps:
  - id: "validate"
    name: "数据验证"
    type: "script"
    config:
      script: "validate.py"
      timeout: "5m"
    
  - id: "process"
    name: "数据处理"
    type: "agent"
    depends_on: ["validate"]
    config:
      agent_id: "data-processor"
      task: "process_data"
    input_mapping:
      file: "{{variables.input_file}}"
      batch_size: "{{variables.batch_size}}"
    
  - id: "notify"
    name: "发送通知"
    type: "email"
    depends_on: ["process"]
    config:
      to: "admin@example.com"
      subject: "数据处理完成"
      template: "process_complete"

triggers:
  - type: "schedule"
    name: "每日执行"
    config:
      cron: "0 2 * * *"
      timezone: "Asia/Shanghai"
  
  - type: "webhook"
    name: "手动触发"
    config:
      path: "/webhook/data-process"
      method: "POST"
```

### JSON 格式示例
```json
{
  "version": "1.0",
  "name": "API数据同步",
  "description": "定期同步外部API数据",
  "variables": {
    "api_endpoint": "https://api.example.com/data",
    "sync_interval": 3600
  },
  "steps": [
    {
      "id": "fetch_data",
      "name": "获取数据",
      "type": "http",
      "config": {
        "method": "GET",
        "url": "{{variables.api_endpoint}}",
        "headers": {
          "Authorization": "Bearer {{secrets.api_token}}"
        },
        "timeout": "30s"
      }
    },
    {
      "id": "transform_data",
      "name": "数据转换",
      "type": "script",
      "depends_on": ["fetch_data"],
      "config": {
        "script": "transform.js",
        "runtime": "node"
      },
      "input_mapping": {
        "raw_data": "{{steps.fetch_data.output.body}}"
      }
    },
    {
      "id": "save_data",
      "name": "保存数据",
      "type": "database",
      "depends_on": ["transform_data"],
      "config": {
        "operation": "insert",
        "table": "synced_data",
        "connection": "main_db"
      },
      "input_mapping": {
        "data": "{{steps.transform_data.output.result}}"
      }
    }
  ]
}
```

## 调度和执行

### 调度器配置
```yaml
scheduler:
  enabled: true
  workers: 10              # 工作协程数
  max_concurrent: 100      # 最大并发执行数
  poll_interval: 5s        # 轮询间隔
  priority_queue: true     # 启用优先级队列
  
  # 重试配置
  retry:
    max_attempts: 3
    initial_delay: 30s
    max_delay: 300s
    backoff_factor: 2.0
```

### 执行引擎
```go
type ExecutionEngine interface {
    Execute(ctx context.Context, workflow *Workflow, input map[string]interface{}) (*Execution, error)
    Pause(ctx context.Context, executionID uuid.UUID) error
    Resume(ctx context.Context, executionID uuid.UUID) error
    Cancel(ctx context.Context, executionID uuid.UUID) error
}

type DefaultExecutionEngine struct {
    stepExecutors map[StepType]StepExecutor
    scheduler     *Scheduler
}

func (e *DefaultExecutionEngine) Execute(ctx context.Context, workflow *Workflow, input map[string]interface{}) (*Execution, error) {
    execution := NewExecution(workflow.ID, input)
    
    // 构建执行图
    graph := e.buildExecutionGraph(workflow.Steps)
    
    // 按依赖顺序执行步骤
    for _, step := range graph.TopologicalSort() {
        stepExecution, err := e.executeStep(ctx, step, execution)
        if err != nil {
            execution.Status = ExecutionStatusFailed
            execution.ErrorMessage = err.Error()
            return execution, err
        }
        
        execution.StepExecutions = append(execution.StepExecutions, stepExecution)
    }
    
    execution.Status = ExecutionStatusCompleted
    return execution, nil
}
```

### 步骤执行器
```go
type StepExecutor interface {
    GetType() StepType
    Execute(ctx context.Context, step *Step, input map[string]interface{}) (*StepExecutionResult, error)
}

// HTTP步骤执行器
type HTTPStepExecutor struct {
    client *http.Client
}

func (e *HTTPStepExecutor) Execute(ctx context.Context, step *Step, input map[string]interface{}) (*StepExecutionResult, error) {
    url := e.resolveTemplate(step.Config["url"].(string), input)
    method := step.Config["method"].(string)
    
    req, err := http.NewRequestWithContext(ctx, method, url, nil)
    if err != nil {
        return nil, err
    }
    
    // 设置请求头
    if headers, ok := step.Config["headers"].(map[string]interface{}); ok {
        for key, value := range headers {
            req.Header.Set(key, e.resolveTemplate(value.(string), input))
        }
    }
    
    resp, err := e.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    
    return &StepExecutionResult{
        Success: resp.StatusCode < 400,
        Output: map[string]interface{}{
            "status_code": resp.StatusCode,
            "headers": resp.Header,
            "body": string(body),
        },
    }, nil
}
```

## 监控和指标

### 工作流指标

- **执行统计**
  - 总执行次数
  - 成功/失败次数
  - 平均执行时间
  - 成功率趋势

- **性能指标**
  - 执行延迟分布
  - 步骤执行时间
  - 资源使用情况
  - 队列长度

- **错误监控**
  - 错误类型分布
  - 失败步骤统计
  - 重试成功率
  - 错误趋势分析

### Prometheus 指标
```go
var (
    workflowExecutionsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "orchestrator_executions_total",
            Help: "Total number of workflow executions",
        },
        []string{"workflow_id", "status"},
    )
    
    executionDurationHistogram = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "orchestrator_execution_duration_seconds",
            Help: "Workflow execution duration",
            Buckets: prometheus.ExponentialBuckets(1, 2, 10),
        },
        []string{"workflow_id"},
    )
    
    activeExecutions = prometheus.NewGauge(
        prometheus.GaugeOpts{
            Name: "orchestrator_active_executions",
            Help: "Number of currently active executions",
        },
    )
)
```

### 健康检查
```http
GET /health
```

返回示例：
```json
{
  "status": "healthy",
  "checks": {
    "database": "ok",
    "scheduler": "ok",
    "executor": "ok"
  },
  "stats": {
    "active_executions": 15,
    "queued_executions": 5,
    "total_workflows": 120,
    "active_workflows": 85
  }
}
```

## 扩展开发

### 自定义步骤类型
```go
// 自定义邮件步骤执行器
type EmailStepExecutor struct {
    smtpConfig SMTPConfig
}

func (e *EmailStepExecutor) GetType() StepType {
    return "email"
}

func (e *EmailStepExecutor) Execute(ctx context.Context, step *Step, input map[string]interface{}) (*StepExecutionResult, error) {
    to := step.Config["to"].(string)
    subject := e.resolveTemplate(step.Config["subject"].(string), input)
    body := e.resolveTemplate(step.Config["body"].(string), input)
    
    err := e.sendEmail(to, subject, body)
    if err != nil {
        return &StepExecutionResult{
            Success: false,
            Error:   err.Error(),
        }, err
    }
    
    return &StepExecutionResult{
        Success: true,
        Output: map[string]interface{}{
            "sent_at": time.Now(),
            "to": to,
        },
    }, nil
}

// 注册执行器
app.OrchestratorService.RegisterStepExecutor(&EmailStepExecutor{
    smtpConfig: loadSMTPConfig(),
})
```

### 自定义触发器
```go
type CustomTrigger struct {
    config     map[string]interface{}
    workflow   *Workflow
    onTrigger  func(*Workflow, map[string]interface{})
}

func (t *CustomTrigger) GetType() TriggerType {
    return "custom"
}

func (t *CustomTrigger) Start(ctx context.Context) error {
    // 启动触发器逻辑
    go t.watchForTriggerConditions(ctx)
    return nil
}

func (t *CustomTrigger) Stop() error {
    // 停止触发器逻辑
    return nil
}
```

### 工作流插件
```go
type WorkflowPlugin interface {
    OnWorkflowStart(ctx context.Context, execution *Execution) error
    OnStepStart(ctx context.Context, stepExecution *StepExecution) error
    OnStepComplete(ctx context.Context, stepExecution *StepExecution) error
    OnWorkflowComplete(ctx context.Context, execution *Execution) error
}

type LoggingPlugin struct {
    logger Logger
}

func (p *LoggingPlugin) OnWorkflowStart(ctx context.Context, execution *Execution) error {
    p.logger.Info("Workflow started",
        "execution_id", execution.ID,
        "workflow_id", execution.WorkflowID,
        "input", execution.Input)
    return nil
}
```

## 集成示例

### 与代理模块集成
```yaml
steps:
  - id: "analysis"
    name: "数据分析"
    type: "agent"
    config:
      agent_id: "data-analyst"
      task: "analyze_sales_data"
      timeout: "10m"
    input_mapping:
      data_file: "{{variables.input_file}}"
      analysis_type: "monthly"
```

### 与LLM模块集成
```yaml
steps:
  - id: "generate_summary"
    name: "生成报告摘要"
    type: "llm"
    config:
      model: "gpt-4"
      prompt: "请为以下数据生成摘要: {{steps.analysis.output.result}}"
      max_tokens: 500
```

### 与MCP模块集成
```yaml
steps:
  - id: "create_session"
    name: "创建对话会话"
    type: "mcp"
    config:
      action: "create_session"
      user_id: "{{input.user_id}}"
      agent_id: "{{variables.assistant_id}}"
```

## 故障排除

### 常见问题

1. **工作流执行失败**
   - 检查步骤配置和依赖关系
   - 验证输入参数格式
   - 查看执行日志

2. **触发器不工作**
   - 检查触发器配置
   - 验证Cron表达式
   - 确认触发器状态

3. **执行超时**
   - 检查步骤超时配置
   - 优化步骤执行逻辑
   - 调整资源限制

4. **并发执行问题**
   - 检查并发限制配置
   - 验证资源依赖
   - 调整调度策略

### 诊断工具

```bash
# 检查工作流状态
curl http://localhost:8084/api/v1/workflows/{id}/status

# 获取执行日志
curl http://localhost:8084/api/v1/executions/{id}/logs

# 检查调度器状态
curl http://localhost:8084/api/v1/scheduler/status

# 查看性能指标
curl http://localhost:8084/metrics
```

## 最佳实践

### 工作流设计原则

1. **单一职责**：每个步骤只做一件事
2. **幂等性**：步骤可以重复执行而不产生副作用
3. **错误处理**：为每个步骤设置合适的错误处理策略
4. **资源管理**：合理设置超时和资源限制
5. **监控友好**：添加必要的日志和指标

### 性能优化

1. **并行执行**：合理使用并行步骤
2. **缓存策略**：缓存中间结果
3. **资源池化**：复用连接和资源
4. **分片处理**：大数据集分片处理
5. **异步执行**：使用异步I/O

### 安全考虑

1. **权限控制**：基于角色的访问控制
2. **密钥管理**：安全存储敏感信息
3. **输入验证**：验证所有输入参数
4. **执行隔离**：隔离不同工作流的执行环境
5. **审计日志**：记录所有操作日志

## 版本历史

- **v1.0.0**: 基础工作流管理功能
- **v1.1.0**: 添加触发器支持
- **v1.2.0**: 实现智能调度
- **v1.3.0**: 增强执行监控
- **v1.4.0**: 支持并行执行
- **v1.5.0**: 添加工作流模板
- **v1.6.0**: 集成其他模块

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 编写测试用例
4. 更新文档
5. 提交更改
6. 创建 Pull Request

## 许可证

MIT License
