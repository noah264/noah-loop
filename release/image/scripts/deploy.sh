#!/bin/bash

# Noah-Loop 部署脚本
# 使用方法：
#   ./deploy.sh [options] [command]
#
# 命令：
#   up          启动所有服务（默认）
#   down        停止所有服务
#   restart     重启所有服务
#   logs        查看日志
#   ps          查看服务状态
#   pull        拉取最新镜像
#   clean       清理未使用的镜像和容器
#
# 选项：
#   -f, --file FILE      指定 compose 文件（默认：docker-compose.yml）
#   -e, --env ENV        指定环境文件（默认：.env）
#   -v, --version VER    指定版本号
#   -d, --detach         后台运行
#   -p, --project NAME   指定项目名称
#   --build             重新构建镜像
#   -h, --help          显示帮助信息

set -euo pipefail

# 默认配置
COMPOSE_FILE="docker-compose.yml"
ENV_FILE=""
VERSION=""
DETACH=""
PROJECT_NAME="noah-loop"
REBUILD=""
COMMAND="up"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示帮助
show_help() {
    cat << EOF
Noah-Loop 部署脚本

使用方法：
    $0 [options] [command]

命令：
    up          启动所有服务（默认）
    down        停止所有服务
    restart     重启所有服务
    logs        查看日志
    ps          查看服务状态
    pull        拉取最新镜像
    clean       清理未使用的镜像和容器
    health      检查服务健康状态

选项：
    -f, --file FILE      指定 compose 文件（默认：${COMPOSE_FILE}）
    -e, --env ENV        指定环境文件（默认：.env）
    -v, --version VER    指定版本号
    -d, --detach         后台运行
    -p, --project NAME   指定项目名称（默认：${PROJECT_NAME}）
    --build             重新构建镜像
    -h, --help          显示帮助信息

示例：
    $0                          # 启动所有服务
    $0 -d up                    # 后台启动所有服务
    $0 down                     # 停止所有服务
    $0 logs -f agent            # 跟踪 agent 服务日志
    $0 -v 1.1.0 up             # 使用指定版本启动
    $0 --build up              # 重新构建镜像后启动

EOF
}

# 解析命令行参数
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -f|--file)
                COMPOSE_FILE="$2"
                shift 2
                ;;
            -e|--env)
                ENV_FILE="$2"
                shift 2
                ;;
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -d|--detach)
                DETACH="-d"
                shift
                ;;
            -p|--project)
                PROJECT_NAME="$2"
                shift 2
                ;;
            --build)
                REBUILD="--build"
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            up|down|restart|logs|ps|pull|clean|health)
                COMMAND="$1"
                shift
                break
                ;;
            -*)
                log_error "未知选项: $1"
                show_help
                exit 1
                ;;
            *)
                COMMAND="$1"
                shift
                break
                ;;
        esac
    done
}

# 检查 Docker Compose 是否可用
check_docker_compose() {
    if command -v docker-compose &> /dev/null; then
        DOCKER_COMPOSE="docker-compose"
    elif docker compose version &> /dev/null; then
        DOCKER_COMPOSE="docker compose"
    else
        log_error "Docker Compose 未安装或不可用"
        exit 1
    fi
}

# 检查必需文件
check_required_files() {
    if [[ ! -f "$COMPOSE_FILE" ]]; then
        log_error "Compose 文件不存在: $COMPOSE_FILE"
        exit 1
    fi

    if [[ -n "$ENV_FILE" ]] && [[ ! -f "$ENV_FILE" ]]; then
        log_error "环境文件不存在: $ENV_FILE"
        exit 1
    fi

    # 检查环境文件
    if [[ -z "$ENV_FILE" ]]; then
        if [[ -f ".env" ]]; then
            ENV_FILE=".env"
        elif [[ -f "env.template" ]]; then
            log_warn "未找到 .env 文件，请从 env.template 复制并修改配置"
            log_info "执行: cp env.template .env"
            exit 1
        fi
    fi
}

# 构建 Docker Compose 命令
build_compose_cmd() {
    local cmd="$DOCKER_COMPOSE"
    
    cmd+=" -p $PROJECT_NAME"
    cmd+=" -f $COMPOSE_FILE"
    
    if [[ -n "$ENV_FILE" ]]; then
        cmd+=" --env-file $ENV_FILE"
    fi
    
    echo "$cmd"
}

# 设置环境变量
setup_environment() {
    if [[ -n "$VERSION" ]]; then
        export VERSION="$VERSION"
    fi
    
    # 设置默认环境变量
    export BUILD_TIME=${BUILD_TIME:-$(date -u +'%Y-%m-%dT%H:%M:%SZ')}
    if command -v git &> /dev/null && git rev-parse --git-dir > /dev/null 2>&1; then
        export GIT_COMMIT=${GIT_COMMIT:-$(git rev-parse --short HEAD)}
    else
        export GIT_COMMIT=${GIT_COMMIT:-unknown}
    fi
}

