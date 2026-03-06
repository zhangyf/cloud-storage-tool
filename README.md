# Cloud Storage Tool

一个统一的云存储工具，支持多种云存储服务。

## 功能特性

- ✅ 支持腾讯云 COS
- ✅ 支持阿里云 OSS
- ✅ 支持 AWS S3
- ✅ 统一的 API 接口
- ✅ 命令行工具
- ✅ 配置文件管理
- ✅ 完整的错误处理
- ✅ 详细的日志系统

## 架构设计

### 核心模块
1. **config** - 配置管理
2. **storage** - 存储接口抽象
3. **providers** - 云存储提供商实现
4. **cli** - 命令行工具
5. **utils** - 工具函数

### 接口设计
```go
type StorageProvider interface {
    Upload(localPath, remotePath string) error
    Download(remotePath, localPath string) error
    List(prefix string) ([]FileInfo, error)
    Delete(path string) error
    Stat(path string) (FileInfo, error)
}
```

## 快速开始

```bash
# 安装
go install ./cmd/cloud-storage

# 配置
cloud-storage config init

# 上传文件
cloud-storage upload ./local/file.txt remote/path/file.txt

# 下载文件
cloud-storage download remote/path/file.txt ./local/file.txt
```

## 配置文件示例

```yaml
default_provider: "tencent_cos"
providers:
  tencent_cos:
    type: "tencent_cos"
    bucket: "your-bucket"
    region: "ap-beijing"
    secret_id: "your-secret-id"
    secret_key: "your-secret-key"
  
  aliyun_oss:
    type: "aliyun_oss"
    bucket: "your-bucket"
    endpoint: "oss-cn-beijing.aliyuncs.com"
    access_key_id: "your-access-key-id"
    access_key_secret: "your-access-key-secret"
  
  aws_s3:
    type: "aws_s3"
    bucket: "your-bucket"
    region: "us-east-1"
    access_key_id: "your-access-key-id"
    secret_access_key: "your-secret-access-key"
```

## 开发计划

### 第一阶段：基础架构和配置 ✅ 已完成
- [x] 项目结构设计
- [x] 配置管理模块
- [x] 存储接口抽象
- [x] 基础命令行框架
- [x] 错误处理系统
- [x] 日志系统
- [x] Makefile 和 Dockerfile

### 第二阶段：提供商实现 🔄 进行中
- [x] 腾讯云 COS 实现
- [x] 阿里云 OSS 实现（简化版，待完善）
- [x] AWS S3 实现
- [ ] 提供商工厂模式
- [ ] 配置验证和测试

### 第三阶段：高级功能 📋 待开发
- [ ] 断点续传
- [ ] 并行上传/下载
- [ ] 进度显示
- [ ] 文件校验
- [ ] 单元测试
- [ ] 集成测试
- [ ] 性能优化

## 技术栈
- Go 1.21+
- 标准库优先
- YAML 配置文件
- Cobra CLI 框架