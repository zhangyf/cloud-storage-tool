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

### 2.1 接口设计原则

遵循**接口隔离原则**，将大的存储提供商接口拆分为多个专注的小接口，便于实现和测试。

#### 2.1.0 接口并发安全性说明

所有存储提供商接口的方法都**必须满足线程安全要求**，除非在方法注释中明确说明。具体规则如下：

1. **读操作**（GetXxx、ListXxx、ExistXxx）：允许多个goroutine并发调用
2. **写操作**（CreateXxx、UpdateXxx、DeleteXxx）：需要实现适当的并发控制，避免资源冲突
3. **连接状态**：共享的连接池必须是线程安全的
4. **错误处理**：并发错误必须正确传播，不会导致panic或数据损坏

**实现要求**：
- 使用`sync.RWMutex`保护共享状态
- 使用`atomic`操作处理计数器等简单状态
- 避免在方法内部创建全局锁
- 遵循Go的"share memory by communicating"原则

**接口方法线程安全级别**：
- ✅ **完全线程安全**：所有实现必须保证方法调用的线程安全
- ⚠️ **有条件线程安全**：某些方法可能需要在实现时添加额外同步
- 🔴 **非线程安全**：必须在文档中明确标注，调用方负责同步

#### 2.1.1 基础接口拆分

```go
// BucketManager 桶管理接口
type BucketManager interface {
    ListBuckets(ctx context.Context) ([]Bucket, error)
    CreateBucket(ctx context.Context, name, region string) error
    DeleteBucket(ctx context.Context, name string, force bool) error
    GetBucketInfo(ctx context.Context, name string) (BucketInfo, error)
    
    // 桶存在性检查
    ExistBucket(ctx context.Context, name string) (bool, error)
    
    // 桶策略管理
    SetBucketPolicy(ctx context.Context, bucket, policy string) error
    GetBucketPolicy(ctx context.Context, bucket) (string, error)
    
    // 桶标签管理
    SetBucketTagging(ctx context.Context, bucket string, tags map[string]string) error
    GetBucketTagging(ctx context.Context, bucket string) (map[string]string, error)
    DeleteBucketTagging(ctx context.Context, bucket string) error
    
    // 桶加密配置
    SetBucketEncryption(ctx context.Context, bucket string, config *SSEConfig) error
    GetBucketEncryption(ctx context.Context, bucket string) (*SSEConfig, error)
    DeleteBucketEncryption(ctx context.Context, bucket string) error
    
    // 桶版本控制
    GetBucketVersioning(ctx context.Context, bucket string) (VersioningStatus, error)
    SetBucketVersioning(ctx context.Context, bucket string, status VersioningStatus) error
}

### 2.1.2 分页支持机制

Cloud Storage Tool 提供完整的分页支持，处理大规模对象列表查询。

#### 分页设计原理
1. **游标分页**：使用`ContinuationToken`代替偏移量，避免数据变化导致的重复或遗漏
2. **可配置页面大小**：通过`MaxKeys`参数控制每页返回的对象数量
3. **递归查询**：支持目录式查询和扁平化查询
4. **高效内存使用**：流式处理分页结果，避免一次性加载所有数据

#### ListObjects 分页示例
```go
// 分页查询对象列表
func listAllObjects(ctx context.Context, provider ObjectManager, bucket, prefix string) ([]Object, error) {
    var allObjects []Object
    var continuationToken string
    
    for {
        opts := ListOptions{
            Prefix:             prefix,
            MaxKeys:            1000, // 每页1000个对象
            ContinuationToken:  continuationToken,
            Recursive:          true,
        }
        
        result, err := provider.ListObjects(ctx, bucket, prefix, opts)
        if err != nil {
            return nil, err
        }
        
        allObjects = append(allObjects, result.Objects...)
        
        if !result.IsTruncated {
            break
        }
        
        continuationToken = result.NextContinuationToken
    }
    
    return allObjects, nil
}
```

#### 分页性能优化
- **并发预取**：后台预取下一页数据
- **缓存策略**：缓存常用前缀的列表结果
- **增量更新**：支持增量查询，只获取新增对象
- **超时控制**：长时间查询自动超时，避免资源占用

#### 分页状态管理
```go
// ListOptions 分页选项
type ListOptions struct {
    Prefix     string          // 对象前缀过滤
    Delimiter  string          // 目录分隔符（如"/"）
    MaxKeys    int             // 每页最大对象数（默认1000）
    ContinuationToken string   // 继续令牌（用于分页）
    Recursive  bool            // 是否递归查询
}

// ListObjectsResult 分页结果
type ListObjectsResult struct {
    Objects           []Object // 当前页对象列表
    IsTruncated       bool     // 是否还有更多数据
    ContinuationToken string   // 当前页令牌（用于调试）
    NextContinuationToken string // 下一页令牌
    CommonPrefixes    []string // 公共前缀（使用Delimiter时）
}
```

```go
// ObjectManager 对象管理接口
type ObjectManager interface {
    // 对象列表（支持分页）
    ListObjects(ctx context.Context, bucket, prefix string, opts ListOptions) (*ListObjectsResult, error)
    
    // 对象上传/下载（支持流式操作）
    UploadFile(ctx context.Context, bucket, key string, data io.Reader, opts UploadOptions) error
    DownloadFile(ctx context.Context, bucket, key string, writer io.Writer, opts DownloadOptions) error
    
    // 对象删除
    DeleteObject(ctx context.Context, bucket, key string) error
    DeleteObjectsBatch(ctx context.Context, bucket string, keys []string) ([]DeleteResult, error)
    
    // 对象复制和移动
    CopyObject(ctx context.Context, srcBucket, srcKey, dstBucket, dstKey string) error
    MoveObject(ctx context.Context, srcBucket, srcKey, dstBucket, dstKey string) error
    
    // 对象元数据操作
    GetObjectInfo(ctx context.Context, bucket, key string) (ObjectInfo, error)
    StatObject(ctx context.Context, bucket, key string) (ObjectStat, error)
    ObjectExists(ctx context.Context, bucket, key string) (bool, error)
    
    // 高级对象操作
    SelectObjectContent(ctx context.Context, bucket, key string, query string, opts SelectOptions) (io.ReadCloser, error)
    
    // 对象标签管理
    SetObjectTagging(ctx context.Context, bucket, key string, tags map[string]string) error
    GetObjectTagging(ctx context.Context, bucket, key string) (map[string]string, error)
    DeleteObjectTagging(ctx context.Context, bucket, key string) error
}

