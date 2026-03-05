# 贡献指南

感谢您有兴趣为 Cloud Storage Tool 项目做出贡献！以下是贡献代码的指南。

## 行为准则

请阅读并遵守我们的 [行为准则](CODE_OF_CONDUCT.md)。

## 如何贡献

### 报告问题
- 使用 GitHub Issues 报告问题
- 描述清晰的问题复现步骤
- 包括环境信息（操作系统、Go版本等）
- 如果是错误报告，包括日志和堆栈跟踪

### 功能请求
- 描述您需要的功能
- 解释为什么这个功能有用
- 如果可能，提供使用场景示例

### 提交代码
1. Fork 项目仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 打开 Pull Request

## 开发环境

### 前提条件
- Go 1.21 或更高版本
- Git
- 可选的：Docker（用于测试）

### 设置项目
```bash
# 克隆项目
git clone https://github.com/zhangyf/cloud-storage-tool.git
cd cloud-storage-tool

# 下载依赖
go mod download

# 构建项目
go build ./cmd/cloud-storage-tool

# 运行测试
go test ./...
```

### 项目结构
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
├── scripts/               # 构建脚本
├── tests/                 # 测试
└── docs/                  # 文档
```

## 代码规范

### Go 代码规范
- 遵循 [Go 官方代码规范](https://golang.org/doc/effective_go)
- 使用 `go fmt` 格式化代码
- 使用 `go vet` 检查代码
- 提交前运行测试

### 提交信息规范
使用 [约定式提交](https://www.conventionalcommits.org/) 格式：
```
<类型>[可选的作用域]: <描述>

[可选的正文]

[可选的脚注]
```

类型包括：
- `feat`: 新功能
- `fix`: 错误修复
- `docs`: 文档更新
- `style`: 代码格式调整
- `refactor`: 代码重构
- `test`: 测试相关
- `chore`: 构建过程或辅助工具的变动

### 测试要求
- 新功能必须包含测试
- 错误修复必须包含回归测试
- 测试覆盖率不应降低

## 添加新的存储提供商

### 实现步骤
1. 在 `internal/provider/` 目录创建新文件
2. 实现 `StorageProvider` 接口
3. 添加提供商特定的配置
4. 添加测试
5. 更新文档

### 接口示例
```go
type StorageProvider interface {
    // 桶操作
    ListBuckets(ctx context.Context) ([]Bucket, error)
    CreateBucket(ctx context.Context, name, region string) error
    // ... 其他方法
}
```

## 文档

### 更新文档
- API 变更必须更新文档
- 新功能需要用户文档
- 配置变更需要配置文档

### 文档位置
- 用户文档: `docs/` 目录
- API 文档: 代码注释
- 示例: `examples/` 目录

## 审查流程

### Pull Request 流程
1. 确保所有测试通过
2. 确保代码符合规范
3. 更新相关文档
4. 等待代码审查
5. 根据反馈进行修改
6. 维护者合并代码

### 审查要点
- 代码质量和可读性
- 测试覆盖率和质量
- 文档完整性
- 向后兼容性
- 性能影响

## 联系方式

- 项目维护者: 张宇峰
- GitHub Issues: [问题跟踪](https://github.com/zhangyf/cloud-storage-tool/issues)
- 讨论区: [GitHub Discussions](https://github.com/zhangyf/cloud-storage-tool/discussions)

## 许可证

通过贡献代码，您同意您的贡献将根据项目的 MIT 许可证进行许可。