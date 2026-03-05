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

---

**文档版本**: 1.0  
**最后更新**: 2026-03-05  
**状态**: 草案