// MultipartUploadManager 分块上传管理接口
type MultipartUploadManager interface {
    // 分块上传生命周期管理
    InitiateMultipartUpload(ctx context.Context, bucket, key string) (string, error)
    UploadPart(ctx context.Context, bucket, key, uploadID string, partNumber int, data io.Reader) (string, error)
    UploadPartFromURL(ctx context.Context, bucket, key, uploadID string, partNumber int, sourceURL string) (string, error)
    CompleteMultipartUpload(ctx context.Context, bucket, key, uploadID string, parts []CompletedPart) error
    AbortMultipartUpload(ctx context.Context, bucket, key, uploadID string) error
    ListMultipartUploads(ctx context.Context, bucket, prefix string) ([]MultipartUpload, error)
    
    // 分块上传工具方法
    PresignUploadPartURL(ctx context.Context, bucket, key, uploadID string, partNumber int, expires time.Duration) (string, error)
    TerminateAllMultipartUploads(ctx context.Context, bucket string, gracePeriod time.Duration) error
    
    // 分块上传状态查询
    GetMultipartUploadStatus(ctx context.Context, bucket, key, uploadID string) (*MultipartUploadStatus, error)
}

// ACLManager 访问控制管理接口
type ACLManager interface {
    GetObjectACL(ctx context.Context, bucket, key string) (string, error)
    SetObjectACL(ctx context.Context, bucket, key string, acl string) error
    GetBucketACL(ctx context.Context, bucket string) (string, error)
    SetBucketACL(ctx context.Context, bucket string, acl string) error
}

// LifecycleManager 生命周期管理接口
type LifecycleManager interface {
    GetBucketLifecycle(ctx context.Context, bucket string) ([]LifecycleRule, error)
    SetBucketLifecycle(ctx context.Context, bucket string, rules []LifecycleRule) error
    DeleteBucketLifecycle(ctx context.Context, bucket string) error
}

// PresignedURLManager 预签名URL管理接口
type PresignedURLManager interface {
    GetPresignedURL(ctx context.Context, bucket, key string, expires time.Duration, method string) (string, error)
    GeneratePresignedPutURL(ctx context.Context, bucket, key string, expires time.Duration, headers map[string]string) (string, error)
    GeneratePresignedGetURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error)
}

// ProviderInfo 提供商信息接口
type ProviderInfo interface {
    // 基础信息
    Name() string
    Version() string
    Capabilities() ProviderCapabilities
    
    // 区域信息
    Region() string
    SupportedRegions() []string
    IsRegionSupported(region string) bool
    
    // 功能检查
    Supports(feature string) bool
    
    // 提供商元数据
    GetProviderMetadata() map[string]string
}

// StorageProvider 统一存储提供商接口（组合接口）
type StorageProvider interface {
    BucketManager
    ObjectManager
    MultipartUploadManager
    ACLManager
    LifecycleManager
    PresignedURLManager
    ProviderInfo
}
```

### 2.2 新增数据结构
```go
// CompletedPart 分块上传完成的部分
type CompletedPart struct {
    ETag       string
    PartNumber int
}

// MultipartUpload 分块上传信息
type MultipartUpload struct {
    UploadID    string
    Key         string
    Initiated   time.Time
    Size        int64
}

// DeleteResult 批量删除结果
type DeleteResult struct {
    Key     string
    Deleted bool
    Error   error
}

// LifecycleRule 生命周期规则
type LifecycleRule struct {
    ID                     string
    Prefix                 string
    Status                 string // Enabled/Disabled
    ExpirationDays         int
    TransitionDays         int
    TransitionStorageClass string
    NoncurrentVersionExpirationDays int
}

// UploadOptions 上传选项
type UploadOptions struct {
    ContentType        string
    ContentEncoding    string
    ContentLanguage    string
    CacheControl       string
    ContentDisposition string
    Metadata           map[string]string
    StorageClass       string
    ACL                string
    SSEConfig          *SSEConfig
    ChecksumAlgorithm  string // MD5, SHA256, CRC32C等
    PartSize           int64  // 分块大小
    Concurrency        int    // 并发数
    WebsiteRedirectLocation string
    Expires            time.Time
    LegalHold          bool
    RetentionMode      string
    RetentionUntilDate time.Time
}

// DownloadOptions 下载选项
type DownloadOptions struct {
    Range              string // HTTP Range头，如"bytes=0-100"
    IfMatch            string // ETag条件
    IfNoneMatch        string // ETag条件（不匹配时下载）
    IfModifiedSince    time.Time
    IfUnmodifiedSince  time.Time
    ChecksumValidation bool   // 是否验证校验和
    Concurrency        int    // 并发下载分块数
    VersionId          string // 对象版本ID
    ResponseHeaders    map[string]string // 响应头覆盖
    ProgressFunc       func(bytesTransferred, totalBytes int64) // 进度回调
}


```

### 2.2 数据结构
```go
// Bucket 桶信息
type Bucket struct {
    Name           string
    Region         string
    CreatedAt      time.Time
    Size           int64
    Objects        int64
    Owner          string
    Tags           map[string]string
    SSEConfig      *SSEConfig
    Versioning     VersioningStatus
    ACL            string
    Policy         string
    LifecycleRules []LifecycleRule
}

// Object 对象信息
type Object struct {
    Key          string
    Size         int64
    LastModified time.Time
    ETag         string
    StorageClass string
    ContentType  string
    Metadata     map[string]string
    ACL          string
    Checksum     string
    IsTruncated  bool
    Owner        string
    Tags         map[string]string
}

// ObjectInfo 对象详细信息
type ObjectInfo struct {
    Bucket                   string
    Key                      string
    Size                     int64
    ContentType              string
    Metadata                 map[string]string
    CacheControl             string
    ContentDisposition       string
    ContentEncoding          string
    ContentLanguage          string
    StorageClass             string
    WebsiteRedirectLocation  string
    ServerSideEncryption     string
    SSEKMSKeyId              string
    VersionId                string
    Expiration               time.Time
    LastModified             time.Time
    ETag                     string
    Checksum                 string
    ACL                      string
    Tags                     map[string]string
}

