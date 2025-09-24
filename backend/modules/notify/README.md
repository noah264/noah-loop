# Notify 通知微服务

Notify模块提供多渠道通知发送功能，支持邮件、短信、推送、Webhook等多种通知方式，是实现系统消息推送的核心组件。

## 功能特性

### 核心功能
- **多渠道通知**: 支持邮件(SMTP)、短信(阿里云)、推送(Bark)、Webhook(Server酱)等
- **模板系统**: 灵活的通知模板管理，支持变量替换和多版本管理
- **渠道配置**: 独立的渠道配置管理，支持多租户配置
- **批量发送**: 支持批量通知发送和接收者管理
- **重试机制**: 自动重试失败的通知，支持指数退避

### 技术特性
- **DDD架构**: 完整的领域驱动设计实现
- **微服务架构**: 独立部署、水平扩展
- **分布式追踪**: 完整的链路追踪支持
- **服务发现**: 基于etcd的服务注册发现
- **定时任务**: 支持定时通知和失败重试

## 支持的通知渠道

### 📧 邮件通知 (Email)
- **提供商**: SMTP
- **配置项**: smtp_host, smtp_port, smtp_username, smtp_password, use_tls
- **功能**: 支持HTML/纯文本、抄送密送、附件

### 📱 短信通知 (SMS)  
- **提供商**: 阿里云短信
- **配置项**: access_key, secret_key, sign_name, template_code
- **功能**: 支持模板短信、变量替换

### 🔔 推送通知 (Push)
- **提供商**: Bark (iOS推送)
- **配置项**: device_key, server_url, sound, group
- **功能**: 支持声音、分组、URL跳转

### 🪝 Webhook通知
- **提供商**: Server酱、通用Webhook
- **配置项**: send_key/url, method, headers
- **功能**: 支持微信推送、自定义Webhook

### 🔔 钉钉通知 (规划中)
- **提供商**: 钉钉机器人
- **功能**: 支持@用户、卡片消息

## 项目结构

```
notify/
├── cmd/main.go                    # 服务启动入口
├── go.mod                         # Go模块定义
├── internal/
│   ├── application/service/       # 应用服务层
│   │   ├── notification_service.go   # 通知服务
│   │   ├── template_service.go       # 模板服务
│   │   ├── channel_service.go        # 渠道服务
│   │   ├── commands.go               # 命令定义
│   │   └── providers.go              # 提供商接口
│   ├── domain/                    # 领域层
│   │   ├── notification.go           # 通知聚合根
│   │   ├── recipient.go              # 接收者实体
│   │   ├── channel.go                # 渠道配置
│   │   ├── template.go               # 通知模板
│   │   ├── errors.go                 # 领域错误
│   │   └── repository/               # 仓储接口
│   ├── infrastructure/            # 基础设施层
│   │   ├── provider/                 # 通知提供商实现
│   │   │   ├── smtp_email_provider.go
│   │   │   ├── aliyun_sms_provider.go
│   │   │   ├── bark_push_provider.go
│   │   │   └── serverchan_webhook_provider.go
│   │   └── repository/               # GORM仓储实现
│   ├── interface/http/            # HTTP接口层
│   │   ├── handler/
│   │   │   └── notify_handler.go     # HTTP处理器
│   │   └── router.go                 # 路由配置
│   └── wire/                      # 依赖注入
│       ├── wire.go                   # Wire配置
│       └── wire_gen.go               # Wire生成文件
└── README.md                      # 项目说明
```

## API端点

### 通知管理

#### 创建通知
```http
POST /api/v1/notifications
Content-Type: application/json

{
  "title": "系统通知",
  "content": "您有新的消息",
  "type": "system",
  "channel": "email",
  "priority": "normal",
  "recipients": [
    {
      "type": "email",
      "identifier": "user@example.com",
      "name": "用户名"
    }
  ],
  "created_by": "system"
}
```

#### 从模板创建通知
```http
POST /api/v1/notifications/template
Content-Type: application/json

{
  "template_id": "welcome_template",
  "channel": "email",
  "variables": {
    "username": "张三",
    "product_name": "Noah-Loop"
  },
  "recipients": [
    {
      "type": "email",
      "identifier": "user@example.com"
    }
  ],
  "created_by": "system"
}
```

#### 发送通知
```http
POST /api/v1/notifications/{id}/send
```

### 模板管理

