# API Gateway - Noah Loop

Noah Loop 项目的 API 网关，提供统一的入口和路由分发功能。

## 功能特性

### 🔄 反向代理
- 自动服务发现和路由分发
- 支持多个微服务的统一入口
- 智能请求转发和响应处理

### ⚖️ 负载均衡
- 轮询负载均衡策略
- 加权轮询支持
- 自动故障转移

### 🔐 安全防护
- JWT token 认证支持
- 角色权限控制
- 请求验证和安全头设置

### 🚦 流量控制
- 基于IP的限流保护
- 服务级别限流
- 请求超时控制

### 🔧 故障处理
- 熔断器机制
- 自动重试策略
- 优雅降级处理

### 📊 监控指标
- 请求监控和统计
- 服务健康检查
- 性能指标收集

## 项目结构

```
api-gateway/
├── cmd/
│   └── main.go                       # 应用入口
├── go.mod                            # 依赖管理
├── Makefile                          # 构建管理
├── README.md                         # 说明文档
├── WIRE_USAGE.md                     # Wire使用指南
├── DDD_ARCHITECTURE.md               # DDD架构文档
└── internal/                         # 内部实现
    ├── domain/                       # 领域层
    │   ├── entity/                   # 实体
    │   │   ├── gateway.go            # 网关聚合根
    │   │   └── service.go            # 服务实体
    │   ├── valueobject/              # 值对象
    │   │   └── route.go              # 路由规则
    │   ├── service/                  # 领域服务
    │   │   ├── loadbalancer.go       # 负载均衡器
    │   │   └── circuit_breaker.go    # 熔断器
    │   └── repository/               # 仓储接口
    │       └── service_repository.go
    ├── application/                  # 应用层
    │   └── service/
    │       └── gateway_service.go    # 网关应用服务
    ├── infrastructure/               # 基础设施层
    │   ├── config/
    │   │   └── config_adapter.go     # 配置适配器
    │   └── repository/
    │       └── in_memory_service_repository.go
    ├── interface/                    # 接口层
    │   └── http/
    │       ├── handler/
    │       │   └── gateway_handler.go # HTTP处理器
    │       ├── router/
    │       │   └── router.go         # 路由配置
    │       └── middleware/           # 中间件
    │           ├── ratelimiter.go
    │           ├── timeout.go
    │           ├── circuit_breaker.go
    │           └── validation.go
    └── wire/                         # Wire依赖注入
        ├── wire.go                   # Wire配置
        └── wire_gen.go               # Wire生成代码
```

## 路由设计

### 健康检查
- `GET /health` - 网关健康状态
- `GET /health/services` - 上游服务健康状态

### 管理接口
- `GET /management/info` - 网关信息
- `GET /management/services` - 服务状态列表

### 业务API代理 (统一前缀: `/api/v1`)
- `/api/v1/agent/*` → Agent服务 (端口:8081)
- `/api/v1/llm/*` → LLM服务 (端口:8082)  
- `/api/v1/mcp/*` → MCP服务 (端口:8083)
- `/api/v1/orchestrator/*` → 编排服务 (端口:8084)

### 监控指标
- `GET /metrics` - Prometheus格式指标

## 配置说明

网关使用共享配置文件 `configs/config.yaml`:

```yaml
# API网关端口配置
http:
  port: 8080
  read_timeout: "30s"
  write_timeout: "30s"
  idle_timeout: "120s"
  shutdown_timeout: "30s"

# 上游服务配置
services:
  agent:
    port: 8081
  llm:
    port: 8082
  mcp:
    port: 8083
  orchestrator:
    port: 8084
```

## 启动方式

### 1. 开发环境启动

```bash
# 进入网关目录
cd backend/api-gateway

# 安装开发工具（首次运行）
make install-tools

# 生成Wire依赖注入代码
make generate

# 启动网关（开发模式，带热重载）
make dev

# 或常规启动
make run
```

### 2. 手动启动

```bash
# 生成Wire代码
wire ./internal/wire

# 启动应用
go run cmd/main.go
```

### 3. 构建和部署

```bash
# 构建二进制文件
make build

# 运行构建后的文件
./build/api-gateway

# Docker构建和运行
make docker-build
make docker-run
```

## 使用示例

### 通过网关访问Agent服务

```bash
# 创建智能体（代理到Agent服务）
curl -X POST http://localhost:8080/api/v1/agent/agents \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Assistant Bot",
    "type": "conversational",
    "description": "A helpful assistant"
  }'
```

### 通过网关访问LLM服务