// 新增：对象统计信息（轻量级元数据）
type ObjectStat struct {
    Key          string
    Size         int64
    LastModified time.Time
    ETag         string
    StorageClass string
    ContentType  string
    Metadata     map[string]string
    ACL          string
    Checksum     string
}

// 新增：列表选项（支持分页）
type ListOptions struct {
    Prefix     string
    Delimiter  string
    MaxKeys    int
    ContinuationToken string
    Recursive  bool
}

// 新增：列表结果（支持分页）
type ListObjectsResult struct {
    Objects           []Object
    IsTruncated       bool
    ContinuationToken string
    NextContinuationToken string
    CommonPrefixes    []string
}

// 新增：版本控制状态
type VersioningStatus string

const (
    VersioningEnabled    VersioningStatus = "Enabled"
    VersioningSuspended  VersioningStatus = "Suspended"
    VersioningDisabled   VersioningStatus = "Disabled"
)

// 新增：服务端加密配置
type SSEConfig struct {
    Algorithm string // "AES256", "aws:kms", "cos:kms"等
    KMSKeyID  string // KMS密钥ID（如使用KMS）
    Context   map[string]string // 加密上下文
}

// 新增：Select查询选项
type SelectOptions struct {
    Expression     string
    ExpressionType string // "SQL"
    InputSerialization  *InputSerialization
    OutputSerialization *OutputSerialization
    ScanRange      *ScanRange
}

// 新增：输入序列化格式
type InputSerialization struct {
    CSV  *CSVInput
    JSON *JSONInput
    Parquet *ParquetInput
}

// 新增：输出序列化格式
type OutputSerialization struct {
    CSV  *CSVOutput
    JSON *JSONOutput
}

// 新增：CSV输入格式
type CSVInput struct {
    FileHeaderInfo       string // "USE", "IGNORE", "NONE"
    Comments             string
    QuoteEscapeCharacter string
    RecordDelimiter      string
    FieldDelimiter       string
    QuoteCharacter       string
    AllowQuotedRecordDelimiter bool
}

// 新增：JSON输入格式
type JSONInput struct {
    Type string // "DOCUMENT", "LINES"
}

// 新增：Parquet输入格式
type ParquetInput struct {
    // Parquet格式不需要额外配置
}

// 新增：CSV输出格式
type CSVOutput struct {
    QuoteFields          string // "ASNEEDED", "ALWAYS"
    QuoteEscapeCharacter string
    RecordDelimiter      string
    FieldDelimiter       string
    QuoteCharacter       string
}

// 新增：JSON输出格式
type JSONOutput struct {
    RecordDelimiter string
}

// 新增：扫描范围
type ScanRange struct {
    Start  int64
    End    int64
}

// 新增：分块上传状态
type MultipartUploadStatus struct {
    UploadID    string
    Key         string
    Initiated   time.Time
    PartsUploaded int
    TotalSize   int64
    CompletedSize int64
    IsCompleted bool
}

