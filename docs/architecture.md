# 架构设计文档

## 1. 总体架构

Cloud Storage Tool 采用**插件化架构**，核心是统一的存储接口，各个云服务提供商实现该接口。

### 1.1 架构图
```
┌─────────────────────────────────────────┐
│           命令行接口 (CLI)               │
│        cloud-storage-tool               │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│           命令分发层                     │
│        commands/*.go                    │
└─────┬───────────┬───────────┬───────────┘
      │           │           │
┌─────▼─────┐ ┌──▼───┐ ┌─────▼─────┐
│ 桶管理    │ │ 对象 │ │  同步     │
│ bucket    │ │object│ │  sync     │
└─────┬─────┘ └──┬───┘ └─────┬─────┘
      │          │           │
┌─────▼──────────▼───────────▼─────┐
│       存储提供商接口              │
│     provider.Interface           │
└─────┬──────────┬──────────┬──────┘
      │          │          │
┌─────▼────┐ ┌──▼──┐ ┌─────▼────┐
│ 腾讯云COS │ │AWS S3│ │ 阿里云OSS│
│  provider │ │provider│ │provider│
└──────────┘ └──────┘ └──────────┘
```

### 1.2 核心组件

#### 1.2.1 命令行接口 (CLI)
- 使用 **Cobra** 框架
- 支持子命令：`cos`, `s3`, `sync`, `config`
- 统一的参数解析和错误处理

#### 1.2.2 命令分发层
- 将CLI命令映射到具体的操作
- 处理命令参数验证
- 调用底层的存储提供商

#### 1.2.3 存储提供商接口
- 统一的存储操作接口
- 每个云服务商实现该接口
- 支持热插拔，易于扩展

## 2. 核心接口设计

### 2.1 存储提供商接口
```go
// StorageProvider 统一的存储提供商接口
type StorageProvider interface {
    // 桶操作
    ListBuckets(ctx context.Context) ([]Bucket, error)
    CreateBucket(ctx context.Context, name, region string) error
    DeleteBucket(ctx context.Context, name string, force bool) error
    GetBucketInfo(ctx context.Context, name string) (BucketInfo, error)
    
    // 对象操作
    ListObjects(ctx context.Context, bucket, prefix string) ([]Object, error)
    UploadFile(ctx context.Context, bucket, key, filepath string, opts UploadOptions) error
    DownloadFile(ctx context.Context, bucket, key, filepath string, opts DownloadOptions) error
    DeleteObject(ctx context.Context, bucket, key string) error
    CopyObject(ctx context.Context, src, dst ObjectInfo) error
    GetObjectInfo(ctx context.Context, bucket, key string) (ObjectInfo, error)
    
    // 高级功能
    GetPresignedURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error)
    SetObjectACL(ctx context.Context, bucket, key string, acl string) error
    
    // 提供商信息
    Name() string
    Version() string
    Capabilities() ProviderCapabilities
}
```

### 2.2 数据结构
```go
// Bucket 桶信息
type Bucket struct {
    Name      string
    Region    string
    CreatedAt time.Time
    Size      int64
    Objects   int64
}

// Object 对象信息
type Object struct {
    Key          string
    Size         int64
    LastModified time.Time
    ETag         string
    StorageClass string
}

// ObjectInfo 对象详细信息
type ObjectInfo struct {
    Bucket      string
    Key         string
    Size        int64
    ContentType string
    Metadata    map[string]string
}

// ProviderCapabilities 提供商能力
type ProviderCapabilities struct {
    SupportsMultipartUpload bool
    SupportsPresignedURL    bool
    SupportsVersioning      bool
    SupportsLifecycle       bool
    MaxPartSize             int64
    MaxObjectSize           int64
}
```

## 3. 配置管理

### 3.1 配置文件结构
采用 **YAML** 格式，支持多级配置和继承。

### 3.2 配置加载顺序
1. 命令行参数（最高优先级）
2. 环境变量
3. 用户配置文件 (`~/.cloud-storage/config.yaml`)
4. 系统配置文件 (`/etc/cloud-storage/config.yaml`)
5. 默认值（最低优先级）

### 3.3 配置加密
- 敏感信息（密钥）支持加密存储
- 支持密钥管理服务集成
- 配置文件中可以使用环境变量引用

## 4. 并发与性能

### 4.1 并发模型
- 使用 **工作池模式** 处理并发上传/下载
- 支持可配置的并发数
- 连接池管理HTTP连接

### 4.2 大文件支持
- **分块上传**：大文件自动分块
- **断点续传**：支持上传中断后继续
- **进度显示**：实时显示上传/下载进度

### 4.3 内存优化
- **流式处理**：避免大文件完全加载到内存
- **缓冲区管理**：可配置的缓冲区大小
- **内存池**：重用内存减少分配

## 5. 错误处理与恢复

### 5.1 错误分类
```go
type ErrorCategory int

const (
    CategoryNetwork     ErrorCategory = iota // 网络错误
    CategoryAuth                             // 认证错误
    CategoryPermission                       // 权限错误
    CategoryResource                         // 资源错误
    CategoryValidation                       // 验证错误
    CategoryInternal                         // 内部错误
)
```

