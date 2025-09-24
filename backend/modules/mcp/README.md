# MCP 模块

## 概述

MCP (Model Context Protocol) 模块是 Noah Loop 系统的会话和上下文管理服务，负责管理用户与代理之间的会话状态、上下文信息和对话历史。该模块提供智能的上下文管理、会话生命周期控制和内存优化功能。

## 主要功能

- 💬 **会话管理**：用户与代理的会话生命周期管理
- 🧠 **上下文管理**：智能的上下文存储、检索和优化
- 🔄 **自动清理**：过期会话和上下文的自动清理
- 📊 **相关性计算**：基于内容的上下文相关性分析
- 🎯 **内存优化**：智能的上下文大小管理和优先级排序
- 📈 **会话分析**：会话活动度和使用模式分析

## 核心概念

### 会话 (Session)
- 用户与特定代理的对话会话
- 包含会话元数据、状态和配置信息
- 支持会话的创建、激活、暂停、归档和过期

### 上下文 (Context)
- 会话中的具体对话内容和上下文信息
- 包含消息、文件、图片等多种类型
- 支持优先级设置和相关性计算

## 会话状态

- **Active**: 活跃状态，正在进行对话
- **Idle**: 空闲状态，暂时没有活动
- **Archived**: 归档状态，长期保存
- **Expired**: 过期状态，等待清理

## 上下文类型

- **Message**: 对话消息
- **File**: 文件内容
- **Image**: 图片信息
- **System**: 系统消息
- **Function**: 函数调用结果

## 快速开始

### 1. 安装依赖

```bash
cd backend/modules/mcp
go mod download
```

### 2. 配置环境

在 `configs/config.yaml` 中配置 MCP 服务：

```yaml
services:
  mcp:
    port: 8083
    database:
      url: "postgres://user:password@localhost/noah_loop?sslmode=disable"
    
    # 会话配置
    session:
      default_ttl: 24h        # 默认会话过期时间
      max_context_size: 8192  # 最大上下文大小
      cleanup_interval: 1h    # 清理任务间隔
      idle_threshold: 2h      # 空闲阈值
    
    # 上下文管理
    context:
      compression_enabled: true
      max_history: 1000
      relevance_threshold: 0.3
```

### 3. 启动服务

```bash
# 开发环境
go run cmd/main.go

# 生产环境
go build -o mcp cmd/main.go
./mcp
```

### 4. 验证服务

```bash
curl http://localhost:8083/health
```

## API 文档

### 会话管理

#### 创建会话
```http
POST /api/v1/sessions
Content-Type: application/json

{
  "user_id": "user-uuid",
  "agent_id": "agent-uuid",
  "title": "智能助手对话",
  "description": "关于项目开发的讨论",
  "max_context_size": 4096,
  "metadata": {
    "project": "noah-loop",
    "priority": "high"
  }
}
```

#### 获取会话列表
```http
GET /api/v1/sessions
GET /api/v1/sessions?user_id=uuid
GET /api/v1/sessions?agent_id=uuid
GET /api/v1/sessions?status=active
```

#### 获取特定会话
```http
GET /api/v1/sessions/{id}
```

#### 更新会话
```http
PUT /api/v1/sessions/{id}
Content-Type: application/json

{
  "title": "更新后的标题",
  "description": "更新后的描述"
}
```

#### 会话状态管理
```http
POST /api/v1/sessions/{id}/activate
POST /api/v1/sessions/{id}/pause
POST /api/v1/sessions/{id}/archive
POST /api/v1/sessions/{id}/extend
```

#### 删除会话
```http
DELETE /api/v1/sessions/{id}
```

### 上下文管理

#### 添加上下文
```http
POST /api/v1/sessions/{session_id}/contexts
Content-Type: application/json

{
  "type": "message",
  "content": "用户的问题或回复",
  "role": "user",
  "priority": 5,
  "metadata": {
    "timestamp": "2024-01-01T12:00:00Z",
    "source": "web"
  }
}
```

