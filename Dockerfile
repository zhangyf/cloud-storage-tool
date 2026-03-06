# Cloud Storage Tool Dockerfile

# 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 设置构建参数
ARG VERSION=0.1.0
ARG BUILD_DATE
ARG GIT_COMMIT

# 构建应用程序
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s \
    -X 'main.buildDate=${BUILD_DATE}' \
    -X 'main.gitCommit=${GIT_COMMIT}'" \
    -o cloud-storage ./cmd/cloud-storage

# 运行阶段
FROM alpine:latest

# 安装必要的工具
RUN apk --no-cache add ca-certificates tzdata

# 创建非root用户
RUN addgroup -g 1000 cloudstorage && \
    adduser -u 1000 -G cloudstorage -s /bin/sh -D cloudstorage

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/cloud-storage .

# 复制配置文件模板
COPY --from=builder /app/config.example.yaml .

# 设置权限
RUN chown -R cloudstorage:cloudstorage /app && \
    chmod +x cloud-storage

# 切换到非root用户
USER cloudstorage

# 创建配置目录
RUN mkdir -p /home/cloudstorage/.cloud-storage

# 设置环境变量
ENV PATH="/app:${PATH}"

# 设置入口点
ENTRYPOINT ["cloud-storage"]

# 设置默认命令
CMD ["--help"]