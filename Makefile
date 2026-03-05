# Cloud Storage Tool Makefile

.PHONY: all build clean test lint format deps help

# 项目信息
PROJECT_NAME := cloud-storage-tool
VERSION := 0.1.0
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Go参数
GO := go
GOFLAGS :=
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_DATE)

# 目录
DIST_DIR := dist
BIN_DIR := $(DIST_DIR)/bin
COVERAGE_FILE := coverage.out
COVERAGE_HTML := coverage.html

# 默认目标
all: build

# 显示帮助
help:
	@echo "Cloud Storage Tool 构建系统"
	@echo ""
	@echo "可用目标:"
	@echo "  make build       构建项目 (默认)"
	@echo "  make debug       构建调试版本"
	@echo "  make release     构建发布版本"
	@echo "  make install     安装到系统"
	@echo "  make clean       清理构建文件"
	@echo "  make test        运行测试"
	@echo "  make test-cover  运行测试并生成覆盖率报告"
	@echo "  make lint        代码检查"
	@echo "  make format      代码格式化"
	@echo "  make deps        下载依赖"
	@echo "  make help        显示此帮助"
	@echo ""
	@echo "示例:"
	@echo "  make debug       构建调试版本"
	@echo "  make test-cover  运行测试并查看覆盖率"

# 构建开发版本
build: deps
	@echo "🚀 构建 $(PROJECT_NAME)..."
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(PROJECT_NAME) ./cmd/$(PROJECT_NAME)
	@echo "✅ 构建完成: $(BIN_DIR)/$(PROJECT_NAME)"

# 构建调试版本
debug: GOFLAGS += -gcflags="all=-N -l"
debug: LDFLAGS += -X main.build=debug
debug: build
	@echo "🔧 调试版本构建完成"

# 构建发布版本
release: LDFLAGS += -X main.build=release -s -w
release: GOFLAGS += -trimpath
release: build
	@echo "📦 发布版本构建完成"

# 安装到系统
install: build
	@echo "📥 安装到系统..."
	@sudo cp $(BIN_DIR)/$(PROJECT_NAME) /usr/local/bin/$(PROJECT_NAME)
	@echo "✅ 安装完成: /usr/local/bin/$(PROJECT_NAME)"

# 清理构建文件
clean:
	@echo "🧹 清理构建文件..."
	@rm -rf $(DIST_DIR)
	@rm -f $(COVERAGE_FILE) $(COVERAGE_HTML)
	@echo "✅ 清理完成"

# 运行测试
test: deps
	@echo "🧪 运行测试..."
	$(GO) test ./... -v

# 运行测试并生成覆盖率报告
test-cover: deps
	@echo "🧪 运行测试并生成覆盖率报告..."
	$(GO) test ./... -v -cover -coverprofile=$(COVERAGE_FILE)
	$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML)
	@echo "📊 覆盖率报告: $(COVERAGE_HTML)"

# 代码检查
lint: deps
	@echo "🔍 代码检查..."
	$(GO) vet ./...
	@echo "✅ 代码检查完成"

# 代码格式化
format:
	@echo "🎨 代码格式化..."
	$(GO) fmt ./...
	@echo "✅ 代码格式化完成"

# 下载依赖
deps:
	@echo "📦 下载依赖..."
	$(GO) mod download
	@echo "✅ 依赖下载完成"

# 交叉编译
cross-build:
	@echo "🌍 交叉编译..."
	@mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(PROJECT_NAME)-linux-amd64 ./cmd/$(PROJECT_NAME)
	GOOS=linux GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(PROJECT_NAME)-linux-arm64 ./cmd/$(PROJECT_NAME)
	GOOS=darwin GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(PROJECT_NAME)-darwin-amd64 ./cmd/$(PROJECT_NAME)
	GOOS=darwin GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(PROJECT_NAME)-darwin-arm64 ./cmd/$(PROJECT_NAME)
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(PROJECT_NAME)-windows-amd64.exe ./cmd/$(PROJECT_NAME)
	@echo "✅ 交叉编译完成，文件在 $(DIST_DIR)/ 目录"

# 版本信息
version:
	@echo "版本: $(VERSION)"
	@echo "提交: $(COMMIT)"
	@echo "构建时间: $(BUILD_DATE)"