#### 获取会话上下文
```http
GET /api/v1/sessions/{session_id}/contexts
GET /api/v1/sessions/{session_id}/contexts?type=message
GET /api/v1/sessions/{session_id}/contexts?limit=50
```

#### 获取相关上下文
```http
POST /api/v1/sessions/{session_id}/contexts/search
Content-Type: application/json

{
  "query": "搜索关键词",
  "limit": 10,
  "threshold": 0.5
}
```

#### 更新上下文
```http
PUT /api/v1/contexts/{id}
Content-Type: application/json

{
  "priority": 8,
  "metadata": {
    "importance": "high"
  }
}
```

#### 删除上下文
```http
DELETE /api/v1/contexts/{id}
```

### 会话统计

#### 获取会话统计信息
```http
GET /api/v1/sessions/{id}/stats
```

返回示例：
```json
{
  "session_id": "uuid",
  "message_count": 156,
  "total_tokens": 12480,
  "current_size": 8192,
  "context_types": {
    "message": 120,
    "file": 15,
    "image": 8,
    "system": 13
  },
  "activity_pattern": {
    "last_activity": "2024-01-01T15:30:00Z",
    "avg_daily_messages": 25,
    "peak_hours": [9, 14, 20]
  }
}
```

### 批量操作

#### 批量清理
```http
POST /api/v1/sessions/batch/cleanup
Content-Type: application/json

{
  "criteria": {
    "expired": true,
    "idle_threshold": "48h",
    "min_activity": 0
  }
}
```

#### 批量导出
```http
POST /api/v1/sessions/batch/export
Content-Type: application/json

{
  "session_ids": ["uuid1", "uuid2"],
  "format": "json",
  "include_contexts": true
}
```

## 数据模型

### 会话实体 (Session)
```go
type Session struct {
    ID             uuid.UUID
    UserID         uuid.UUID
    AgentID        uuid.UUID
    Status         SessionStatus
    Title          string
    Description    string
    Metadata       map[string]interface{}
    MaxContextSize int
    CurrentSize    int
    MessageCount   int
    LastActivity   time.Time
    ExpiresAt      *time.Time
    Contexts       []*Context
}
```

### 上下文实体 (Context)
```go
type Context struct {
    ID         uuid.UUID
    SessionID  uuid.UUID
    Type       ContextType
    Content    string
    Role       string
    Priority   int
    TokenCount int
    Embedding  []float64
    Metadata   map[string]interface{}
    CreatedAt  time.Time
    AccessedAt time.Time
}
```

## 上下文管理策略

### 大小管理
当会话上下文超过最大限制时，系统会：

1. **优先级排序**：按优先级和相关性排序
2. **智能裁剪**：保留重要和最近的上下文
3. **压缩存储**：对历史上下文进行压缩
4. **分片存储**：将长上下文分片处理

```yaml
context_management:
  strategies:
    - name: "size_based"
      max_size: 8192
      keep_ratio: 0.7  # 保留70%的内容
      priority_weight: 0.6
      recency_weight: 0.4
    
    - name: "relevance_based"
      min_relevance: 0.3
      max_contexts: 100
      compression_ratio: 0.5
```

### 相关性计算
```go
type RelevanceCalculator interface {
    CalculateRelevance(context *Context, query string) float64
}

// 基于嵌入的相关性计算
type EmbeddingRelevanceCalculator struct {
    embeddingService EmbeddingService
}

func (c *EmbeddingRelevanceCalculator) CalculateRelevance(context *Context, query string) float64 {
    if context.Embedding == nil {
        // 计算嵌入向量
        embedding := c.embeddingService.GetEmbedding(context.Content)
        context.Embedding = embedding
    }
    
    queryEmbedding := c.embeddingService.GetEmbedding(query)
    return cosineSimilarity(context.Embedding, queryEmbedding)
}
```

## 会话生命周期管理

### 自动清理任务

系统每小时执行以下清理任务：

