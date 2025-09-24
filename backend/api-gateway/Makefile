# API Gateway Makefile

# Go相关变量
BINARY_NAME=api-gateway
MAIN_PATH=./cmd
BUILD_DIR=./build
WIRE_PATH=./internal/wire
PROTO_PATH=./proto

# 默认目标
.PHONY: all
all: clean deps generate build

# 安装依赖
.PHONY: deps
deps:
	go mod download
	go mod tidy

# 生成Protobuf和Wire代码
.PHONY: generate
generate: proto-gen wire-gen

# 生成Protobuf代码
.PHONY: proto-gen
proto-gen:
	@echo "Generating Protobuf code..."
	@if command -v protoc > /dev/null; then \
		protoc --proto_path=$(PROTO_PATH) \
			--go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
			$(PROTO_PATH)/**/*.proto; \
	else \
		echo "protoc not found. Please install Protocol Buffers compiler."; \
		echo "See: https://grpc.io/docs/protoc-installation/"; \
	fi

# 生成Wire依赖注入代码
.PHONY: wire-gen
wire-gen:
	@echo "Generating Wire dependencies..."
	wire $(WIRE_PATH)

# 构建
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# 运行
.PHONY: run
run: generate
	@echo "Running $(BINARY_NAME)..."
	go run $(MAIN_PATH)/main.go

# 开发运行（带热重载监控）
.PHONY: dev
dev:
	@echo "Running in development mode..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not found. Installing..."; \
		go install github.com/cosmtrek/air@latest; \
		air; \
	fi

# 测试
.PHONY: test
test:
	go test -v ./...

# 测试覆盖率
.PHONY: test-coverage
test-coverage:
	go test -v -cover ./...

# 清理
.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	go clean

# 清理生成的代码
.PHONY: clean-gen
clean-gen:
	@echo "Cleaning generated code..."
	find $(PROTO_PATH) -name "*.pb.go" -delete
	find $(PROTO_PATH) -name "*_grpc.pb.go" -delete

# 格式化代码
.PHONY: fmt
fmt:
	go fmt ./...

# 检查代码
.PHONY: lint
lint:
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Please install it first."; \
		echo "go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# 安装开发工具
.PHONY: install-tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/cosmtrek/air@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# 安装Protobuf编译器（系统级）
.PHONY: install-protoc
install-protoc:
	@echo "Please install protoc manually:"
	@echo "  macOS: brew install protobuf"
	@echo "  Ubuntu: apt install protobuf-compiler"
	@echo "  Windows: Download from https://github.com/protocolbuffers/protobuf/releases"

# Docker相关
.PHONY: docker-build
docker-build:
	docker build -t noah-loop/api-gateway:latest .

.PHONY: docker-run
docker-run:
	docker run --rm -p 8080:8080 noah-loop/api-gateway:latest

# 帮助
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  all              - Clean, install deps, generate code and build"
	@echo "  deps             - Install Go dependencies"
	@echo "  generate         - Generate Protobuf and Wire code"
	@echo "  proto-gen        - Generate Protobuf code only"
	@echo "  wire-gen         - Generate Wire dependency injection code only"
	@echo "  build            - Build the binary"
	@echo "  run              - Generate and run the application"
	@echo "  dev              - Run with hot reload (requires air)"
	@echo "  test             - Run tests"
	@echo "  test-coverage    - Run tests with coverage"
	@echo "  clean            - Clean build files"
	@echo "  clean-gen        - Clean generated code files"
	@echo "  fmt              - Format code"
	@echo "  lint             - Run linter"
	@echo "  install-tools    - Install development tools"
	@echo "  install-protoc   - Show protoc installation instructions"
	@echo "  docker-build     - Build Docker image"
	@echo "  docker-run       - Run Docker container"
	@echo "  help             - Show this help"
