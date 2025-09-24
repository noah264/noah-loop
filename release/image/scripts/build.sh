#!/bin/bash

# Noah-Loop 镜像构建脚本
# 使用方法：
#   ./build.sh [options] [services...]
#
# 选项：
#   -v, --version VERSION    指定版本号（默认：1.0.0）
#   -r, --registry REGISTRY  指定镜像仓库（默认：docker.io）
#   -n, --namespace NS       指定命名空间（默认：noah-loop）
#   -p, --parallel N         并行构建数量（默认：4）
#   --no-cache              不使用缓存构建
#   --push                  构建后推送到仓库
#   -h, --help              显示帮助信息
#
# 示例：
#   ./build.sh                           # 构建所有服务
#   ./build.sh agent llm                 # 只构建 agent 和 llm 服务
#   ./build.sh -v 1.1.0 --push         # 构建版本 1.1.0 并推送
#   ./build.sh --no-cache agent        # 无缓存构建 agent 服务

set -euo pipefail

# 默认配置
VERSION="1.0.0"
REGISTRY="docker.io"
NAMESPACE="noah-loop"
PARALLEL=4
NO_CACHE=""
PUSH=false
SERVICES=()

# 所有可用的服务
ALL_SERVICES=(
    "api-gateway"
    "agent"
    "llm" 
    "mcp"
    "orchestrator"
    "rag"
    "notify"
)

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
Noah-Loop 镜像构建脚本

使用方法：
    $0 [options] [services...]

选项：
    -v, --version VERSION    指定版本号（默认：${VERSION}）
    -r, --registry REGISTRY  指定镜像仓库（默认：${REGISTRY}）
    -n, --namespace NS       指定命名空间（默认：${NAMESPACE}）
    -p, --parallel N         并行构建数量（默认：${PARALLEL}）
    --no-cache              不使用缓存构建
    --push                  构建后推送到仓库
    -h, --help              显示帮助信息

可用服务：
    ${ALL_SERVICES[*]}

示例：
    $0                           # 构建所有服务
    $0 agent llm                 # 只构建 agent 和 llm 服务
    $0 -v 1.1.0 --push         # 构建版本 1.1.0 并推送
    $0 --no-cache agent        # 无缓存构建 agent 服务

EOF
}