1. **过期会话清理**：删除已过期的会话
2. **空闲会话管理**：将长时间无活动的会话设为空闲
3. **上下文压缩**：压缩低优先级的历史上下文
4. **垃圾回收**：清理孤立的上下文记录

```go
func (s *MCPService) CleanupTasks(ctx context.Context) error {
    // 1. 清理过期会话
    if err := s.CleanupExpiredSessions(ctx); err != nil {
        return err
    }
    
    // 2. 管理空闲会话
    if err := s.ManageIdleSessions(ctx, 2*time.Hour); err != nil {
        return err
    }
    
    // 3. 压缩历史上下文
    if err := s.CompressHistoryContexts(ctx); err != nil {
        return err
    }
    
    return nil
}
```

### 会话归档
```go
type ArchiveService interface {
    ArchiveSession(ctx context.Context, sessionID uuid.UUID) error
    RestoreSession(ctx context.Context, sessionID uuid.UUID) error
    ListArchivedSessions(ctx context.Context, userID uuid.UUID) ([]*Session, error)
}
```

## 性能优化

### 数据库索引
```sql
-- 会话索引
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_agent_id ON sessions(agent_id);
CREATE INDEX idx_sessions_status ON sessions(status);
CREATE INDEX idx_sessions_last_activity ON sessions(last_activity);

-- 上下文索引
CREATE INDEX idx_contexts_session_id ON contexts(session_id);
CREATE INDEX idx_contexts_type ON contexts(type);
CREATE INDEX idx_contexts_priority ON contexts(priority DESC);
CREATE INDEX idx_contexts_created_at ON contexts(created_at DESC);
```

### 缓存策略
```yaml
cache:
  sessions:
    ttl: 1h
    max_size: 10000
    eviction_policy: "lru"
  
  contexts:
    ttl: 30m
    max_size: 50000
    compression: true
  
  embeddings:
    ttl: 24h
    max_size: 100000
```

### 分页和限制
```go
type PaginationConfig struct {
    DefaultLimit int `yaml:"default_limit"`
    MaxLimit     int `yaml:"max_limit"`
}

func (s *MCPService) GetSessions(ctx context.Context, req *GetSessionsRequest) (*GetSessionsResponse, error) {
    limit := req.Limit
    if limit == 0 {
        limit = s.config.DefaultLimit
    }
    if limit > s.config.MaxLimit {
        limit = s.config.MaxLimit
    }
    
    offset := req.Offset
    if offset < 0 {
        offset = 0
    }
    
    // 执行分页查询
    sessions, total, err := s.repository.FindPaginated(ctx, limit, offset, req.Filters)
    if err != nil {
        return nil, err
    }
    
    return &GetSessionsResponse{
        Sessions: sessions,
        Total:    total,
        Limit:    limit,
        Offset:   offset,
    }, nil
}
```

## 监控和指标

### 关键指标

- **会话指标**
  - 活跃会话数
  - 新建会话数
  - 平均会话持续时间
  - 会话转换率（活跃→归档）

- **上下文指标**
  - 平均上下文大小
  - 上下文压缩率
  - 相关性查询性能
  - 清理效率

- **性能指标**
  - API 响应时间
  - 数据库查询延迟
  - 缓存命中率
  - 内存使用量

### Prometheus 指标
```go
var (
    sessionsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "mcp_sessions_total",
            Help: "Total number of sessions created",
        },
        []string{"status"},
    )
    
    contextSizeHistogram = prometheus.NewHistogram(
        prometheus.HistogramOpts{
            Name: "mcp_context_size_bytes",
            Help: "Distribution of context sizes",
            Buckets: prometheus.ExponentialBuckets(100, 2, 10),
        },
    )
    
    cleanupDuration = prometheus.NewHistogram(
        prometheus.HistogramOpts{
            Name: "mcp_cleanup_duration_seconds",
            Help: "Time spent on cleanup tasks",
        },
    )
)
```

## 扩展开发

