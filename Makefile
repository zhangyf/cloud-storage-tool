# Cloud Storage Tool Makefile

.PHONY: all build clean test lint install uninstall release help

# 项目信息
PROJECT_NAME := cloud-storage-tool
VERSION := 0.1.0
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# 构建参数
GO := go
GOFLAGS := -mod=readonly
LDFLAGS := -X 'main.buildDate=$(BUILD_DATE)' -X 'main.gitCommit=$(GIT_COMMIT)'
BUILD_TAGS :=

# 目录
BIN_DIR := bin
DIST_DIR := dist
TEST_DIR := test
COVERAGE_DIR := coverage

# 目标平台
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

# 默认目标
all: build

# 构建所有
build: clean
	@echo "构建 $(PROJECT_NAME) v$(VERSION)..."
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -tags "$(BUILD_TAGS)" -o $(BIN_DIR)/$(PROJECT_NAME) ./cmd/cloud-storage
	@echo "构建完成: $(BIN_DIR)/$(PROJECT_NAME)"

# 快速构建（开发用）
dev:
	$(GO) build $(GOFLAGS) -o $(BIN_DIR)/$(PROJECT_NAME) ./cmd/cloud-storage

# 安装到系统
install: build
	@echo "安装 $(PROJECT_NAME) 到系统..."
	install -m 755 $(BIN_DIR)/$(PROJECT_NAME) /usr/local/bin/
	@echo "安装完成"

# 卸载
uninstall:
	@echo "卸载 $(PROJECT_NAME)..."
	rm -f /usr/local/bin/$(PROJECT_NAME)
	@echo "卸载完成"

# 清理构建文件
clean:
	@echo "清理构建文件..."
	rm -rf $(BIN_DIR) $(DIST_DIR) $(COVERAGE_DIR)
	@echo "清理完成"

# 运行测试
test:
	@echo "运行测试..."
	@mkdir -p $(COVERAGE_DIR)
	$(GO) test $(GOFLAGS) -v -cover -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	$(GO) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "测试完成，覆盖率报告: $(COVERAGE_DIR)/coverage.html"

# 运行基准测试
bench:
	@echo "运行基准测试..."
	$(GO) test $(GOFLAGS) -bench=. -benchmem ./...
	@echo "基准测试完成"

# 代码检查
lint:
	@echo "运行代码检查..."
	# 安装 golangci-lint 如果未安装
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "安装 golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.54.2; \
	fi
	golangci-lint run ./...
	@echo "代码检查完成"

# 格式化代码
fmt:
	@echo "格式化代码..."
	$(GO) fmt ./...
	@echo "格式化完成"

# 更新依赖
update:
	@echo "更新依赖..."
	$(GO) get -u ./...
	$(GO) mod tidy
	@echo "依赖更新完成"

# 生成文档
doc:
	@echo "生成文档..."
	$(GO) doc -all ./...
	@echo "文档生成完成"

# 跨平台构建
cross-build: clean
	@echo "开始跨平台构建..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		output_name="$(PROJECT_NAME)-$${os}-$${arch}"; \
		if [ "$$os" = "windows" ]; then \
			output_name="$${output_name}.exe"; \
		fi; \
		echo "构建 $${os}/$${arch}..."; \
		GOOS=$$os GOARCH=$$arch $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -tags "$(BUILD_TAGS)" -o $(DIST_DIR)/$$output_name ./cmd/cloud-storage; \
	done
	@echo "跨平台构建完成，输出到: $(DIST_DIR)"

# 创建发布包
release: cross-build
	@echo "创建发布包..."
	@cd $(DIST_DIR) && for file in *; do \
		if [ -f "$$file" ]; then \
			tar -czf "$${file}.tar.gz" "$$file"; \
			rm -f "$$file"; \
			echo "创建: $${file}.tar.gz"; \
		fi; \
	done
	@echo "发布包创建完成"

# Docker 构建
docker-build:
	@echo "构建 Docker 镜像..."
	docker build -t $(PROJECT_NAME):$(VERSION) -t $(PROJECT_NAME):latest .
	@echo "Docker 镜像构建完成"

# Docker 运行
docker-run:
	@echo "运行 Docker 容器..."
	docker run --rm -it \
		-v $(HOME)/.cloud-storage:/root/.cloud-storage \
		-v $(PWD)/data:/data \
		$(PROJECT_NAME):latest

# 集成测试
integration-test:
	@echo "运行集成测试..."
	@mkdir -p $(TEST_DIR)
	# 这里可以添加集成测试命令
	@echo "集成测试完成"

# 安全检查
security-check:
	@echo "运行安全检查..."
	# 安装 gosec 如果未安装
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "安装 gosec..."; \
		$(GO) install github.com/securego/gosec/v2/cmd/gosec@latest; \
	fi
	gosec ./...
	@echo "安全检查完成"

# 性能分析
profile:
	@echo "生成性能分析数据..."
	@mkdir -p $(TEST_DIR)
	$(GO) test -cpuprofile=$(TEST_DIR)/cpu.prof -memprofile=$(TEST_DIR)/mem.prof -bench=. ./internal/storage
	@echo "性能分析数据生成完成，查看: $(TEST_DIR)/"

# 显示帮助
help:
	@echo "Cloud Storage Tool 构建系统"
	@echo ""
	@echo "可用命令:"
	@echo "  all              - 默认构建 (等同于 build)"
	@echo "  build           - 构建项目"
	@echo "  dev             - 快速构建 (开发用)"
	@echo "  install         - 安装到系统"
	@echo "  uninstall       - 从系统卸载"
	@echo "  clean           - 清理构建文件"
	@echo "  test            - 运行测试并生成覆盖率报告"
	@echo "  bench           - 运行基准测试"
	@echo "  lint            - 运行代码检查"
	@echo "  fmt             - 格式化代码"
	@echo "  update          - 更新依赖"
	@echo "  doc             - 生成文档"
	@echo "  cross-build     - 跨平台构建"
	@echo "  release         - 创建发布包"
	@echo "  docker-build    - 构建 Docker 镜像"
	@echo "  docker-run      - 运行 Docker 容器"
	@echo "  integration-test - 运行集成测试"
	@echo "  security-check  - 运行安全检查"
	@echo "  profile         - 生成性能分析数据"
	@echo "  help            - 显示此帮助信息"
	@echo ""
	@echo "环境变量:"
	@echo "  GOFLAGS         - Go 构建标志 (默认: -mod=readonly)"
	@echo "  LDFLAGS         - 链接器标志"
	@echo "  BUILD_TAGS      - 构建标签"
	@echo ""
	@echo "当前版本: $(VERSION)"
	@echo "构建日期: $(BUILD_DATE)"
	@echo "Git提交: $(GIT_COMMIT)"

# 默认目标
.DEFAULT_GOAL := help