### 5.2 重试机制
- 指数退避重试策略
- 可配置的最大重试次数
- 可重试错误自动重试

### 5.3 事务性操作
- 重要操作支持回滚
- 批量操作的原子性保证
- 操作日志记录便于恢复

## 6. 扩展性设计

### 6.1 插件系统
- 提供商接口易于实现
- 动态加载提供商插件
- 配置文件自动发现插件

### 6.2 中间件支持
```go
// StorageMiddleware 存储中间件
type StorageMiddleware interface {
    BeforeOperation(ctx context.Context, op Operation) (context.Context, error)
    AfterOperation(ctx context.Context, op Operation, result interface{}, err error) error
}

// 支持中间件：日志、监控、缓存、限流等
```

## 7. 安全设计

### 7.1 认证与授权
- 支持多种认证方式（密钥对、IAM角色等）
- 最小权限原则
- 临时凭证自动刷新

### 7.2 数据安全
- 支持客户端加密
- 传输层加密（TLS 1.3）
- 数据完整性校验（MD5/SHA256）

### 7.3 审计日志
- 所有操作记录审计日志
- 支持结构化日志输出
- 日志可配置输出到文件/系统

## 8. 监控与可观测性

### 8.1 指标收集
- 操作成功率、延迟
- 带宽使用情况
- 并发连接数

### 8.2 日志分级
- DEBUG：详细调试信息
- INFO：常规操作信息
- WARN：警告信息
- ERROR：错误信息

### 8.3 跟踪支持
- 支持分布式跟踪
- 请求ID贯穿所有操作
- 便于问题排查

## 9. 部署与维护

### 9.1 构建系统
- 支持交叉编译
- 版本信息嵌入二进制
- 自动化测试和构建

### 9.2 依赖管理
- Go Modules 管理依赖
- 版本锁定确保一致性
- 依赖安全检查

### 9.3 升级策略
- 向后兼容的API设计
- 配置迁移工具
- 版本发布说明

## 10. 代码规范约定

### 10.1 代码组织规范

#### 10.1.1 项目目录结构
```
cloud-storage-tool/
├── cmd/cloud-storage-tool/     # 命令行入口
│   └── main.go
├── internal/                   # 内部包（不对外暴露）
│   ├── commands/              # 命令实现
│   ├── providers/             # 存储提供商实现
│   ├── config/                # 配置管理
│   ├── utils/                 # 工具函数
│   └── types/                 # 类型定义
├── pkg/                       # 可对外暴露的包
│   ├── storage/               # 存储接口定义
│   └── errors/                # 错误类型定义
├── scripts/                   # 构建和部署脚本
├── tests/                     # 测试文件
├── docs/                      # 文档
└── .github/workflows/         # CI/CD配置
```

#### 10.1.2 包设计原则
- **单一职责**：每个包只关注一个功能领域
- **依赖倒置**：高层模块不依赖低层模块，都依赖抽象
- **接口隔离**：定义小而专注的接口
- **包可见性**：internal包内的代码不对外暴露

### 10.2 命名约定

#### 10.2.1 文件命名
- **Go源文件**：使用小写蛇形命名，如 `bucket_manager.go`
- **测试文件**：源文件名加 `_test` 后缀，如 `bucket_manager_test.go`
- **配置文件**：使用 `.yaml` 或 `.yml` 扩展名

#### 10.2.2 标识符命名
- **包名**：小写单数名词，简短明了
- **接口名**：`er` 结尾，如 `Provider`, `Uploader`
- **变量名**：驼峰式，见名知意
- **常量名**：全大写蛇形，如 `MAX_RETRY_COUNT`

#### 10.2.3 方法命名
- **Getter方法**：不需要 `Get` 前缀，如 `Name()` 而不是 `GetName()`
- **布尔方法**：使用 `Is`, `Has`, `Can` 前缀，如 `IsValid()`
- **操作方法**：动词开头，如 `UploadFile()`, `DeleteObject()`

### 10.3 错误处理规范

#### 10.3.1 错误定义
```go
// 自定义错误类型
type StorageError struct {
    Code    ErrorCode
    Message string
    Cause   error
    Context map[string]interface{}
}

func (e *StorageError) Error() string {
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}
```

#### 10.3.2 错误处理原则
- **错误传播**：低层错误应该包装上下文信息后向上传递
- **错误日志**：错误发生时要记录足够的上下文信息
- **用户友好**：给用户的错误信息要友好，技术细节记录在日志中
- **错误恢复**：可恢复错误应该提供重试机制

#### 10.3.3 错误码定义
```go
const (
    ErrBucketNotFound    ErrorCode = "BUCKET_NOT_FOUND"
    ErrAccessDenied      ErrorCode = "ACCESS_DENIED"
    ErrNetworkTimeout    ErrorCode = "NETWORK_TIMEOUT"
    ErrInvalidParameter  ErrorCode = "INVALID_PARAMETER"
    ErrInternalError     ErrorCode = "INTERNAL_ERROR"
)
```

### 10.4 测试规范

