# Noah-Loop 镜像构建和部署

本目录包含 Noah-Loop 微服务平台的 Docker 镜像构建和部署配置。

## 📁 目录结构

```
release/image/
├── Dockerfile.*              # 各微服务的 Dockerfile
├── docker-compose.build.yml  # 镜像构建配置
├── docker-compose.yml        # 完整部署配置
├── env.template              # 环境变量模板
├── scripts/
│   ├── build.sh             # 镜像构建脚本
│   └── deploy.sh            # 部署管理脚本
└── README.md                # 本文件
```

## 🚀 快速开始

### 1. 准备环境

```bash
# 复制环境变量模板
cp env.template .env

# 根据实际情况修改配置
vim .env
```

### 2. 构建镜像

```bash
# 构建所有微服务镜像
./scripts/build.sh

# 构建指定服务
./scripts/build.sh agent llm

# 构建并推送到镜像仓库
./scripts/build.sh -v 1.0.0 --push

# 无缓存构建
./scripts/build.sh --no-cache
```

### 3. 部署服务

```bash
# 启动所有服务（前台）
./scripts/deploy.sh up

# 后台启动服务
./scripts/deploy.sh -d up

# 停止所有服务
./scripts/deploy.sh down

# 重启服务
./scripts/deploy.sh restart

# 查看服务状态
./scripts/deploy.sh ps

# 查看日志
./scripts/deploy.sh logs

# 检查健康状态
./scripts/deploy.sh health
```

## 🏗️ 镜像构建

### 构建脚本使用

`scripts/build.sh` 支持以下选项：

- `-v, --version VERSION`: 指定版本号（默认：1.0.0）
- `-r, --registry REGISTRY`: 指定镜像仓库（默认：docker.io）
- `-n, --namespace NS`: 指定命名空间（默认：noah-loop）
- `-p, --parallel N`: 并行构建数量（默认：4）
- `--no-cache`: 不使用缓存构建
- `--push`: 构建后推送到仓库
- `-h, --help`: 显示帮助信息

### 构建示例

```bash
# 构建所有服务
./scripts/build.sh

# 构建特定版本
./scripts/build.sh -v 1.1.0

# 构建并推送到自定义仓库
./scripts/build.sh -r registry.example.com -n my-namespace --push

# 高并发构建
./scripts/build.sh -p 8

# 构建特定服务
./scripts/build.sh agent llm rag
```

## 🐳 Docker Compose

### 构建配置 (docker-compose.build.yml)

用于构建所有微服务镜像：

```bash
# 构建所有镜像
docker-compose -f docker-compose.build.yml build

# 构建特定服务
docker-compose -f docker-compose.build.yml build agent

# 无缓存构建
docker-compose -f docker-compose.build.yml build --no-cache
```

### 部署配置 (docker-compose.yml)

包含完整的应用栈：

**基础设施服务：**
- PostgreSQL：数据库
- Redis：缓存
- etcd：服务发现和配置中心
- Jaeger：分布式链路追踪

**应用服务：**
- api-gateway：API 网关
- agent：智能体服务
- llm：大语言模型服务
- mcp：模型上下文协议服务
- orchestrator：编排器服务
- rag：检索增强生成服务
- notify：通知服务

## 🔧 配置管理

### 环境变量

主要环境变量配置：

```bash
# 应用配置
VERSION=1.0.0
ENVIRONMENT=development
DEBUG=false

# 数据库配置
DATABASE_HOST=postgres
DATABASE_PASSWORD=postgres123

# Redis 配置
REDIS_PASSWORD=redis123

# LLM 配置
OPENAI_API_KEY=your_api_key

# 通知配置
SMTP_HOST=smtp.gmail.com
SMTP_USERNAME=your_email@gmail.com
```

### 端口映射

| 服务 | HTTP 端口 | gRPC 端口 | Web UI |
|------|-----------|-----------|---------|
| API Gateway | 8080 | 9090 | - |
| Agent | 8081 | 9091 | - |
| LLM | 8082 | 9092 | - |
| MCP | 8083 | 9093 | - |
| Orchestrator | 8084 | 9094 | - |
| RAG | 8085 | 9095 | - |
| Notify | 8086 | 9096 | - |
| PostgreSQL | 5432 | - | - |
| Redis | 6379 | - | - |
| etcd | 2379/2380 | - | - |
| Jaeger | - | - | 16686 |

## 📊 监控和日志

### 查看日志

```bash
# 查看所有服务日志
./scripts/deploy.sh logs

# 查看特定服务日志
./scripts/deploy.sh logs agent

# 实时跟踪日志
./scripts/deploy.sh logs -f

# 查看最近100行日志
./scripts/deploy.sh logs --tail=100
```

### 健康检查

```bash
# 检查所有服务健康状态
./scripts/deploy.sh health

# 使用 curl 检查特定服务
curl http://localhost:8080/health  # API Gateway
curl http://localhost:8081/health  # Agent
curl http://localhost:8082/health  # LLM
```

### 链路追踪

访问 Jaeger UI：http://localhost:16686

## 🔍 故障排除

### 常见问题

1. **端口冲突**
   ```bash
   # 检查端口占用
   netstat -tlnp | grep :8080
   # 或使用 lsof
   lsof -i :8080
   ```

2. **内存不足**
   ```bash
   # 检查 Docker 资源使用
   docker stats
   
   # 清理未使用资源
   ./scripts/deploy.sh clean
   ```

3. **数据库连接失败**
   ```bash
   # 检查数据库状态
   docker-compose exec postgres pg_isready -U postgres
   
   # 查看数据库日志
   ./scripts/deploy.sh logs postgres
   ```

4. **镜像拉取失败**
   ```bash
   # 手动拉取镜像
   ./scripts/deploy.sh pull
   
   # 或使用特定版本
   VERSION=1.0.0 ./scripts/deploy.sh pull
   ```

### 调试模式

启用调试模式：

```bash
# 设置环境变量
echo "DEBUG=true" >> .env
echo "LOG_LEVEL=debug" >> .env

# 重启服务
./scripts/deploy.sh restart
```

## 🚀 生产部署

### 安全配置

1. **修改默认密码**
   ```bash
   # 生成强密码
   DATABASE_PASSWORD=$(openssl rand -base64 32)
   REDIS_PASSWORD=$(openssl rand -base64 32)
   JWT_SECRET=$(openssl rand -base64 64)
   ```

2. **启用 TLS**
   ```bash
   # 配置 HTTPS
   ETCD_TLS_CERT_FILE=/path/to/cert
   ETCD_TLS_KEY_FILE=/path/to/key
   ```

3. **限制资源**
   ```yaml
   # 在 docker-compose.yml 中添加资源限制
   deploy:
     resources:
       limits:
         cpus: '2.0'
         memory: 2G
   ```

### 性能优化

1. **调整并发数**
   ```bash
   GO_MAX_PROCS=8
   DB_MAX_OPEN_CONNS=50
   REDIS_POOL_SIZE=20
   ```

2. **启用缓存**
   ```bash
   CACHE_ENABLED=true
   CACHE_TTL=3600
   ```

3. **配置 JVM 参数**（如果使用 Java 组件）
   ```bash
   JAVA_OPTS="-Xms2g -Xmx4g -XX:+UseG1GC"
   ```

## 📚 更多文档

- [Helm Chart 部署文档](../deployment/noah-loop/README.md)
- [开发指南](../../backend/README.md)
- [API 文档](../../docs/api/)
- [运维手册](../../docs/ops/)
