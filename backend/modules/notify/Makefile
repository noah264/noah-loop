# Notify Module Makefile

.PHONY: deps generate proto-gen wire-gen build run dev test lint clean install-tools docker

# 变量定义
MODULE_NAME := notify
BINARY_NAME := notify-service
PKG_PATH := github.com/noah-loop/backend/modules/notify

# 依赖管理
deps:
	@echo "Installing dependencies..."
	@if command -v go > /dev/null 2>&1; then \
		go mod download; \
		go mod tidy; \
	else \
		echo "Go is not installed or not in PATH"; \
	fi

# 生成代码
generate: proto-gen wire-gen

# Protobuf生成
proto-gen:
	@echo "Generating protobuf files..."
	@if command -v protoc > /dev/null 2>&1; then \
		protoc --go_out=. --go_opt=paths=source_relative \
		       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
		       ../../api-gateway/proto/notify/v1/*.proto; \
	else \
		echo "protoc not found. Please install Protocol Buffers compiler."; \
		echo "Visit: https://grpc.io/docs/protoc-installation/"; \
	fi

# Wire生成
wire-gen:
	@echo "Generating wire files..."
	@if command -v wire > /dev/null 2>&1; then \
		cd internal/wire && wire; \
	else \
		echo "Wire not found. Installing..."; \
		if command -v go > /dev/null 2>&1; then \
			go install github.com/google/wire/cmd/wire@latest; \
			cd internal/wire && wire; \
		else \
			echo "Go is not installed. Cannot install wire."; \
		fi \
	fi

# 构建
build: generate
	@echo "Building $(BINARY_NAME)..."
	@if command -v go > /dev/null 2>&1; then \
		CGO_ENABLED=0 go build -o bin/$(BINARY_NAME) ./cmd/main.go; \
	else \
		echo "Go is not installed or not in PATH"; \
	fi

# 运行
run: build
	@echo "Running $(BINARY_NAME)..."
	@./bin/$(BINARY_NAME)

# 开发模式运行
dev: generate
	@echo "Running $(BINARY_NAME) in development mode..."
	@if command -v go > /dev/null 2>&1; then \
		go run ./cmd/main.go; \
	else \
		echo "Go is not installed or not in PATH"; \
	fi

# 测试
test:
	@echo "Running tests..."
	@if command -v go > /dev/null 2>&1; then \
		go test -v ./...; \
	else \
		echo "Go is not installed or not in PATH"; \
	fi

# 测试覆盖率
test-coverage:
	@echo "Running tests with coverage..."
	@if command -v go > /dev/null 2>&1; then \
		go test -v -race -coverprofile=coverage.out ./...; \
		go tool cover -html=coverage.out -o coverage.html; \
		echo "Coverage report generated: coverage.html"; \
	else \
		echo "Go is not installed or not in PATH"; \
	fi

# 代码检查
lint:
	@echo "Running linters..."
	@if command -v golangci-lint > /dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found. Please install it: https://golangci-lint.run/usage/install/"; \
		if command -v go > /dev/null 2>&1; then \
			echo "Running basic checks..."; \
			go fmt ./...; \
			go vet ./...; \
		fi \
	fi

# 清理
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@if [ -f internal/wire/wire_gen.go ]; then rm internal/wire/wire_gen.go; fi

# 安装工具
install-tools:
	@echo "Installing development tools..."
	@if command -v go > /dev/null 2>&1; then \
		echo "Installing Wire..."; \
		go install github.com/google/wire/cmd/wire@latest; \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		echo "Installing protoc-gen-go..."; \
		go install google.golang.org/protobuf/cmd/protoc-gen-go@latest; \
		echo "Installing protoc-gen-go-grpc..."; \
		go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest; \
		echo "Tools installed successfully!"; \
	else \
		echo "Go is not installed. Please install Go first."; \
	fi

# Docker构建
docker:
	@echo "Building Docker image..."
	@docker build -t noah-loop/$(BINARY_NAME) .

# Docker运行
docker-run:
	@echo "Running Docker container..."
	@docker run -p 8086:8086 -p 9096:9096 noah-loop/$(BINARY_NAME)

# 格式化代码
fmt:
	@echo "Formatting code..."
	@if command -v go > /dev/null 2>&1; then \
		go fmt ./...; \
	else \
		echo "Go is not installed or not in PATH"; \
	fi

# 更新依赖
update:
	@echo "Updating dependencies..."
	@if command -v go > /dev/null 2>&1; then \
		go get -u ./...; \
		go mod tidy; \
	else \
		echo "Go is not installed or not in PATH"; \
	fi

# 检查漏洞
security:
	@echo "Checking for security vulnerabilities..."
	@if command -v govulncheck > /dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not found. Installing..."; \
		if command -v go > /dev/null 2>&1; then \
			go install golang.org/x/vuln/cmd/govulncheck@latest; \
			govulncheck ./...; \
		else \
			echo "Go is not installed. Cannot install govulncheck."; \
		fi \
	fi

# 帮助信息
help:
	@echo "Available targets:"
	@echo "  deps          - Install dependencies"
	@echo "  generate      - Generate all code (protobuf + wire)"
	@echo "  proto-gen     - Generate protobuf files"
	@echo "  wire-gen      - Generate wire files"
	@echo "  build         - Build the binary"
	@echo "  run           - Build and run the service"
	@echo "  dev           - Run in development mode"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  lint          - Run linters"
	@echo "  fmt           - Format code"
	@echo "  clean         - Clean build artifacts"
	@echo "  install-tools - Install development tools"
	@echo "  docker        - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  update        - Update dependencies"
	@echo "  security      - Check for security vulnerabilities"
	@echo "  help          - Show this help"

# 默认目标
.DEFAULT_GOAL := help