#### 10.4.1 测试覆盖率要求
- **单元测试**：核心功能覆盖率 ≥ 80%
- **集成测试**：关键路径覆盖率 ≥ 90%
- **测试文件**：每个源文件都有对应的测试文件

#### 10.4.2 测试组织
```go
// 表驱动测试
func TestUploadFile(t *testing.T) {
    testCases := []struct {
        name        string
        input       UploadInput
        expectError bool
        errorCode   ErrorCode
    }{
        {
            name: "正常上传",
            input: UploadInput{...},
            expectError: false,
        },
        {
            name: "文件不存在",
            input: UploadInput{...},
            expectError: true,
            errorCode: ErrFileNotFound,
        },
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // 测试逻辑
        })
    }
}
```

#### 10.4.3 Mock和Stub
- **接口隔离**：通过接口实现测试替身
- **gomock**：使用gomock生成mock代码
- **测试数据**：使用testdata目录存放测试数据

### 10.5 文档规范

#### 10.5.1 代码注释
- **包注释**：每个包都要有包注释，说明包的功能
- **导出注释**：所有导出的函数、类型、变量都要有注释
- **示例代码**：复杂功能要提供使用示例

#### 10.5.2 GoDoc要求
```go
// BucketManager 桶管理器，负责桶的创建、删除、查询等操作
//
// 示例：
//   manager := NewBucketManager(provider)
//   err := manager.CreateBucket("my-bucket", "ap-singapore")
//   if err != nil {
//       log.Fatal(err)
//   }
type BucketManager struct {
    // 字段说明
    provider StorageProvider
}

// CreateBucket 创建存储桶
//
// 参数：
//   name: 桶名称，必须全局唯一
//   region: 区域代码，如 "ap-singapore"
//
// 返回：
//   error: 创建失败时返回错误信息
func (m *BucketManager) CreateBucket(name, region string) error {
    // 实现逻辑
}
```

### 10.6 并发规范

#### 10.6.1 Goroutine管理
- **生命周期**：明确Goroutine的创建和终止
- **错误传播**：Goroutine中的错误要能传播到主流程
- **资源清理**：Goroutine退出时要清理占用的资源

#### 10.6.2 并发模式
```go
// 使用工作池模式处理并发任务
func processConcurrently(tasks []Task, concurrency int) error {
    var wg sync.WaitGroup
    taskChan := make(chan Task)
    errChan := make(chan error, concurrency)
    
    // 创建工作池
    for i := 0; i < concurrency; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for task := range taskChan {
                if err := processTask(task); err != nil {
                    errChan <- err
                }
            }
        }()
    }
    
    // 分发任务
    for _, task := range tasks {
        taskChan <- task
    }
    close(taskChan)
    
    // 等待完成
    wg.Wait()
    close(errChan)
    
    // 收集错误
    var errors []error
    for err := range errChan {
        errors = append(errors, err)
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("处理失败: %v", errors)
    }
    return nil
}
```

#### 10.6.3 同步原语
- **互斥锁**：保护共享资源的访问
- **读写锁**：读多写少的场景
- **条件变量**：复杂的同步需求
- **原子操作**：简单的计数器等

### 10.7 安全规范

#### 10.7.1 输入验证
- **边界检查**：所有输入都要检查边界
- **类型验证**：确保输入符合预期的类型
- **格式验证**：验证URL、路径等格式

#### 10.7.2 密钥管理
- **环境变量**：AK/SK等敏感信息必须通过环境变量配置
- **加密存储**：配置文件中的敏感信息要加密
- **访问控制**：最小权限原则，按需分配权限

#### 10.7.3 日志安全
- **脱敏处理**：日志中不能包含敏感信息
- **访问控制**：日志文件要有适当的权限控制
- **日志轮转**：定期清理旧日志，避免存储过多敏感信息

### 10.8 性能规范

#### 10.8.1 内存管理
- **对象复用**：避免频繁创建和销毁对象
- **缓冲区池**：使用sync.Pool复用缓冲区
- **内存监控**：监控内存使用，避免内存泄漏

#### 10.8.2 网络优化
- **连接复用**：复用HTTP连接，减少连接建立开销
- **压缩传输**：大文件传输时使用压缩
- **分块处理**：大文件分块传输，支持断点续传

#### 10.8.3 算法复杂度
- **时间复杂度**：核心操作要控制在O(n)或更好
- **空间复杂度**：避免大内存占用，使用流式处理
- **并发优化**：充分利用多核CPU

### 10.9 持续集成规范

#### 10.9.1 代码检查
- **golangci-lint**：配置统一的代码检查规则
- **静态分析**：使用go vet进行静态分析
- **安全扫描**：集成安全扫描工具

#### 10.9.2 构建流程
```yaml
# .github/workflows/ci.yml 示例
name: CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Setup Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    - name: Run tests
      run: go test ./... -v -coverprofile=coverage.out
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        files: coverage.out
```

#### 10.9.3 发布流程
- **版本号**：遵循语义化版本控制（SemVer）
- **变更日志**：每次发布都要更新CHANGELOG.md
- **二进制发布**：提供各平台的二进制文件下载

---

**文档版本**: 1.1  
**最后更新**: 2026-03-05  
**状态**: 草案