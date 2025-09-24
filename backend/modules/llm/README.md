# LLM 模块

## 概述

LLM 模块是 Noah Loop 系统的大语言模型服务，提供统一的大语言模型接口和管理功能。该模块支持多种模型提供商，包括 OpenAI、Anthropic、本地模型等，为系统其他模块提供强大的语言处理能力。

## 主要功能

- 🌐 **多提供商支持**：OpenAI、Anthropic、本地模型、自定义提供商
- 🎯 **统一接口**：标准化的模型调用接口
- 📊 **模型管理**：模型注册、配置、监控
- 💰 **成本控制**：使用量统计和成本计算
- 🔄 **智能路由**：根据任务类型选择最适合的模型
- 📈 **性能监控**：响应时间、成功率等指标

## 支持的模型类型

### 1. 聊天模型 (Chat)
- 对话式交互
- 支持系统提示和多轮对话
- 适用于问答、对话场景

### 2. 补全模型 (Completion)
- 文本补全和生成
- 适用于创作、代码生成

### 3. 嵌入模型 (Embedding)
- 文本向量化
- 适用于搜索、相似度计算

### 4. 图像模型 (Image)
- 图像生成和理解
- 支持多模态交互

### 5. 音频模型 (Audio)
- 语音转文字、文字转语音
- 支持语音交互

## 支持的提供商

### OpenAI
- GPT-3.5/4系列
- DALL-E
- Whisper
- Text-embedding-ada-002

### Anthropic
- Claude系列模型

### 本地模型
- Ollama
- LLaMA
- 自部署模型

### 自定义提供商
- 支持扩展集成

## 快速开始

### 1. 安装依赖

```bash
cd backend/modules/llm
go mod download
```

### 2. 配置环境

在 `configs/config.yaml` 中配置 LLM 服务：

```yaml
services:
  llm:
    port: 8082
    database:
      url: "postgres://user:password@localhost/noah_loop?sslmode=disable"

# 提供商配置
providers:
  openai:
    api_key: "${OPENAI_API_KEY}"
    base_url: "https://api.openai.com/v1"
  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"
  local:
    endpoint: "http://localhost:11434"
```

### 3. 设置环境变量

```bash
export OPENAI_API_KEY="your-openai-api-key"
export ANTHROPIC_API_KEY="your-anthropic-api-key"
```

### 4. 启动服务

```bash
# 开发环境
go run cmd/main.go

# 生产环境
go build -o llm cmd/main.go
./llm
```

### 5. 验证服务

```bash
curl http://localhost:8082/health
```

## API 文档

### 模型管理

#### 注册模型
```http
POST /api/v1/models
Content-Type: application/json

{
  "name": "gpt-3.5-turbo",
  "provider": "openai",
  "type": "chat",
  "version": "0613",
  "description": "GPT-3.5 Turbo模型",
  "max_tokens": 4096,
  "price_per_k": 0.002,
  "config": {
    "temperature": 0.7,
    "top_p": 1.0
  },
  "capabilities": ["chat", "function_calling"]
}
```

#### 获取模型列表
```http
GET /api/v1/models
```

#### 获取特定模型
```http
GET /api/v1/models/{id}
```

#### 更新模型配置
```http
PUT /api/v1/models/{id}
Content-Type: application/json

{
  "config": {
    "temperature": 0.5,
    "max_tokens": 2048
  }
}
```

#### 激活/停用模型
```http
POST /api/v1/models/{id}/activate
POST /api/v1/models/{id}/deactivate
```

### 模型调用

#### 聊天补全
```http
POST /api/v1/chat/completions
Content-Type: application/json

{
  "model": "gpt-3.5-turbo",
  "messages": [
    {
      "role": "system",
      "content": "你是一个有用的AI助手"
    },
    {
      "role": "user", 
      "content": "请介绍一下人工智能"
    }
  ],
  "temperature": 0.7,
  "max_tokens": 1000
}
```

#### 文本补全
```http
POST /api/v1/completions
Content-Type: application/json

{
  "model": "text-davinci-003",
  "prompt": "人工智能的发展历史",
  "max_tokens": 500,
  "temperature": 0.7
}
```

