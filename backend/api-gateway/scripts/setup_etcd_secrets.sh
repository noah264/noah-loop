#!/bin/bash

# Noah-Loop etcd密钥设置脚本
# 用于初始化etcd中的关键密钥和配置

set -e

# etcd配置
ETCD_ENDPOINT=${ETCD_ENDPOINT:-"http://localhost:2379"}
ETCDCTL_API=3

echo "Setting up etcd secrets and configuration..."

# 检查etcd连接
if ! etcdctl --endpoints="$ETCD_ENDPOINT" endpoint health > /dev/null 2>&1; then
    echo "Error: Cannot connect to etcd at $ETCD_ENDPOINT"
    echo "Please make sure etcd is running:"
    echo "  docker run -d --name etcd-server -p 2379:2379 -p 2380:2380 --env ALLOW_NONE_AUTHENTICATION=yes --env ETCD_ADVERTISE_CLIENT_URLS=http://0.0.0.0:2379 bitnami/etcd:latest"
    exit 1
fi

echo "✅ etcd connection successful"

# 设置OpenAI API密钥
read -s -p "请输入OpenAI API Key (留空跳过): " OPENAI_KEY
if [ ! -z "$OPENAI_KEY" ]; then
    etcdctl --endpoints="$ETCD_ENDPOINT" put /noah-loop/secrets/openai/api_key "$OPENAI_KEY"
    echo "✅ OpenAI API密钥已设置"
fi

# 设置数据库密码
read -s -p "请输入数据库密码 (默认: postgres): " DB_PASSWORD
DB_PASSWORD=${DB_PASSWORD:-postgres}
etcdctl --endpoints="$ETCD_ENDPOINT" put /noah-loop/secrets/database/password "$DB_PASSWORD"
echo "✅ 数据库密码已设置"

# 设置Redis密码
read -s -p "请输入Redis密码 (留空跳过): " REDIS_PASSWORD
if [ ! -z "$REDIS_PASSWORD" ]; then
    etcdctl --endpoints="$ETCD_ENDPOINT" put /noah-loop/secrets/redis/password "$REDIS_PASSWORD"
    echo "✅ Redis密码已设置"
fi

# 设置JWT密钥
JWT_SECRET=$(openssl rand -base64 32 2>/dev/null || echo "your-super-secret-jwt-key-$(date +%s)")
etcdctl --endpoints="$ETCD_ENDPOINT" put /noah-loop/secrets/jwt/secret "$JWT_SECRET"
echo "✅ JWT密钥已生成并设置"

# 设置API网关配置
etcdctl --endpoints="$ETCD_ENDPOINT" put /noah-loop/config/gateway/max_requests_per_minute "1000"
etcdctl --endpoints="$ETCD_ENDPOINT" put /noah-loop/config/gateway/timeout_seconds "30"
etcdctl --endpoints="$ETCD_ENDPOINT" put /noah-loop/config/gateway/circuit_breaker_threshold "5"
echo "✅ API网关配置已设置"

# 设置LLM服务配置
etcdctl --endpoints="$ETCD_ENDPOINT" put /noah-loop/config/llm/default_model "gpt-3.5-turbo"
etcdctl --endpoints="$ETCD_ENDPOINT" put /noah-loop/config/llm/max_tokens "4096"
etcdctl --endpoints="$ETCD_ENDPOINT" put /noah-loop/config/llm/temperature "0.7"
echo "✅ LLM服务配置已设置"

# 设置Agent服务配置
etcdctl --endpoints="$ETCD_ENDPOINT" put /noah-loop/config/agent/max_memory_size "10000"
etcdctl --endpoints="$ETCD_ENDPOINT" put /noah-loop/config/agent/default_timeout "60"
echo "✅ Agent服务配置已设置"

# 设置日志配置
etcdctl --endpoints="$ETCD_ENDPOINT" put /noah-loop/config/app/log_level "info"
etcdctl --endpoints="$ETCD_ENDPOINT" put /noah-loop/config/app/log_format "json"
echo "✅ 日志配置已设置"

# 设置服务发现配置
etcdctl --endpoints="$ETCD_ENDPOINT" put /noah-loop/config/discovery/health_check_interval "10s"
etcdctl --endpoints="$ETCD_ENDPOINT" put /noah-loop/config/discovery/service_timeout "30s"
echo "✅ 服务发现配置已设置"

echo ""
echo "🎉 etcd密钥和配置设置完成！"
echo ""
echo "查看设置的密钥:"
echo "  etcdctl --endpoints=$ETCD_ENDPOINT get /noah-loop/secrets/ --prefix --keys-only"
echo ""
echo "查看设置的配置:"
echo "  etcdctl --endpoints=$ETCD_ENDPOINT get /noah-loop/config/ --prefix"
echo ""
echo "启动Noah-Loop服务:"
echo "  cd backend/api-gateway && make run"