# 启动服务
cmd_up() {
    local compose_cmd=$(build_compose_cmd)
    
    log_info "启动 Noah-Loop 服务..."
    if [[ -n "$VERSION" ]]; then
        log_info "版本: $VERSION"
    fi
    
    # 创建网络和卷（如果不存在）
    $compose_cmd up --no-start
    
    # 启动基础设施服务
    log_info "启动基础设施服务..."
    $compose_cmd up $DETACH postgres redis etcd jaeger
    
    # 等待基础设施服务就绪
    if [[ -z "$DETACH" ]]; then
        sleep 10
    else
        log_info "等待基础设施服务启动..."
        for i in {1..30}; do
            if $compose_cmd ps postgres redis etcd jaeger | grep -q "Up (healthy)"; then
                break
            fi
            sleep 2
        done
    fi
    
    # 启动应用服务
    log_info "启动应用服务..."
    $compose_cmd up $DETACH $REBUILD
    
    if [[ -n "$DETACH" ]]; then
        log_success "服务已在后台启动"
        echo
        log_info "查看服务状态: $0 ps"
        log_info "查看日志: $0 logs"
        log_info "停止服务: $0 down"
    fi
}

# 停止服务
cmd_down() {
    local compose_cmd=$(build_compose_cmd)
    
    log_info "停止 Noah-Loop 服务..."
    $compose_cmd down --remove-orphans
    
    log_success "服务已停止"
}

# 重启服务
cmd_restart() {
    cmd_down
    sleep 2
    cmd_up
}

# 查看日志
cmd_logs() {
    local compose_cmd=$(build_compose_cmd)
    
    $compose_cmd logs "$@"
}

# 查看服务状态
cmd_ps() {
    local compose_cmd=$(build_compose_cmd)
    
    log_info "Noah-Loop 服务状态:"
    $compose_cmd ps
}

# 拉取镜像
cmd_pull() {
    local compose_cmd=$(build_compose_cmd)
    
    log_info "拉取最新镜像..."
    $compose_cmd pull
    
    log_success "镜像拉取完成"
}

# 清理资源
cmd_clean() {
    log_info "清理未使用的 Docker 资源..."
    
    # 停止并删除容器
    docker container prune -f
    
    # 删除未使用的镜像
    docker image prune -f
    
    # 删除未使用的网络
    docker network prune -f
    
    # 删除未使用的卷（谨慎操作）
    echo -n "是否删除未使用的卷？这将删除所有数据！[y/N] "
    read -r response
    if [[ "$response" =~ ^[Yy]$ ]]; then
        docker volume prune -f
        log_warn "未使用的卷已删除"
    fi
    
    log_success "清理完成"
}

# 检查服务健康状态
cmd_health() {
    local compose_cmd=$(build_compose_cmd)
    
    log_info "检查服务健康状态..."
    
    # 获取所有服务状态
    local services=($($compose_cmd ps --services))
    local healthy=0
    local total=${#services[@]}
    
    echo
    printf "%-20s %-15s %-15s %s\n" "SERVICE" "STATUS" "HEALTH" "ENDPOINT"
    echo "--------------------------------------------------------------------------------------------------------"
    
    for service in "${services[@]}"; do
        local status=$($compose_cmd ps -q "$service" | xargs -I {} docker inspect --format='{{.State.Status}}' {} 2>/dev/null || echo "not found")
        local health=$($compose_cmd ps -q "$service" | xargs -I {} docker inspect --format='{{.State.Health.Status}}' {} 2>/dev/null || echo "none")
        
        # 确定端点
        local endpoint=""
        case "$service" in
            "api-gateway") endpoint="http://localhost:8080/health" ;;
            "agent") endpoint="http://localhost:8081/health" ;;
            "llm") endpoint="http://localhost:8082/health" ;;
            "mcp") endpoint="http://localhost:8083/health" ;;
            "orchestrator") endpoint="http://localhost:8084/health" ;;
            "rag") endpoint="http://localhost:8085/health" ;;
            "notify") endpoint="http://localhost:8086/health" ;;
            "jaeger") endpoint="http://localhost:16686/" ;;
            *) endpoint="-" ;;
        esac
        
        # 颜色状态显示
        local status_color="$RED"
        if [[ "$status" == "running" ]]; then
            status_color="$GREEN"
            if [[ "$health" == "healthy" ]] || [[ "$health" == "none" && "$status" == "running" ]]; then
                ((healthy++))
            fi
        fi
        
        local health_color="$RED"
        if [[ "$health" == "healthy" ]] || [[ "$health" == "none" ]]; then
            health_color="$GREEN"
        fi
        
        printf "%-20s ${status_color}%-15s${NC} ${health_color}%-15s${NC} %s\n" "$service" "$status" "$health" "$endpoint"
    done
    
    echo
    if [[ $healthy -eq $total ]]; then
        log_success "所有服务运行正常 ($healthy/$total)"
    else
        log_warn "部分服务存在问题 ($healthy/$total)"
    fi
}

# 主函数
main() {
    # 解析参数
    parse_args "$@"
    
    # 检查环境
    check_docker_compose
    check_required_files
    setup_environment
    
    # 执行命令
    case "$COMMAND" in
        up)
            cmd_up
            ;;
        down)
            cmd_down
            ;;
        restart)
            cmd_restart
            ;;
        logs)
            cmd_logs "$@"
            ;;
        ps)
            cmd_ps
            ;;
        pull)
            cmd_pull
            ;;
        clean)
            cmd_clean
            ;;
        health)
            cmd_health
            ;;
        *)
            log_error "未知命令: $COMMAND"
            show_help
            exit 1
            ;;
    esac
}

# 脚本入口
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