### 自定义上下文类型
```go
// 注册自定义上下文类型
type CustomContextHandler struct{}

func (h *CustomContextHandler) GetType() ContextType {
    return "custom"
}

func (h *CustomContextHandler) ProcessContent(content string) (string, error) {
    // 自定义处理逻辑
    return processedContent, nil
}

func (h *CustomContextHandler) CalculateTokens(content string) int {
    // 自定义Token计算
    return tokenCount
}

// 注册处理器
app.MCPService.RegisterContextHandler(&CustomContextHandler{})
```

### 自定义相关性计算
```go
type BusinessRelevanceCalculator struct {
    rules []RelevanceRule
}

type RelevanceRule struct {
    Pattern    string
    Weight     float64
    BoostTerms []string
}

func (c *BusinessRelevanceCalculator) CalculateRelevance(context *Context, query string) float64 {
    score := 0.0
    
    for _, rule := range c.rules {
        if matched, _ := regexp.MatchString(rule.Pattern, context.Content); matched {
            score += rule.Weight
            
            for _, term := range rule.BoostTerms {
                if strings.Contains(strings.ToLower(query), strings.ToLower(term)) {
                    score += 0.1
                }
            }
        }
    }
    
    return math.Min(score, 1.0)
}
```

### 会话插件系统
```go
type SessionPlugin interface {
    OnSessionCreated(ctx context.Context, session *Session) error
    OnSessionUpdated(ctx context.Context, session *Session) error
    OnSessionClosed(ctx context.Context, session *Session) error
}

type AuditPlugin struct {
    logger Logger
}

func (p *AuditPlugin) OnSessionCreated(ctx context.Context, session *Session) error {
    p.logger.Info("Session created", 
        "session_id", session.ID,
        "user_id", session.UserID,
        "agent_id", session.AgentID)
    return nil
}
```

## 故障排除

### 常见问题

1. **会话创建失败**
   - 检查用户ID和代理ID的有效性
   - 验证数据库连接
   - 确认权限设置

2. **上下文大小超限**
   - 检查 max_context_size 配置
   - 启用自动清理功能
   - 调整上下文管理策略

3. **相关性搜索慢**
   - 检查嵌入向量索引
   - 优化查询条件
   - 调整缓存策略

4. **内存使用过高**
   - 检查上下文缓存配置
   - 启用压缩功能
   - 调整清理频率

### 诊断工具

```bash
# 检查会话状态
curl http://localhost:8083/api/v1/sessions/stats

# 检查清理任务状态
curl http://localhost:8083/api/v1/cleanup/status

# 查看性能指标
curl http://localhost:8083/metrics

# 测试相关性搜索
curl -X POST http://localhost:8083/api/v1/sessions/{id}/contexts/search \
  -H "Content-Type: application/json" \
  -d '{"query":"测试查询","limit":5}'
```

## 配置参考

### 完整配置示例
```yaml
services:
  mcp:
    port: 8083
    
    # 数据库配置
    database:
      url: "postgres://user:password@localhost/noah_loop?sslmode=disable"
      max_open_conns: 100
      max_idle_conns: 10
      conn_max_lifetime: 1h
    
    # 会话配置
    session:
      default_ttl: 24h
      max_context_size: 8192
      cleanup_interval: 1h
      idle_threshold: 2h
      batch_size: 100
    
    # 上下文配置
    context:
      compression_enabled: true
      compression_threshold: 1000
      max_history: 1000
      relevance_threshold: 0.3
      embedding_model: "text-embedding-ada-002"
    
    # 缓存配置
    cache:
      type: "redis"
      url: "redis://localhost:6379/0"
      default_ttl: 1h
      max_memory: "256mb"
    
    # 监控配置
    metrics:
      enabled: true
      port: 9090
      path: "/metrics"
```

## 版本历史

- **v1.0.0**: 基础会话管理功能
- **v1.1.0**: 添加上下文管理
- **v1.2.0**: 实现相关性搜索
- **v1.3.0**: 增强清理机制
- **v1.4.0**: 支持会话归档
- **v1.5.0**: 添加性能监控

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 编写测试用例
4. 提交更改
5. 创建 Pull Request

## 许可证

MIT License