#### 文本嵌入
```http
POST /api/v1/embeddings
Content-Type: application/json

{
  "model": "text-embedding-ada-002",
  "input": [
    "人工智能技术",
    "机器学习算法"
  ]
}
```

#### 图像生成
```http
POST /api/v1/images/generations
Content-Type: application/json

{
  "model": "dall-e-3",
  "prompt": "一只可爱的机器猫在太空中飞行",
  "size": "1024x1024",
  "quality": "standard",
  "n": 1
}
```

### 请求管理

#### 获取请求历史
```http
GET /api/v1/requests
```

#### 获取特定请求详情
```http
GET /api/v1/requests/{id}
```

#### 请求统计
```http
GET /api/v1/requests/stats
```

## 数据模型

### 模型实体 (Model)
```go
type Model struct {
    ID           uuid.UUID
    Name         string
    Provider     ModelProvider
    Type         ModelType
    Version      string
    Description  string
    Config       map[string]interface{}
    Capabilities []string
    MaxTokens    int
    PricePerK    float64
    IsActive     bool
}
```

### 请求实体 (Request)
```go
type Request struct {
    ID           uuid.UUID
    ModelID      uuid.UUID
    UserID       uuid.UUID
    Type         RequestType
    Input        map[string]interface{}
    Output       map[string]interface{}
    TokensUsed   int
    Cost         float64
    Duration     time.Duration
    Status       RequestStatus
    ErrorMessage string
}
```

## 提供商配置

### OpenAI 提供商
```yaml
providers:
  openai:
    api_key: "${OPENAI_API_KEY}"
    organization: "${OPENAI_ORG_ID}"  # 可选
    base_url: "https://api.openai.com/v1"
    timeout: 30s
    retry_attempts: 3
```

### Anthropic 提供商
```yaml
providers:
  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"
    base_url: "https://api.anthropic.com"
    timeout: 30s
```

### 本地模型提供商
```yaml
providers:
  local:
    endpoint: "http://localhost:11434"
    timeout: 60s
    models:
      - name: "llama2"
        type: "chat"
        max_tokens: 4096
```

## 模型路由

### 智能路由配置
```yaml
routing:
  strategies:
    - name: "cost_optimized"
      rules:
        - if: "tokens < 1000"
          model: "gpt-3.5-turbo"
        - if: "tokens >= 1000"
          model: "claude-2"
    
    - name: "performance_optimized"
      rules:
        - if: "task_type == 'code'"
          model: "gpt-4"
        - if: "task_type == 'creative'"
          model: "claude-2"
```

### 自定义路由策略
```go
type RouterStrategy interface {
    SelectModel(ctx context.Context, request *ModelRequest) (*Model, error)
}

type CostOptimizedRouter struct {
    models []*Model
}

func (r *CostOptimizedRouter) SelectModel(ctx context.Context, request *ModelRequest) (*Model, error) {
    // 实现成本优化逻辑
    return cheapestSuitableModel, nil
}
```

## 监控和指标

### 关键指标

- **请求量**：每分钟/小时/天的请求数
- **响应时间**：平均响应时间、P95、P99
- **成功率**：请求成功率
- **Token 使用量**：输入/输出 Token 统计
- **成本**：按模型、用户的成本统计
- **错误率**：各类错误的发生率

### 指标端点
```http
GET /metrics
```

### 监控仪表板

推荐使用 Grafana + Prometheus：

```yaml
# docker-compose.yml
services:
  grafana:
    image: grafana/grafana
    ports:
      - "3000:3000"
  
  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
```

## 成本管理

### 成本统计
```http
GET /api/v1/costs/summary
GET /api/v1/costs/by-model
GET /api/v1/costs/by-user
```

### 预算控制
```yaml
budget:
  daily_limit: 100.0    # 每日预算限制（美元）
  user_limits:
    default: 10.0       # 默认用户限制
    premium: 50.0       # 高级用户限制
  
alerts:
  - threshold: 0.8      # 80% 预算时警告
    action: "notify"
  - threshold: 1.0      # 100% 预算时限制
    action: "block"
```

## 扩展开发

### 自定义提供商

1. 实现 `ModelProvider` 接口：

