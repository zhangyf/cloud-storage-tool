#!/bin/bash
# Cloud Storage Tool 构建脚本

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示帮助
show_help() {
    echo "用法: $0 [选项]"
    echo ""
    echo "选项:"
    echo "  -h, --help          显示此帮助信息"
    echo "  -d, --debug         构建调试版本"
    echo "  -r, --release       构建发布版本"
    echo "  -t, --test          运行测试"
    echo "  -c, --clean         清理构建文件"
    echo "  -p, --platform      指定目标平台 (格式: OS/ARCH)"
    echo ""
    echo "示例:"
    echo "  $0 --debug          构建调试版本"
    echo "  $0 --release        构建发布版本"
    echo "  $0 --clean          清理构建文件"
    echo "  $0 --test           运行所有测试"
    echo "  $0 --platform linux/amd64 构建Linux 64位版本"
}

# 默认参数
DEBUG=false
RELEASE=false
TEST=false
CLEAN=false
PLATFORM=""

# 解析参数
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -d|--debug)
            DEBUG=true
            shift
            ;;
        -r|--release)
            RELEASE=true
            shift
            ;;
        -t|--test)
            TEST=true
            shift
            ;;
        -c|--clean)
            CLEAN=true
            shift
            ;;
        -p|--platform)
            PLATFORM="$2"
            shift 2
            ;;
        -*)
            log_error "未知选项: $1"
            show_help
            exit 1
            ;;
        *)
            log_error "未知参数: $1"
            show_help
            exit 1
            ;;
    esac
done

# 项目信息
PROJECT_NAME="cloud-storage-tool"
VERSION="0.1.0"
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# 清理构建文件
if [ "$CLEAN" = true ]; then
    log_info "清理构建文件..."
    rm -rf ./${PROJECT_NAME}
    rm -rf ./dist/
    rm -f ./coverage.out
    rm -f ./coverage.html
    log_success "清理完成"
    exit 0
fi

# 运行测试
if [ "$TEST" = true ]; then
    log_info "运行测试..."
    go test ./... -v -cover -coverprofile=coverage.out
    if [ $? -eq 0 ]; then
        go tool cover -html=coverage.out -o coverage.html
        log_success "测试完成，覆盖率报告已生成: coverage.html"
    else
        log_error "测试失败"
        exit 1
    fi
    exit 0
fi

# 构建参数
LDFLAGS="-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${BUILD_DATE}"
BUILD_FLAGS=""

if [ "$DEBUG" = true ]; then
    log_info "构建调试版本..."
    LDFLAGS="${LDFLAGS} -X main.build=debug"
    BUILD_FLAGS="-gcflags=\"all=-N -l\""
elif [ "$RELEASE" = true ]; then
    log_info "构建发布版本..."
    LDFLAGS="${LDFLAGS} -X main.build=release -s -w"
    BUILD_FLAGS="-trimpath"
else
    log_info "构建开发版本..."
    LDFLAGS="${LDFLAGS} -X main.build=dev"
fi

# 检查Go版本
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.21"
if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    log_error "需要Go版本 >= ${REQUIRED_VERSION}，当前版本: ${GO_VERSION}"
    exit 1
fi

# 下载依赖
log_info "下载依赖..."
go mod download
if [ $? -ne 0 ]; then
    log_error "依赖下载失败"
    exit 1
fi

# 构建
log_info "开始构建..."
mkdir -p ./dist

if [ -n "$PLATFORM" ]; then
    # 交叉编译
    OS=$(echo $PLATFORM | cut -d'/' -f1)
    ARCH=$(echo $PLATFORM | cut -d'/' -f2)
    OUTPUT_NAME="${PROJECT_NAME}-${OS}-${ARCH}"
    
    log_info "交叉编译: ${OS}/${ARCH}"
    GOOS=$OS GOARCH=$ARCH go build ${BUILD_FLAGS} -ldflags "${LDFLAGS}" \
        -o "./dist/${OUTPUT_NAME}" ./cmd/${PROJECT_NAME}
else
    # 本地编译
    OUTPUT_NAME="${PROJECT_NAME}"
    go build ${BUILD_FLAGS} -ldflags "${LDFLAGS}" \
        -o "./dist/${OUTPUT_NAME}" ./cmd/${PROJECT_NAME}
fi

if [ $? -eq 0 ]; then
    # 检查文件大小
    FILE_SIZE=$(stat -f%z "./dist/${OUTPUT_NAME}" 2>/dev/null || stat -c%s "./dist/${OUTPUT_NAME}" 2>/dev/null)
    FILE_SIZE_MB=$(echo "scale=2; $FILE_SIZE / 1024 / 1024" | bc)
    
    log_success "构建成功!"
    log_info "输出文件: ./dist/${OUTPUT_NAME}"
    log_info "文件大小: ${FILE_SIZE_MB} MB"
    log_info "版本信息: ${VERSION} (${COMMIT})"
    
    # 显示构建信息
    echo ""
    echo "🎉 构建完成!"
    echo "   版本: ${VERSION} (${COMMIT})"
    echo "   构建时间: ${BUILD_DATE}"
    echo "   文件: ./dist/${OUTPUT_NAME}"
    echo "   大小: ${FILE_SIZE_MB} MB"
    echo ""
    echo "💡 使用提示:"
    echo "   ./dist/${OUTPUT_NAME} --help"
else
    log_error "构建失败"
    exit 1
fi