```bash
# 创建模型（代理到LLM服务）
curl -X POST http://localhost:8080/api/v1/llm/models \
  -H "Content-Type: application/json" \
  -d '{
    "name": "gpt-4",
    "provider": "openai",
    "type": "chat"
  }'
```

### 检查服务状态

```bash
# 网关健康检查
curl http://localhost:8080/health

# 上游服务状态
curl http://localhost:8080/health/services

# 网关信息
curl http://localhost:8080/management/info
```

## 中间件功能

### 限流器
- 基于IP的全局限流：每分钟100请求，突发20请求
- 可配置的服务级别限流

### 熔断器
- 失败阈值：连续5次失败触发熔断
- 超时时间：60秒后尝试半开
- 半开状态：允许3个探测请求

### 认证授权（可选）
- 支持Bearer Token认证
- 角色权限控制
- 公共路径白名单

### 请求超时
- 默认30秒超时保护
- 可配置的超时时间

## 监控和日志

### 日志格式
网关使用结构化日志，包含以下信息：
- 请求ID跟踪
- 请求方法和路径
- 响应状态码
- 处理时间
- 客户端IP
- 代理目标服务

### 指标收集
- 请求总数和成功率
- 响应时间分布
- 服务健康状态
- 熔断器状态

## 高级功能

### 自定义负载均衡

网关支持多种负载均衡策略：

1. **轮询（Round Robin）** - 默认策略
2. **加权轮询（Weighted Round Robin）** - 支持权重配置

### 服务发现

网关会自动：
- 检测上游服务健康状态
- 动态路由健康的服务实例
- 故障服务自动摘除

### 请求转换

网关自动添加以下请求头：
- `X-Gateway`: 网关标识
- `X-Gateway-Version`: 网关版本
- `X-Forwarded-By`: 转发标识
- `X-Original-Host`: 原始主机信息

## 故障排查

### 常见问题

1. **服务连接失败**
   ```bash
   # 检查上游服务是否启动
   curl http://localhost:8081/health  # Agent服务
   curl http://localhost:8082/health  # LLM服务
   ```

2. **限流触发**
   ```bash
   # 检查限流状态，调整请求频率
   # 响应：HTTP 429 Too Many Requests
   ```

3. **熔断器触发**
   ```bash
   # 检查服务状态
   curl http://localhost:8080/health/services
   # 响应：HTTP 503 Service Unavailable
   ```

### 日志查看

```bash
# 实时查看日志
go run main.go

# 日志包含的关键信息：
# - request_id: 请求跟踪ID
# - method: HTTP方法
# - path: 请求路径
# - status: 响应状态码
# - latency: 处理时间
```

## 扩展开发

### 添加新的中间件

```go
func CustomMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 处理逻辑
        c.Next()
    }
}
```

### 添加新的服务路由

在 `handler/gateway_handler.go` 中添加服务配置：

```go
h.services["new_service"] = &ServiceConfig{
    Name: "new_service", 
    Host: "localhost",
    Port: 8085,
    Path: "/api/v1/new_service",
}
```

## 性能考量

- **并发处理**: 使用Gin框架的高性能HTTP处理
- **连接池**: 复用HTTP连接减少延迟
- **内存管理**: 合理的缓存和对象复用
- **监控指标**: 实时性能监控和告警

## 安全建议

1. **启用认证**: 生产环境建议启用JWT认证
2. **HTTPS配置**: 配置SSL证书加密传输
3. **限流策略**: 根据业务场景调整限流参数
4. **监控告警**: 设置异常状态告警机制
5. **日志审计**: 保存关键操作的审计日志

这个API网关为Noah Loop微服务架构提供了强大的统一入口，具备了生产环境所需的各种功能。

## Wire依赖注入

网关使用Google Wire进行编译时依赖注入，相比运行时注入有以下优势：

### 优势
- ✅ **编译时生成**: 无运行时开销，性能更优
- ✅ **类型安全**: 编译时检查依赖关系
- ✅ **易于调试**: 生成代码清晰可读
- ✅ **统一管理**: 与其他模块保持一致的依赖管理方式

### 使用方法
```bash
# 安装开发工具
make install-tools

# 生成Wire代码
make generate

# 运行应用
make run
```

详细使用说明请参考 [WIRE_USAGE.md](./WIRE_USAGE.md)

## DDD架构

网关采用标准的DDD（领域驱动设计）架构，与项目其他模块保持一致：

- **领域层**: Gateway聚合根、Service实体、Route值对象
- **应用层**: GatewayService应用服务
- **基础设施层**: 配置适配器、内存仓储
- **接口层**: HTTP处理器、路由、中间件

详细架构说明请参考 [DDD_ARCHITECTURE.md](./DDD_ARCHITECTURE.md)