```go
type CustomProvider struct {
    apiKey  string
    baseURL string
}

func (p *CustomProvider) GetName() string {
    return "custom"
}

func (p *CustomProvider) CreateChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
    // 实现聊天补全逻辑
    return response, nil
}

func (p *CustomProvider) CreateCompletion(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
    // 实现文本补全逻辑
    return response, nil
}
```

2. 注册提供商：

```go
// 在初始化时注册
app.LLMService.RegisterProvider("custom", &CustomProvider{
    apiKey:  os.Getenv("CUSTOM_API_KEY"),
    baseURL: "https://api.custom.com",
})
```

### 中间件支持

```go
type Middleware func(ModelHandler) ModelHandler

// 日志中间件
func LoggingMiddleware(next ModelHandler) ModelHandler {
    return func(ctx context.Context, req *ModelRequest) (*ModelResponse, error) {
        start := time.Now()
        resp, err := next(ctx, req)
        duration := time.Since(start)
        
        log.Printf("Model request: %s, Duration: %v, Error: %v", 
            req.Model, duration, err)
        
        return resp, err
    }
}

// 限流中间件
func RateLimitMiddleware(limiter *rate.Limiter) Middleware {
    return func(next ModelHandler) ModelHandler {
        return func(ctx context.Context, req *ModelRequest) (*ModelResponse, error) {
            if !limiter.Allow() {
                return nil, ErrRateLimit
            }
            return next(ctx, req)
        }
    }
}
```

## 缓存策略

### Redis 缓存配置
```yaml
cache:
  redis:
    addr: "localhost:6379"
    password: ""
    db: 0
  
  policies:
    embeddings:
      ttl: 24h
      max_size: 10000
    
    completions:
      ttl: 1h
      max_size: 1000
      enable_compression: true
```

### 缓存实现
```go
type CacheConfig struct {
    TTL         time.Duration
    MaxSize     int
    Compression bool
}

func (s *LLMService) GetCachedEmbedding(ctx context.Context, text string) ([]float64, bool) {
    key := fmt.Sprintf("embedding:%s", hash(text))
    return s.cache.Get(key)
}
```

## 故障排除

### 常见问题

1. **API 密钥无效**
   - 检查环境变量设置
   - 验证密钥格式和权限

2. **请求超时**
   - 调整 timeout 配置
   - 检查网络连接

3. **配额超限**
   - 查看提供商配额状态
   - 实施请求限流

4. **模型不可用**
   - 检查模型状态
   - 验证提供商配置

### 调试工具

```bash
# 检查模型状态
curl http://localhost:8082/api/v1/models

# 测试模型调用
curl -X POST http://localhost:8082/api/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"Hello"}]}'

# 查看请求日志
tail -f logs/llm.log
```

## 性能优化

### 连接池优化
```yaml
http_client:
  max_idle_conns: 100
  max_idle_conns_per_host: 10
  idle_conn_timeout: 90s
  timeout: 30s
```

### 批量处理
```go
// 批量嵌入处理
func (s *LLMService) BatchEmbedding(ctx context.Context, texts []string, batchSize int) ([][]float64, error) {
    var results [][]float64
    
    for i := 0; i < len(texts); i += batchSize {
        end := i + batchSize
        if end > len(texts) {
            end = len(texts)
        }
        
        batch := texts[i:end]
        embeddings, err := s.CreateEmbeddings(ctx, batch)
        if err != nil {
            return nil, err
        }
        
        results = append(results, embeddings...)
    }
    
    return results, nil
}
```

## 安全考虑

### API 密钥管理
- 使用环境变量存储敏感信息
- 定期轮换 API 密钥
- 实施最小权限原则

### 输入验证
```go
func validateInput(input string) error {
    if len(input) > maxInputLength {
        return ErrInputTooLong
    }
    
    if containsSensitiveData(input) {
        return ErrSensitiveContent
    }
    
    return nil
}
```

### 输出过滤
```go
func filterOutput(output string) string {
    // 过滤敏感信息
    return sanitize(output)
}
```

## 版本历史

- **v1.0.0**: 基础功能实现
- **v1.1.0**: 添加 Anthropic 支持
- **v1.2.0**: 实现智能路由
- **v1.3.0**: 增加成本控制
- **v1.4.0**: 支持本地模型

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 添加测试用例
4. 提交更改
5. 创建 Pull Request

## 许可证

MIT License
