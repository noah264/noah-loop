# RAG (检索增强生成) 微服务

RAG模块提供文档存储、向量嵌入、相似性搜索和知识检索等功能，是实现智能问答系统的核心组件。

## 功能特性

### 核心功能
- **文档管理**: 支持多种文档格式（文本、PDF、Markdown、HTML、Word）
- **智能分块**: 自动将文档分割成语义相关的分块
- **向量嵌入**: 使用OpenAI等模型生成文本向量表示
- **语义搜索**: 基于向量相似度的语义搜索
- **知识库管理**: 多租户知识库管理和权限控制

### 技术特性
- **DDD架构**: 完整的领域驱动设计实现
- **微服务架构**: 独立部署、水平扩展
- **分布式追踪**: 完整的链路追踪支持
- **服务发现**: 基于etcd的服务注册发现
- **配置管理**: 动态配置和密钥管理
- **高性能**: 批量处理、异步索引、缓存优化

## 项目结构

```
rag/
├── cmd/
│   └── main.go                 # 服务启动入口
├── go.mod                      # Go模块定义
├── internal/
│   ├── application/            # 应用层
│   │   └── service/           
│   │       ├── rag_service.go      # 主要业务服务
│   │       ├── commands.go         # 命令定义
│   │       ├── embedding_service.go # 嵌入服务接口
│   │       └── chunking_service.go  # 分块服务接口
│   ├── domain/                 # 领域层
│   │   ├── document.go            # 文档聚合根
│   │   ├── chunk.go               # 分块实体
│   │   ├── knowledge_base.go      # 知识库聚合根
│   │   ├── tag.go                 # 标签实体
│   │   ├── search_result.go       # 搜索结果值对象
│   │   ├── errors.go              # 领域错误
│   │   └── repository/            # 仓储接口
│   │       ├── document_repository.go
│   │       ├── chunk_repository.go
│   │       ├── knowledge_base_repository.go
│   │       └── vector_repository.go
│   ├── infrastructure/         # 基础设施层
│   │   ├── repository/            # 仓储实现
│   │   │   ├── gorm_document_repository.go
│   │   │   ├── gorm_chunk_repository.go
│   │   │   └── gorm_knowledge_base_repository.go
│   │   ├── vector/                # 向量存储实现
│   │   │   └── milvus_vector_repository.go
│   │   └── embedding/             # 嵌入服务实现
│   │       └── openai_embedding_service.go
│   ├── interface/              # 接口层
│   │   └── http/
│   │       ├── handler/
│   │       │   └── rag_handler.go  # HTTP处理器
│   │       └── router.go           # 路由配置
│   └── wire/                   # 依赖注入
│       ├── wire.go                # Wire配置
│       └── wire_gen.go            # Wire生成文件
└── README.md                   # 项目说明
```

## 核心概念

### 知识库 (Knowledge Base)
- 文档的容器，支持多租户隔离
- 可配置分块策略、嵌入模型等参数
- 支持标签分类和访问权限控制

### 文档 (Document)
- 支持多种格式：文本、PDF、Markdown、HTML、Word
- 自动计算内容哈希，避免重复存储
- 支持元数据扩展和标签管理

### 分块 (Chunk)
- 文档的语义片段，支持多种分块策略
- 存储向量嵌入用于相似度搜索
- 保留原文档位置信息便于定位

### 向量搜索
- 支持多种距离度量：余弦相似度、欧氏距离、点积
- 支持过滤条件和重排序
- 批量搜索优化性能

## API文档

### 知识库管理

#### 创建知识库
```http
POST /api/v1/knowledge-bases
Content-Type: application/json

{
  "name": "技术文档库",
  "description": "存储技术相关文档",
  "owner_id": "user_123",
  "settings": {
    "chunk_size": 1000,
    "chunk_overlap": 200,
    "embedding_model": "text-embedding-ada-002"
  }
}
```

#### 获取知识库
```http
GET /api/v1/knowledge-bases/{id}?include_documents=true&include_stats=true
```

#### 更新知识库
```http
PUT /api/v1/knowledge-bases/{id}
Content-Type: application/json

{
  "name": "更新后的名称",
  "description": "更新后的描述"
}
```

#### 列出知识库
```http
GET /api/v1/knowledge-bases?owner_id=user_123&offset=0&limit=20
```

### 文档管理

#### 添加文档
```http
POST /api/v1/documents
Content-Type: application/json

{
  "title": "API设计指南",
  "content": "本文档介绍RESTful API设计的最佳实践...",
  "type": "text",
  "knowledge_base_id": "kb_123",
  "source": "internal_docs",
  "metadata": {
    "author": "张三",
    "category": "技术文档"
  }
}
```

#### 处理文档（分块和向量化）
```http
POST /api/v1/documents/{id}/process
Content-Type: application/json

{
  "force_reprocess": false
}
```

