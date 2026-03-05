# Cloud Storage Tool

统一的多云存储管理工具，支持腾讯云COS和AWS S3。

## 特性

- **统一接口**：一套命令管理多个云存储服务
- **完整功能**：桶管理、对象操作、同步迁移
- **高性能**：多线程并发、断点续传、大文件支持
- **易用性**：类似aws s3、coscmd的熟悉语法
- **安全可靠**：完善的错误处理、数据校验、审计日志

## 支持的服务

- ✅ 腾讯云COS
- ✅ AWS S3
- 🔄 阿里云OSS（规划中）
- 🔄 华为云OBS（规划中）
- 🔄 Google Cloud Storage（规划中）

## 快速开始

### 安装

```bash
# 从源码编译
git clone https://github.com/zhangyf/cloud-storage-tool.git
cd cloud-storage-tool
go build -o cloud-storage-tool ./cmd/cloud-storage-tool
sudo mv cloud-storage-tool /usr/local/bin/
```

### 配置

```bash
# 初始化配置
cloud-storage-tool config init

# 测试配置
cloud-storage-tool config test
```

### 基本使用

```bash
# 列出所有桶
cloud-storage-tool cos ls
cloud-storage-tool s3 ls

# 上传文件
cloud-storage-tool cos cp local/file.txt cos://my-bucket/path/
cloud-storage-tool s3 cp local/file.txt s3://my-bucket/path/

# 下载文件
cloud-storage-tool cos cp cos://bucket/file.txt local/path/
cloud-storage-tool s3 cp s3://bucket/file.txt local/path/

# 同步目录
cloud-storage-tool sync local/dir/ cos://bucket/path/
cloud-storage-tool sync cos://bucket1/path/ s3://bucket2/path/
```

## 详细文档

- [产品需求文档 (PRD)](docs/cloud-storage-tool-prd.md) - 完整的功能需求和设计
- [架构设计](docs/architecture.md) - 技术架构和实现细节
- [API参考](docs/api-reference.md) - 完整的命令参考
- [开发指南](docs/development.md) - 贡献和开发指南

## 项目结构

```
cloud-storage-tool/
├── cmd/                    # 命令行入口
├── internal/              # 内部包
│   ├── provider/          # 云服务提供商实现
│   ├── commands/          # 命令实现
│   ├── config/            # 配置管理
│   └── utils/             # 工具函数
├── pkg/                   # 公共包
├── config/                # 配置文件示例
├── scripts/               # 构建和部署脚本
├── tests/                 # 测试文件
└── docs/                  # 文档
```

## 开发状态

| 模块 | 状态 | 进度 |
|------|------|------|
| 基础架构 | 🔄 规划中 | 0% |
| COS提供商 | 🔄 规划中 | 0% |
| S3提供商 | 🔄 规划中 | 0% |
| 桶管理命令 | 🔄 规划中 | 0% |
| 对象操作命令 | 🔄 规划中 | 0% |
| 同步功能 | 🔄 规划中 | 0% |

## 技术栈

- **语言**: Go 1.21+
- **云服务SDK**:
  - 腾讯云COS: `github.com/tencentyun/cos-go-sdk-v5`
  - AWS S3: `github.com/aws/aws-sdk-go-v2/service/s3`
- **命令行框架**: Cobra + Viper
- **配置文件**: YAML
- **测试框架**: Go test

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request！

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 打开Pull Request

## 联系方式

- 项目维护者: zhangyf
- GitHub: [zhangyf](https://github.com/zhangyf)
- 问题跟踪: [GitHub Issues](https://github.com/zhangyf/cloud-storage-tool/issues)