#### 创建模板
```http
POST /api/v1/templates
Content-Type: application/json

{
  "name": "欢迎模板",
  "code": "welcome_template",
  "type": "html",
  "subject": "欢迎使用{{product_name}}",
  "content": "亲爱的{{username}}，欢迎使用{{product_name}}！",
  "variables": [
    {
      "name": "username",
      "display_name": "用户名",
      "required": true
    },
    {
      "name": "product_name",
      "display_name": "产品名称",
      "default_value": "Noah-Loop"
    }
  ],
  "created_by": "admin"
}
```

### 渠道配置

#### 创建邮件渠道配置
```http
POST /api/v1/channels
Content-Type: application/json

{
  "channel": "email",
  "name": "公司邮件服务",
  "config": {
    "smtp_host": "smtp.example.com",
    "smtp_port": "587",
    "smtp_username": "noreply@example.com",
    "smtp_password": "password",
    "from_name": "Noah-Loop系统",
    "use_tls": "true"
  },
  "owner_id": "admin"
}
```

#### 创建Bark推送配置
```http
POST /api/v1/channels
Content-Type: application/json

{
  "channel": "bark",
  "name": "Bark推送",
  "config": {
    "device_key": "your_bark_device_key",
    "server_url": "https://api.day.app",
    "sound": "default",
    "group": "Noah-Loop"
  },
  "owner_id": "admin"
}
```

#### 测试渠道配置
```http
POST /api/v1/channels/test
Content-Type: application/json

{
  "channel_id": "channel_123",
  "test_data": {
    "email": "test@example.com"
  }
}
```

## 配置说明

### 服务配置 (config.yaml)
```yaml
services:
  notify:
    port: 8086
    grpc_port: 9096
```

### 环境变量
- `DATABASE_URL`: 数据库连接
- `ETCD_ENDPOINTS`: etcd集群地址
- `SMTP_PASSWORD`: SMTP密码（推荐使用etcd存储）
- `ALIYUN_ACCESS_KEY`: 阿里云访问密钥
- `BARK_DEVICE_KEY`: Bark设备密钥

## 快速开始

### 本地开发
```bash
cd modules/notify
go mod download
make dev
```

### Docker部署
```bash
make docker
docker run -p 8086:8086 noah-loop/notify-service
```

### 使用示例

#### 1. 发送简单邮件通知
```bash
curl -X POST http://localhost:8086/api/v1/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "title": "测试邮件",
    "content": "这是一条测试消息",
    "type": "system",
    "channel": "email", 
    "recipients": [
      {
        "type": "email",
        "identifier": "test@example.com"
      }
    ],
    "created_by": "test"
  }'
```

#### 2. 发送Bark推送
```bash
curl -X POST http://localhost:8086/api/v1/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "title": "系统警报",
    "content": "服务器CPU使用率过高",
    "type": "alert",
    "channel": "bark",
    "priority": "high",
    "recipients": [
      {
        "type": "device",
        "identifier": "admin_device"
      }
    ],
    "created_by": "monitor"
  }'
```

## 扩展开发

### 添加新的通知提供商

1. 实现Provider接口
```go
type CustomProvider struct {
    logger infrastructure.Logger
}

func (p *CustomProvider) SendNotification(ctx context.Context, data *NotificationData, config *ChannelConfig) error {
    // 实现发送逻辑
    return nil
}
```

2. 在Wire中注册
```go
var CustomProviderSet = wire.NewSet(
    NewCustomProvider,
    wire.Bind(new(service.CustomProvider), new(*CustomProvider)),
)
```

### 自定义模板引擎
可以扩展模板引擎支持更复杂的模板语法，如条件判断、循环等。

## 监控运维

### 健康检查
```bash
curl http://localhost:8086/health
```

### 关键指标
- 通知发送成功率
- 各渠道响应时间
- 模板渲染性能
- 失败重试次数

### 故障排查
1. **邮件发送失败**: 检查SMTP配置和网络连接
2. **短信发送失败**: 检查阿里云配置和余额
3. **推送失败**: 检查Bark设备密钥和服务器地址
4. **模板渲染错误**: 检查变量名称和必需参数

## 最佳实践

1. **渠道配置**: 敏感信息存储在etcd中，支持动态更新
2. **错误处理**: 失败通知自动重试，支持指数退避
3. **性能优化**: 批量发送、异步处理、连接池
4. **安全考虑**: API认证、敏感数据加密、访问日志

Notify模块为Noah-Loop系统提供了强大的多渠道通知能力，支持各种业务场景的消息推送需求！📬