# 解析命令行参数
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -r|--registry)
                REGISTRY="$2"
                shift 2
                ;;
            -n|--namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            -p|--parallel)
                PARALLEL="$2"
                shift 2
                ;;
            --no-cache)
                NO_CACHE="--no-cache"
                shift
                ;;
            --push)
                PUSH=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            -*)
                log_error "未知选项: $1"
                show_help
                exit 1
                ;;
            *)
                # 检查是否为有效服务
                if [[ " ${ALL_SERVICES[*]} " =~ " $1 " ]]; then
                    SERVICES+=("$1")
                else
                    log_error "未知服务: $1"
                    log_info "可用服务: ${ALL_SERVICES[*]}"
                    exit 1
                fi
                shift
                ;;
        esac
    done

    # 如果没有指定服务，构建所有服务
    if [[ ${#SERVICES[@]} -eq 0 ]]; then
        SERVICES=("${ALL_SERVICES[@]}")
    fi
}

# 检查 Docker 是否可用
check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装或不可用"
        exit 1
    fi

    if ! docker info &> /dev/null; then
        log_error "Docker 守护进程未运行"
        exit 1
    fi
}

# 检查项目结构
check_project_structure() {
    local required_files=(
        "../../backend/go.work"
        "Dockerfile.api-gateway"
        "docker-compose.build.yml"
    )

    for file in "${required_files[@]}"; do
        if [[ ! -f "$file" ]]; then
            log_error "缺少必需文件: $file"
            exit 1
        fi
    done
}

# 获取构建信息
get_build_info() {
    export BUILD_TIME=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
    if command -v git &> /dev/null && git rev-parse --git-dir > /dev/null 2>&1; then
        export GIT_COMMIT=$(git rev-parse --short HEAD)
        export GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
    else
        export GIT_COMMIT="unknown"
        export GIT_BRANCH="unknown"
    fi
    export VERSION="$VERSION"
}

# 构建单个服务
build_service() {
    local service="$1"
    local image_name="${REGISTRY}/${NAMESPACE}/${service}:${VERSION}"
    local dockerfile="Dockerfile.${service}"
    
    log_info "开始构建服务: ${service}"
    log_info "镜像名称: ${image_name}"
    
    # 构建镜像
    if docker build \
        ${NO_CACHE} \
        -f "${dockerfile}" \
        -t "${image_name}" \
        --build-arg BUILD_TIME="${BUILD_TIME}" \
        --build-arg GIT_COMMIT="${GIT_COMMIT}" \
        --build-arg VERSION="${VERSION}" \
        ../../; then
        
        log_success "构建成功: ${service} -> ${image_name}"
        
        # 推送镜像
        if [[ "$PUSH" == "true" ]]; then
            log_info "推送镜像: ${image_name}"
            if docker push "${image_name}"; then
                log_success "推送成功: ${image_name}"
            else
                log_error "推送失败: ${image_name}"
                return 1
            fi
        fi
        
        return 0
    else
        log_error "构建失败: ${service}"
        return 1
    fi
}

# 并行构建服务
build_services_parallel() {
    local services=("$@")
    local pids=()
    local results=()
    local active_jobs=0
    
    log_info "开始并行构建 ${#services[@]} 个服务（并行数：${PARALLEL}）"
    
    for service in "${services[@]}"; do
        # 等待空闲槽位
        while [[ $active_jobs -ge $PARALLEL ]]; do
            # 检查已完成的任务
            for i in "${!pids[@]}"; do
                if [[ -n "${pids[i]}" ]] && ! kill -0 "${pids[i]}" 2>/dev/null; then
                    wait "${pids[i]}"
                    results[i]=$?
                    unset pids[i]
                    ((active_jobs--))
                fi
            done
            sleep 0.1
        done
        
        # 启动新任务
        build_service "$service" &
        local pid=$!
        pids+=("$pid")
        ((active_jobs++))
        
        log_info "启动构建任务: ${service} (PID: ${pid})"
    done
    
    # 等待所有任务完成
    log_info "等待所有构建任务完成..."
    for pid in "${pids[@]}"; do
        if [[ -n "$pid" ]]; then
            wait "$pid"
            results+=($?)
        fi
    done
    
    # 检查结果
    local failed=0
    for i in "${!services[@]}"; do
        if [[ ${results[i]:-1} -ne 0 ]]; then
            log_error "服务构建失败: ${services[i]}"
            ((failed++))
        fi
    done
    
    if [[ $failed -eq 0 ]]; then
        log_success "所有服务构建成功！"
        return 0
    else
        log_error "有 ${failed} 个服务构建失败"
        return 1
    fi
}

# 显示构建摘要
show_summary() {
    echo
    log_info "=============== 构建摘要 ==============="
    log_info "版本: ${VERSION}"
    log_info "仓库: ${REGISTRY}/${NAMESPACE}"
    log_info "构建时间: ${BUILD_TIME}"
    log_info "Git提交: ${GIT_COMMIT}"
    log_info "服务列表: ${SERVICES[*]}"
    log_info "并行数: ${PARALLEL}"
    [[ -n "$NO_CACHE" ]] && log_info "缓存: 禁用"
    [[ "$PUSH" == "true" ]] && log_info "推送: 启用"
    log_info "========================================"
    echo
}

# 清理函数
cleanup() {
    log_warn "收到中断信号，正在清理..."
    # 杀死所有子进程
    jobs -p | xargs -r kill
    exit 130
}

# 主函数
main() {
    # 设置中断处理
    trap cleanup SIGINT SIGTERM
    
    # 解析参数
    parse_args "$@"
    
    # 检查环境
    check_docker
    check_project_structure
    
    # 获取构建信息
    get_build_info
    
    # 显示摘要
    show_summary
    
    # 确认继续
    if [[ -t 0 ]] && [[ "${CI:-}" != "true" ]]; then
        echo -n "确认开始构建？[y/N] "
        read -r response
        if [[ ! "$response" =~ ^[Yy]$ ]]; then
            log_info "构建已取消"
            exit 0
        fi
    fi
    
    echo
    log_info "开始构建..."
    local start_time=$(date +%s)
    
    # 构建服务
    if build_services_parallel "${SERVICES[@]}"; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        log_success "所有构建任务完成！用时: ${duration}s"
        
        echo
        log_info "构建的镜像："
        for service in "${SERVICES[@]}"; do
            echo "  ${REGISTRY}/${NAMESPACE}/${service}:${VERSION}"
        done
        
        exit 0
    else
        log_error "构建失败！"
        exit 1
    fi
}

# 脚本入口
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