// ProviderCapabilities 提供商能力
type ProviderCapabilities struct {
    SupportsMultipartUpload        bool
    SupportsPresignedURL           bool
    SupportsVersioning             bool
    SupportsLifecycle              bool
    SupportsServerSideEncryption   bool
    SupportsBucketPolicy           bool
    SupportsACL                    bool
    SupportsObjectLock             bool
    SupportsIntelligentTiering     bool
    SupportsBatchOperations        bool
    SupportsSelectContent          bool
    MaxPartSize                    int64
    MaxObjectSize                  int64
    MaxMultipartParts              int
    MaxBatchDeleteSize             int
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

### 4.4 资源清理机制

Cloud Storage Tool 提供完善的资源清理机制，防止资源泄漏和悬空状态。

#### 4.4.1 分块上传资源清理
分块上传失败或中断时，系统自动清理悬空资源：

```go
// 强制终止所有分块上传（生产环境重要操作）
TerminateAllMultipartUploads(ctx context.Context, bucket string, gracePeriod time.Duration) error

// 自动清理机制
- 分块上传创建时记录到持久化存储
- 定期扫描超时的分块上传（默认24小时）
- 提供API手动清理特定上传
- 支持grace period避免误删进行中的上传
```

#### 4.4.2 连接和缓冲区清理
- **连接池自动回收**：闲置连接超时自动关闭
- **缓冲区重用池**：避免重复分配内存
- **文件句柄管理**：确保上传/下载后文件正确关闭
- **临时文件清理**：操作失败时自动清理临时文件

#### 4.4.3 监控和告警
- **资源使用监控**：实时监控连接数、内存使用、文件句柄
- **自动告警**：资源泄漏时自动触发告警
- **健康检查**：定期检查资源状态并自动恢复

#### 4.4.4 优雅关闭
```go
// StorageProvider 关闭接口
type CleanupProvider interface {
    // 优雅关闭，清理所有资源
    Close() error
    
    // 强制关闭，立即释放资源
    ForceClose() error
    
    // 获取资源状态
    ResourceStats() ResourceStats
}

// 资源状态统计
type ResourceStats struct {
    OpenConnections    int
    ActiveUploads      int
    ActiveDownloads    int
    MultipartUploads   int
    MemoryInUse        int64
    BufferPoolSize     int
}
```

## 5. 错误处理与恢复

### 5.1 增强的错误分类与标准化

**错误分类机制的重要性**：
准确的错误分类是系统可靠性的基础，它允许：
1. **智能重试决策**：区分可重试错误和不可重试错误
2. **用户体验优化**：提供有针对性的错误提示和建议
3. **运维监控**：基于错误类型进行监控和告警
4. **故障诊断**：快速定位问题根源

**错误分类原则**：
1. **完整性**：覆盖所有可能的错误场景
2. **一致性**：相同类型的错误使用相同的错误码
3. **可操作性**：每个错误都提供明确的处理建议
4. **可扩展性**：支持未来新增错误类型

#### 5.1.1 详细错误分类
```go
type ErrorCategory int

const (
    CategoryNetwork       ErrorCategory = iota // 网络错误（可重试）
    CategoryAuth                               // 认证错误（不可重试）
    CategoryPermission                         // 权限错误（不可重试）
    CategoryResource                           // 资源错误（通常不可重试）
    CategoryValidation                         // 验证错误（不可重试）
    CategoryInternal                           // 内部错误（视情况）
    CategoryRateLimited                        // 限流错误（可重试）
    CategoryConflict                           // 冲突错误（如并发修改）
    CategoryTimeout                            // 超时错误（可重试）
    CategoryQuotaExceeded                      // 配额超出错误（不可重试）
    CategoryMaintenance                        // 服务维护错误（可重试）
    CategoryUnavailable                        // 服务不可用错误（可重试）
)

// 错误码标准化
const (
    // 网络相关错误
    ErrNetworkTimeout      ErrorCode = "NETWORK_TIMEOUT"
    ErrConnectionFailed    ErrorCode = "CONNECTION_FAILED"
    ErrSSLHandshakeFailed  ErrorCode = "SSL_HANDSHAKE_FAILED"
    
    // 认证授权错误
    ErrAccessDenied        ErrorCode = "ACCESS_DENIED"
    ErrInvalidCredentials  ErrorCode = "INVALID_CREDENTIALS"
    ErrTokenExpired        ErrorCode = "TOKEN_EXPIRED"
    ErrSignatureInvalid    ErrorCode = "SIGNATURE_INVALID"
    
    // 资源错误
    ErrBucketNotFound      ErrorCode = "BUCKET_NOT_FOUND"
    ErrObjectNotFound      ErrorCode = "OBJECT_NOT_FOUND"
    ErrBucketAlreadyExists ErrorCode = "BUCKET_ALREADY_EXISTS"
    ErrObjectAlreadyExists ErrorCode = "OBJECT_ALREADY_EXISTS"
    
    // 验证错误
    ErrInvalidParameter    ErrorCode = "INVALID_PARAMETER"
    ErrMissingParameter    ErrorCode = "MISSING_PARAMETER"
    ErrInvalidRange        ErrorCode = "INVALID_RANGE"
    ErrChecksumMismatch    ErrorCode = "CHECKSUM_MISMATCH"
    
    // 系统错误
    ErrInternalError       ErrorCode = "INTERNAL_ERROR"
    ErrOutOfMemory         ErrorCode = "OUT_OF_MEMORY"
    ErrDiskFull            ErrorCode = "DISK_FULL"
    
    // 业务错误
    ErrRateLimited         ErrorCode = "RATE_LIMITED"
    ErrQuotaExceeded       ErrorCode = "QUOTA_EXCEEDED"
    ErrServiceUnavailable  ErrorCode = "SERVICE_UNAVAILABLE"
    ErrOperationTimeout    ErrorCode = "OPERATION_TIMEOUT"
)

// StorageError 增强的错误类型
type StorageError struct {
    Code         ErrorCode               // 错误码
    Category     ErrorCategory           // 错误分类
    HTTPStatus   int                     // HTTP状态码映射
    Message      string                  // 用户友好错误信息
    Cause        error                   // 底层错误
    Context      map[string]interface{}  // 上下文信息
    Timestamp    time.Time               // 错误发生时间
    Retryable    bool                    // 是否可重试
    SuggestedAction string               // 建议操作
    RequestID    string                  // 请求ID（用于追踪）
    
    // 可重试相关字段
    RetryAfter   time.Duration           // 建议重试等待时间
    MaxRetries   int                     // 最大重试次数
}

func (e *StorageError) Error() string {
    if e.Cause != nil {
        return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
    }
    return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// HTTP状态码映射
func (e *StorageError) MapToHTTPStatus() int {
    switch e.Category {
    case CategoryNetwork, CategoryTimeout, CategoryUnavailable:
        return http.StatusServiceUnavailable
    case CategoryAuth, CategoryPermission:
        return http.StatusForbidden
    case CategoryResource:
        return http.StatusNotFound
    case CategoryValidation:
        return http.StatusBadRequest
    case CategoryRateLimited:
        return http.StatusTooManyRequests
    case CategoryConflict:
        return http.StatusConflict
    case CategoryQuotaExceeded:
        return http.StatusInsufficientStorage
    default:
        return http.StatusInternalServerError
    }
}
```

### 5.2 智能重试机制

#### 5.2.1 重试配置
```go
// RetryConfig 重试配置
type RetryConfig struct {
    MaxRetries          int           // 最大重试次数
    InitialDelay        time.Duration // 初始延迟
    MaxDelay            time.Duration // 最大延迟
    BackoffFactor       float64       // 退避因子（指数退避）
    Jitter              float64       // 抖动因子（随机化延迟）
    RetryableCodes      []ErrorCode   // 可重试错误码白名单
    NonRetryableCodes   []ErrorCode   // 不可重试错误码黑名单
    TimeoutMultiplier   float64       // 超时乘数（每次重试增加超时）
}

// 默认重试配置
var DefaultRetryConfig = RetryConfig{
    MaxRetries:          3,
    InitialDelay:        100 * time.Millisecond,
    MaxDelay:            30 * time.Second,
    BackoffFactor:       2.0,
    Jitter:              0.1, // 10%抖动
    RetryableCodes: []ErrorCode{
        ErrNetworkTimeout,
        ErrConnectionFailed,
        ErrRateLimited,
        ErrServiceUnavailable,
        ErrOperationTimeout,
    },
    NonRetryableCodes: []ErrorCode{
        ErrAccessDenied,
        ErrInvalidCredentials,
        ErrBucketNotFound,
        ErrObjectNotFound,
        ErrInvalidParameter,
        ErrQuotaExceeded,
    },
    TimeoutMultiplier: 1.5, // 每次重试增加50%超时
}

// 操作特定的重试配置
var OperationRetryConfigs = map[string]RetryConfig{
    "upload": {
        MaxRetries:    5,
        InitialDelay:  1 * time.Second,
        MaxDelay:      60 * time.Second,
        BackoffFactor: 2.0,
    },
    "download": {
        MaxRetries:    3,
        InitialDelay:  500 * time.Millisecond,
        MaxDelay:      10 * time.Second,
        BackoffFactor: 1.5,
    },
    "delete": {
        MaxRetries:    2, // 删除操作谨慎重试
        InitialDelay:  2 * time.Second,
        MaxDelay:      5 * time.Second,
        BackoffFactor: 1.2,
    },
}
```

#### 5.2.2 智能重试决策
```go
// RetryDecider 重试决策器
type RetryDecider interface {
    ShouldRetry(err error, attempt int, operation string) (bool, time.Duration)
    RecordResult(err error, attempt int, duration time.Duration)
}

// AdaptiveRetryDecider 自适应重试决策器
type AdaptiveRetryDecider struct {
    config        RetryConfig
    stats         *RetryStats
    successRate   float64
    mu            sync.RWMutex
}

// ShouldRetry 判断是否应该重试
func (d *AdaptiveRetryDecider) ShouldRetry(err error, attempt int, operation string) (bool, time.Duration) {
    d.mu.RLock()
    defer d.mu.RUnlock()
    
    // 检查是否超过最大重试次数
    if attempt >= d.config.MaxRetries {
        return false, 0
    }
    
    // 转换为StorageError
    var storageErr *StorageError
    if errors.As(err, &storageErr) {
        // 检查错误码黑名单
        for _, code := range d.config.NonRetryableCodes {
            if storageErr.Code == code {
                return false, 0
            }
        }
        
        // 检查错误码白名单
        for _, code := range d.config.RetryableCodes {
            if storageErr.Code == code {
                delay := d.calculateDelay(attempt, storageErr.RetryAfter)
                return true, delay
            }
        }
        
        // 根据错误分类决定
        if storageErr.Retryable {
            delay := d.calculateDelay(attempt, storageErr.RetryAfter)
            return true, delay
        }
    }
    
    // 默认不重试
    return false, 0
}

// calculateDelay 计算重试延迟（指数退避+抖动）
func (d *AdaptiveRetryDecider) calculateDelay(attempt int, suggested time.Duration) time.Duration {
    if suggested > 0 {
        return suggested
    }
    
    // 指数退避：delay = initial * backoff^attempt
    delay := float64(d.config.InitialDelay) * math.Pow(d.config.BackoffFactor, float64(attempt))
    
    // 添加随机抖动：delay = delay * (1 ± jitter)
    jitter := 1.0 + (rand.Float64()*2-1)*d.config.Jitter
    delay = delay * jitter
    
    // 限制最大延迟
    if time.Duration(delay) > d.config.MaxDelay {
        return d.config.MaxDelay
    }
    
    return time.Duration(delay)
}
```

### 5.3 增强的事务性操作

#### 5.3.1 操作回滚支持
```go
// TransactionalOperation 事务性操作
type TransactionalOperation struct {
    ID          string
    Operation   string
    State       OperationState
    Steps       []OperationStep
    CreatedAt   time.Time
    UpdatedAt   time.Time
    RollbackFn  func() error
    CleanupFn   func() error
}

// OperationStep 操作步骤
type OperationStep struct {
    ID        string
    Action    string
    Params    map[string]interface{}
    Result    interface{}
    Error     error
    Committed bool
    Rollback  func() error
}

// ExecuteWithRollback 执行带回滚的操作
func ExecuteWithRollback(ctx context.Context, steps []OperationStep) error {
    var executedSteps []OperationStep
    var lastError error
    
    for i, step := range steps {
        // 执行步骤
        if err := executeStep(ctx, step); err != nil {
            lastError = err
            // 回滚已执行的步骤
            rollbackErr := rollbackSteps(executedSteps)
            if rollbackErr != nil {
                return fmt.Errorf("operation failed: %v, rollback also failed: %v", err, rollbackErr)
            }
            return fmt.Errorf("operation failed at step %d: %v", i, err)
        }
        executedSteps = append(executedSteps, step)
    }
    
    return nil
}

// rollbackSteps 回滚步骤（逆序执行）
func rollbackSteps(steps []OperationStep) error {
    var errors []error
    
    // 逆序回滚
    for i := len(steps) - 1; i >= 0; i-- {
        step := steps[i]
        if step.Rollback != nil {
            if err := step.Rollback(); err != nil {
                errors = append(errors, fmt.Errorf("rollback step %d failed: %v", i, err))
            }
        }
    }
    
    if len(errors) > 0 {
        return &MultiError{Errors: errors}
    }
    return nil
}
```

#### 5.3.2 操作日志与恢复
```go
// OperationLog 操作日志
type OperationLog struct {
    ID          string
    Operation   string
    User        string
    Resource    string
    Parameters  map[string]interface{}
    Result      interface{}
    Error       *StorageError
    Duration    time.Duration
    Timestamp   time.Time
    IP          string
    UserAgent   string
    
    // 恢复信息
    CanRecover  bool
    RecoveryFn  func() error
    Checkpoint  interface{}
}

// LogRecoverySystem 日志恢复系统
type LogRecoverySystem struct {
    logs    []OperationLog
    mu      sync.RWMutex
    storage LogStorage
}

// RecoverFailedOperations 恢复失败的操作
func (r *LogRecoverySystem) RecoverFailedOperations(ctx context.Context, since time.Time) (int, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    
    var recovered int
    var errors []error
    
    for _, log := range r.logs {
        if log.Timestamp.After(since) && log.Error != nil && log.CanRecover {
            if log.RecoveryFn != nil {
                if err := log.RecoveryFn(); err != nil {
                    errors = append(errors, fmt.Errorf("recovery failed for log %s: %v", log.ID, err))
                } else {
                    recovered++
                }
            }
        }
    }
    
    if len(errors) > 0 {
        return recovered, &MultiError{Errors: errors}
    }
    return recovered, nil
}
```

## 6. 扩展性设计

### 6.1 插件系统

#### 6.1.1 插件架构设计
Cloud Storage Tool 采用**完全插件化的架构**，所有存储提供商都以插件形式存在，支持动态加载和热插拔。

#### 6.1.2 插件发现机制
```go
// PluginDescriptor 插件描述符
type PluginDescriptor struct {
    Name           string                 // 插件名称（如 "cos"、"s3"）
    Version        string                 // 插件版本（语义化版本）
    Description    string                 // 插件描述
    ProviderType   string                 // 提供商类型（"cos"、"s3"、"oss"等）
    EntryPoint     string                 // 插件入口函数名
    Dependencies   []PluginDependency     // 依赖的其他插件
    Capabilities   ProviderCapabilities   // 插件能力声明
    Metadata       map[string]string      // 插件元数据
}

// PluginDependency 插件依赖
type PluginDependency struct {
    Name    string  // 依赖插件名称
    Version string  // 版本约束（如 ">=1.0.0, <2.0.0"）
}

// PluginRegistry 插件注册表
type PluginRegistry interface {
    // 插件发现
    DiscoverPlugins(ctx context.Context, searchPaths []string) ([]PluginDescriptor, error)
    RegisterPlugin(descriptor PluginDescriptor, factory ProviderFactory) error
    UnregisterPlugin(name string) error
    
    // 插件加载
    LoadPlugin(ctx context.Context, name string) (StorageProvider, error)
    UnloadPlugin(name string) error
    
    // 插件查询
    GetPlugin(name string) (*PluginDescriptor, error)
    ListPlugins() ([]PluginDescriptor, error)
    HasPlugin(name string) bool
}

// ProviderFactory 提供商工厂函数
type ProviderFactory func(ctx context.Context, config ProviderConfig) (StorageProvider, error)
```

#### 6.1.3 插件加载流程
1. **扫描阶段**：在预定义的插件目录中扫描插件文件
2. **验证阶段**：验证插件签名和完整性
3. **注册阶段**：解析插件描述符并注册到插件注册表
4. **初始化阶段**：按需加载插件，调用入口函数创建提供商实例

#### 6.1.4 插件目录结构
```
~/.cloud-storage/plugins/
├── cos-plugin/
│   ├── plugin.json        # 插件描述文件
│   ├── cos-provider.so    # 插件二进制（Linux）
│   ├── cos-provider.dylib # 插件二进制（macOS）
│   └── cos-provider.dll   # 插件二进制（Windows）
├── s3-plugin/
│   ├── plugin.json
│   └── s3-provider.so
└── oss-plugin/
    ├── plugin.json
    └── oss-provider.so
```

#### 6.1.5 插件描述文件示例（plugin.json）
```json
{
  "name": "cos",
  "version": "1.0.0",
  "description": "Tencent Cloud COS storage provider",
  "provider_type": "cos",
  "entry_point": "NewCOSProvider",
  "dependencies": [],
  "capabilities": {
    "supports_multipart_upload": true,
    "supports_presigned_url": true,
    "supports_versioning": true,
    "supports_lifecycle": true,
    "supports_server_side_encryption": true,
    "max_part_size": 5368709120,
    "max_object_size": 5497558138880
  },
  "metadata": {
    "vendor": "Tencent Cloud",
    "service_name": "COS",
    "documentation_url": "https://cloud.tencent.com/document/product/436"
  }
}
```

#### 6.1.6 动态加载实现
```go
// PluginLoader 插件加载器
type PluginLoader struct {
    registry PluginRegistry
    plugins  map[string]*plugin.Plugin
    mu       sync.RWMutex
}

// LoadPlugin 动态加载插件
func (l *PluginLoader) LoadPlugin(ctx context.Context, pluginPath string) error {
    l.mu.Lock()
    defer l.mu.Unlock()
    
    // 1. 打开插件文件
    p, err := plugin.Open(pluginPath)
    if err != nil {
        return fmt.Errorf("failed to open plugin %s: %v", pluginPath, err)
    }
    
    // 2. 查找描述符符号
    var descriptorSymbol interface{}
    if descriptorSymbol, err = p.Lookup("PluginDescriptor"); err != nil {
        return fmt.Errorf("plugin %s missing PluginDescriptor: %v", pluginPath, err)
    }
    
    // 3. 解析描述符
    descriptor, ok := descriptorSymbol.(*PluginDescriptor)
    if !ok {
        return fmt.Errorf("invalid plugin descriptor type in %s", pluginPath)
    }
    
    // 4. 查找工厂函数
    var factorySymbol interface{}
    if factorySymbol, err = p.Lookup(descriptor.EntryPoint); err != nil {
        return fmt.Errorf("plugin %s missing entry point %s: %v", 
            pluginPath, descriptor.EntryPoint, err)
    }
    
    factory, ok := factorySymbol.(ProviderFactory)
    if !ok {
        return fmt.Errorf("invalid factory function type in %s", pluginPath)
    }
    
    // 5. 注册插件
    if err := l.registry.RegisterPlugin(*descriptor, factory); err != nil {
        return fmt.Errorf("failed to register plugin %s: %v", descriptor.Name, err)
    }
    
    l.plugins[descriptor.Name] = p
    return nil
}
```

### 6.2 版本兼容性管理

#### 6.2.1 版本策略
- **语义化版本控制**：遵循 `MAJOR.MINOR.PATCH` 格式
- **向后兼容**：MINOR 版本增加保证向后兼容
- **破坏性变更**：MAJOR 版本增加表示有破坏性变更

#### 6.2.2 接口版本管理
```go
// VersionedInterface 版本化接口
type VersionedInterface struct {
    InterfaceName string    // 接口名称
    Version       string    // 接口版本（如 "v1.0.0"）
    Methods       []MethodInfo // 接口方法信息
    Deprecated    bool      // 是否已弃用
    Replacement   string    // 替代接口名称
}

// MethodInfo 方法信息
type MethodInfo struct {
    Name       string   // 方法名称
    Signature  string   // 方法签名
    Deprecated bool     // 是否已弃用
    Since      string   // 引入版本
    ChangedIn  string   // 修改版本
}

// InterfaceRegistry 接口注册表
type InterfaceRegistry interface {
    RegisterInterface(iface VersionedInterface) error
    GetInterface(name, version string) (*VersionedInterface, error)
    ListInterfaces() ([]VersionedInterface, error)
    CheckCompatibility(provider, required VersionedInterface) (bool, []string)
}
```

#### 6.2.3 插件版本兼容性检查
```go
// VersionCompatibilityChecker 版本兼容性检查器
type VersionCompatibilityChecker struct {
    registry InterfaceRegistry
}

// CheckPluginCompatibility 检查插件兼容性
func (c *VersionCompatibilityChecker) CheckPluginCompatibility(
    pluginDesc PluginDescriptor,
    requiredInterfaces map[string]string,
) (bool, []CompatibilityIssue) {
    var issues []CompatibilityIssue
    
    for ifaceName, requiredVersion := range requiredInterfaces {
        // 获取插件实现的接口版本
        pluginIface, err := c.registry.GetInterface(ifaceName, pluginDesc.InterfaceVersions[ifaceName])
        if err != nil {
            issues = append(issues, CompatibilityIssue{
                Interface: ifaceName,
                Issue:     fmt.Sprintf("plugin does not implement interface %s", ifaceName),
                Severity:  SeverityError,
            })
            continue
        }
        
        // 获取要求的接口版本
        requiredIface, err := c.registry.GetInterface(ifaceName, requiredVersion)
        if err != nil {
            issues = append(issues, CompatibilityIssue{
                Interface: ifaceName,
                Issue:     fmt.Sprintf("required interface %s version %s not found", ifaceName, requiredVersion),
                Severity:  SeverityWarning,
            })
            continue
        }
        
        // 检查兼容性
        compatible, methodIssues := c.registry.CheckCompatibility(pluginIface, requiredIface)
        if !compatible {
            for _, methodIssue := range methodIssues {
                issues = append(issues, CompatibilityIssue{
                    Interface: ifaceName,
                    Issue:     methodIssue,
                    Severity:  SeverityError,
                })
            }
        }
    }
    
    return len(issues) == 0, issues
}
```

#### 6.2.4 运行时版本适配
```go
// VersionAdapter 版本适配器
type VersionAdapter struct {
    adapters map[string]map[string]func(interface{}) interface{} // fromVersion -> toVersion -> adapter
}

// AdaptInterface 适配接口版本
func (a *VersionAdapter) AdaptInterface(
    instance interface{},
    fromVersion, toVersion string,
    interfaceName string,
) (interface{}, error) {
    
    // 查找适配器链
    adapterChain, err := a.findAdapterChain(fromVersion, toVersion, interfaceName)
    if err != nil {
        return nil, err
    }
    
    // 应用适配器链
    result := instance
    for _, adapter := range adapterChain {
        result = adapter(result)
    }
    
    return result, nil
}

// RegisterAdapter 注册版本适配器
func (a *VersionAdapter) RegisterAdapter(
    fromVersion, toVersion, interfaceName string,
    adapter func(interface{}) interface{},
) {
    if a.adapters[interfaceName] == nil {
        a.adapters[interfaceName] = make(map[string]func(interface{}) interface{})
    }
    
    key := fmt.Sprintf("%s->%s", fromVersion, toVersion)
    a.adapters[interfaceName][key] = adapter
}
```

#### 6.2.5 向后兼容性保证
1. **方法签名不变**：已发布接口的方法签名不会改变
2. **新增方法**：新增方法不影响现有功能
3. **可选接口**：高级功能通过可选接口实现
4. **版本探测**：运行时自动检测插件版本并适配
5. **优雅降级**：插件缺失某些接口时自动降级功能

### 6.3 中间件支持
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

#### 10.6.2 并发模式（增强版）

**改进的工作池模式**，支持上下文传播、取消机制和资源清理：

```go
// processConcurrently 并发处理任务（支持上下文取消）
func processConcurrently(tasks []Task, concurrency int, ctx context.Context) error {
    var wg sync.WaitGroup
    taskChan := make(chan Task, concurrency) // 缓冲通道提高性能
    errChan := make(chan error, len(tasks))
    
    // 启动固定数量工作协程
    for i := 0; i < concurrency; i++ {
        wg.Add(1)
        go func(workerID int) {
            defer wg.Done()
            for task := range taskChan {
                // 传播context以支持取消和超时
                taskCtx := context.WithValue(ctx, "workerID", workerID)
                if err := processTask(taskCtx, task); err != nil {
                    select {
                    case errChan <- err:
                        // 错误成功发送
                    case <-ctx.Done():
                        // 上下文被取消，停止发送错误
                        return
                    }
                }
            }
        }(i)
    }
    
    // 异步分发任务（支持取消）
    go func() {
        defer close(taskChan)
        for _, task := range tasks {
            select {
            case taskChan <- task:
                // 任务成功分发
            case <-ctx.Done():
                // 上下文被取消，停止分发
                return
            }
        }
    }()
    
    // 等待所有工作协程完成
    done := make(chan struct{})
    go func() {
        wg.Wait()
        close(errChan)
        close(done)
    }()
    
    // 等待完成或取消
    select {
    case <-done:
        // 正常完成，收集错误
        return collectErrors(errChan)
    case <-ctx.Done():
        // 被取消，关闭任务通道让工作协程退出
        close(taskChan)
        wg.Wait() // 等待工作协程清理
        return ctx.Err()
    }
}

// collectErrors 收集并合并错误
func collectErrors(errChan <-chan error) error {
    var errors []error
    for err := range errChan {
        errors = append(errors, err)
    }
    
    if len(errors) == 0 {
        return nil
    } else if len(errors) == 1 {
        return errors[0]
    }
    return &MultiError{Errors: errors}
}

// MultiError 多错误包装器
type MultiError struct {
    Errors []error
}

func (e *MultiError) Error() string {
    var sb strings.Builder
    sb.WriteString(fmt.Sprintf("%d errors occurred:\n", len(e.Errors)))
    for i, err := range e.Errors {
        sb.WriteString(fmt.Sprintf("\t%d. %v\n", i+1, err))
    }
    return sb.String()
}
```

#### 10.6.4 连接池与健康检查

**连接池实现**，支持健康检查和自动重建：

```go
// ConnectionPool 连接池接口
type ConnectionPool interface {
    Get(ctx context.Context) (interface{}, error)
    Put(conn interface{})
    Close() error
    Stats() PoolStats
}

// HTTPConnectionPool HTTP连接池实现
type HTTPConnectionPool struct {
    pool        *sync.Pool
    dialFunc    func() (interface{}, error)
    checkFunc   func(conn interface{}) bool
    maxIdle     int
    maxOpen     int
    idleTimeout time.Duration
    stats       *PoolStats
    mu          sync.RWMutex
    closed      bool
}

// PoolStats 连接池统计信息
type PoolStats struct {
    OpenConnections   int64
    IdleConnections   int64
    WaitCount         int64
    WaitDuration      time.Duration
    MaxIdleClosed     int64
    MaxLifetimeClosed int64
    HealthCheckPassed int64
    HealthCheckFailed int64
}

// Get 从连接池获取连接（带健康检查）
func (p *HTTPConnectionPool) Get(ctx context.Context) (interface{}, error) {
    p.mu.RLock()
    if p.closed {
        p.mu.RUnlock()
        return nil, ErrPoolClosed
    }
    p.mu.RUnlock()
    
    // 尝试从池中获取
    if conn := p.pool.Get(); conn != nil {
        // 健康检查
        if p.checkFunc(conn) {
            atomic.AddInt64(&p.stats.HealthCheckPassed, 1)
            return conn, nil
        }
        atomic.AddInt64(&p.stats.HealthCheckFailed, 1)
    }
    
    // 池中无可用连接或健康检查失败，创建新连接
    return p.dialFunc()
}

// Put 将连接放回池中
func (p *HTTPConnectionPool) Put(conn interface{}) {
    if conn == nil {
        return
    }
    
    p.mu.RLock()
    if p.closed {
        p.mu.RUnlock()
        return
    }
    
    // 检查连接是否仍然健康
    if !p.checkFunc(conn) {
        p.mu.RUnlock()
        return
    }
    
    p.mu.RUnlock()
    p.pool.Put(conn)
}
```

#### 10.6.5 进度追踪与监控

**并发安全的进度追踪器**：

```go
// ProgressTracker 进度追踪器
type ProgressTracker struct {
    total     int64
    completed int64
    failed    int64
    mu        sync.RWMutex
    startTime time.Time
}

// NewProgressTracker 创建进度追踪器
func NewProgressTracker(total int64) *ProgressTracker {
    return &ProgressTracker{
        total:     total,
        completed: 0,
        failed:    0,
        startTime: time.Now(),
    }
}

// IncrementCompleted 增加完成计数
func (p *ProgressTracker) IncrementCompleted() {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.completed++
}

// IncrementFailed 增加失败计数
func (p *ProgressTracker) IncrementFailed() {
    p.mu.Lock()
    defer p.mu.Unlock()
    p.failed++
}

// Progress 获取当前进度
func (p *ProgressTracker) Progress() (completed, failed, total int64, percentage float64) {
    p.mu.RLock()
    defer p.mu.RUnlock()
    
    completed = p.completed
    failed = p.failed
    total = p.total
    
    if total > 0 {
        percentage = float64(completed+failed) / float64(total) * 100
    }
    return
}

// ETA 计算预计剩余时间
func (p *ProgressTracker) ETA() time.Duration {
    p.mu.RLock()
    defer p.mu.RUnlock()
    
    if p.completed == 0 {
        return 0
    }
    
    elapsed := time.Since(p.startTime)
    avgTimePerTask := elapsed / time.Duration(p.completed)
    remainingTasks := p.total - p.completed - p.failed
    
    return avgTimePerTask * time.Duration(remainingTasks)
}
```

#### 10.6.6 同步原语
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
- **量化质量红线**：设置硬性代码质量指标

##### 量化质量红线（不可妥协）
以下指标必须在CI中强制执行，违反任何一条都将导致构建失败：

| 指标 | 限制值 | 检查工具 | 说明 |
|------|--------|----------|------|
| **单文件行数** | ≤ 500行 | `gocyclo` + 自定义检查 | 包括空行和注释，Go社区推荐标准 |
| **单函数行数** | ≤ 25行 | `funlen` | 函数体行数（不包括签名和注释），从30降至25 |
| **嵌套层数** | ≤ 3层 | `nestif` | if/for/switch等语句的嵌套深度 |
| **分支数量** | ≤ 4个 | `cyclop` | 函数中的条件分支数量，从3增至4更实际 |
| **单文件最大结构体数** | ≤ 5个 | 自定义检查 | 避免文件臃肿，强制模块化设计 |
| **最大导出符号数** | ≤ 20个 | `revive` | 控制包接口复杂度，促进封装 |
| **包依赖循环数** | = 0 | `gocyclo` | 强制无循环依赖，保持架构清晰 |
| **测试覆盖率变化** | ≤ -5% | `coverage` | 防止测试覆盖率退化 |

##### 新增关键质量指标说明
1. **单文件最大结构体数**：控制单个文件中定义的结构体数量，避免功能过于集中
2. **最大导出符号数**：限制包的公开接口数量，促进接口最小化和封装
3. **包依赖循环数**：强制消除循环依赖，这是Go项目常见的架构问题
4. **测试覆盖率变化**：防止在开发过程中测试覆盖率下降超过5%

##### 增强的golangci-lint配置示例
```yaml
# .golangci.yml
linters:
  enable:
    - funlen      # 函数长度检查
    - gocyclo     # 圈复杂度检查
    - nestif      # 嵌套if检查
    - cyclop      # 圈复杂度和分支检查
    - revive      # 导出符号检查
    - staticcheck # 静态分析
    - govet       # go vet检查

linters-settings:
  funlen:
    lines: 25           # 函数最多25行（从30降至25）
    statements: 20      # 函数最多20条语句（从25降至20）
  
  gocyclo:
    min-complexity: 8   # 圈复杂度警告阈值（从10降至8）
    # 结合自定义检查控制文件行数≤500行
  
  nestif:
    min-complexity: 3   # 嵌套复杂度阈值（从4降至3）
  
  cyclop:
    max-complexity: 4   # 最大圈复杂度（从3增至4）
    max-branches: 4     # 显式分支限制（新增）
    package-average: 2  # 包平均圈复杂度
  
  revive:
    rules:
      - name: max-public-structs
        arguments: [5]  # 单文件最多5个导出结构体
      - name: max-public-symbols
        arguments: [20] # 单包最多20个导出符号
      - name: package-comments
        severity: warning

# 自定义检查规则（通过脚本实现）
custom-checks:
  - name: "no-circular-deps"
    command: "go mod graph | awk '{print $1}' | sort | uniq -c | awk '$1 > 1 {exit 1}'"
    description: "检查包依赖循环"
  
  - name: "coverage-regression"
    command: "scripts/check-coverage-regression.sh"
    description: "检查测试覆盖率退化"
```

##### 质量红线执行机制
1. **CI集成**：在GitHub Actions中集成golangci-lint检查
2. **预提交钩子**：本地开发时通过git pre-commit钩子检查
3. **IDE集成**：配置IDE实时显示违规警告
4. **审查流程**：代码审查时必须验证质量红线
5. **例外审批**：极少数情况需要例外时，必须经过技术负责人审批并记录原因

##### 质量红线价值
- **可维护性**：小文件、短函数更易于理解和修改
- **可测试性**：简单的函数逻辑更易于编写单元测试
- **可读性**：减少嵌套和分支提高代码可读性
- **团队协作**：统一的质量标准降低协作成本

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

**文档版本**: 1.5  
**最后更新**: 2026-03-05  
**状态**: 草案