#### 批量添加文档
```http
POST /api/v1/documents/batch
Content-Type: application/json

{
  "knowledge_base_id": "kb_123",
  "documents": [
    {
      "title": "文档1",
      "content": "内容1...",
      "type": "text"
    },
    {
      "title": "文档2", 
      "content": "内容2...",
      "type": "text"
    }
  ]
}
```

### 语义搜索

#### 搜索相关内容
```http
POST /api/v1/search
Content-Type: application/json

{
  "query": "如何设计RESTful API",
  "knowledge_base_id": "kb_123",
  "top_k": 5,
  "score_threshold": 0.7,
  "search_type": "semantic",
  "filters": {
    "document_types": ["text"],
    "tags": ["技术文档"]
  },
  "include_metadata": true
}
```

## 配置说明

### 嵌入服务配置
```go
type EmbeddingConfig struct {
    Provider   string  // "openai", "huggingface", "local"
    Model      string  // "text-embedding-ada-002"
    APIKey     string  // API密钥
    Dimension  int     // 向量维度
    BatchSize  int     // 批量大小
    Timeout    int     // 超时时间（秒）
}
```

### 分块策略配置
```go
type ChunkingConfig struct {
    Strategy     string    // "fixed_size", "semantic", "structural"
    ChunkSize    int       // 分块大小（字符数）
    ChunkOverlap int       // 重叠大小
    MinChunkSize int       // 最小分块大小
    MaxChunkSize int       // 最大分块大小
    Separators   []string  // 分隔符列表
}
```

### 向量存储配置
```go
type MilvusConfig struct {
    Host       string  // Milvus服务器地址
    Port       int     // 端口
    Username   string  // 用户名
    Password   string  // 密码
    Database   string  // 数据库名
    Timeout    int     // 超时时间
}
```

## 部署指南

### 依赖服务
- PostgreSQL: 存储结构化数据
- Milvus: 向量数据库
- etcd: 服务发现和配置管理
- Jaeger: 链路追踪

### 环境变量
```bash
# OpenAI配置
OPENAI_API_KEY=sk-xxx

# 数据库配置
DATABASE_URL=postgres://user:pass@localhost:5432/noah_loop

# etcd配置
ETCD_ENDPOINTS=localhost:2379

# Milvus配置
MILVUS_HOST=localhost
MILVUS_PORT=19530
```

### Docker部署
```bash
# 构建镜像
make docker

# 启动服务
docker-compose up -d rag-service
```

### Kubernetes部署
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: rag-service
spec:
  replicas: 2
  selector:
    matchLabels:
      app: rag-service
  template:
    metadata:
      labels:
        app: rag-service
    spec:
      containers:
      - name: rag-service
        image: noah-loop/rag-service:latest
        ports:
        - containerPort: 8084
        - containerPort: 9084
        env:
        - name: OPENAI_API_KEY
          valueFrom:
            secretKeyRef:
              name: rag-secrets
              key: openai-api-key
```

## 开发指南

### 本地开发
```bash
# 安装依赖
go mod download

# 生成Wire代码
make wire-gen

# 启动服务
make dev

# 运行测试
make test
```

### 扩展支持的文档格式
1. 在`DocumentType`中添加新类型
2. 在`ChunkingService`中实现对应的预处理逻辑
3. 添加相应的测试用例

### 添加新的嵌入服务提供商
1. 实现`EmbeddingService`接口
2. 在Wire配置中添加提供者
3. 更新配置结构支持新提供商

### 性能优化建议
- 使用Redis缓存常用的嵌入向量
- 实现异步文档处理队列
- 优化数据库查询和索引
- 使用连接池管理向量数据库连接

## 监控和运维

### 健康检查
```bash
# HTTP健康检查
curl http://localhost:8084/health

# gRPC健康检查
grpc-health-probe -addr=localhost:9084
```

### 指标监控
- 文档处理速度
- 嵌入生成延迟
- 搜索查询QPS
- 向量数据库性能

### 日志级别
- ERROR: 系统错误和异常
- WARN: 警告信息和降级处理
- INFO: 关键业务操作
- DEBUG: 详细的调试信息

## 故障排查

### 常见问题
1. **嵌入生成失败**: 检查OpenAI API密钥和网络连接
2. **向量搜索慢**: 检查Milvus索引状态和查询参数
3. **文档处理卡住**: 检查分块配置和内存使用
4. **服务注册失败**: 检查etcd连接和服务配置

### 调试工具
- Jaeger UI: 查看链路追踪
- Prometheus: 监控指标
- 日志聚合: 集中日志分析

## 更新日志

### v1.0.0
- 初始版本发布
- 支持基本的文档管理和语义搜索
- 完整的DDD架构实现
- 微服务基础设施集成
