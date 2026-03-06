// Package storage 提供云存储的抽象接口
package storage

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/zhangyf/cloud-storage-tool/internal/providers"
)

// FileInfo 表示文件或对象的元数据信息
type FileInfo struct {
	// Name 文件或对象的名称
	Name string
	
	// Size 文件大小（字节）
	Size int64
	
	// LastModified 最后修改时间
	LastModified time.Time
	
	// ETag 对象的ETag（用于版本控制）
	ETag string
	
	// StorageClass 存储类型
	StorageClass string
	
	// IsDir 是否为目录（对于支持目录的存储）
	IsDir bool
}

// StorageProvider 定义云存储提供商的统一接口
type StorageProvider interface {
	// Upload 上传本地文件到云存储
	Upload(localPath, remotePath string) error
	
	// UploadStream 从流上传数据到云存储
	UploadStream(reader io.Reader, remotePath string, size int64) error
	
	// Download 从云存储下载文件到本地
	Download(remotePath, localPath string) error
	
	// DownloadStream 从云存储下载数据到流
	DownloadStream(remotePath string) (io.ReadCloser, error)
	
	// List 列出指定前缀的对象
	List(prefix string) ([]FileInfo, error)
	
	// Delete 删除云存储中的对象
	Delete(path string) error
	
	// DeleteMultiple 批量删除对象
	DeleteMultiple(paths []string) error
	
	// Stat 获取对象的元数据信息
	Stat(path string) (FileInfo, error)
	
	// Copy 复制云存储中的对象
	Copy(srcPath, dstPath string) error
	
	// Move 移动云存储中的对象
	Move(srcPath, dstPath string) error
	
	// Exists 检查对象是否存在
	Exists(path string) (bool, error)
	
	// GetURL 获取对象的访问URL
	GetURL(path string, expires time.Duration) (string, error)
	
	// Close 关闭存储提供商连接
	Close() error
	
	// ProviderName 返回提供商名称
	ProviderName() string
}

// ProviderType 提供商类型枚举
type ProviderType string

const (
	ProviderTencentCOS ProviderType = "tencent_cos"
	ProviderAliyunOSS  ProviderType = "aliyun_oss"
	ProviderAWSS3      ProviderType = "aws_s3"
)

// Config 存储提供商配置
type Config struct {
	// Type 提供商类型
	Type ProviderType
	
	// Bucket 存储桶名称
	Bucket string
	
	// Region 区域
	Region string
	
	// Endpoint 端点（对于阿里云OSS）
	Endpoint string
	
	// Credentials 认证信息
	Credentials Credentials
	
	// Timeout 操作超时时间（秒）
	Timeout int
	
	// MaxRetries 最大重试次数
	MaxRetries int
	
	// EnableSSL 是否启用SSL
	EnableSSL bool
	
	// EnableDebug 是否启用调试模式
	EnableDebug bool
}

// Credentials 认证信息
type Credentials struct {
	// 腾讯云COS
	SecretID  string
	SecretKey string
	
	// 阿里云OSS
	AccessKeyID     string
	AccessKeySecret string
	
	// AWS S3
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSSessionToken    string
}

// NewProvider 根据配置创建存储提供商实例
func NewProvider(config Config) (StorageProvider, error) {
	switch config.Type {
	case ProviderTencentCOS:
		return providers.NewTencentCOSProvider(config)
	case ProviderAliyunOSS:
		return providers.NewAliyunOSSProvider(config)
	case ProviderAWSS3:
		return providers.NewAWSS3Provider(config)
	default:
		return nil, fmt.Errorf("不支持的存储提供商类型: %s", config.Type)
	}
}

// Common errors
var (
	ErrNotFound        = errors.New("对象不存在")
	ErrAccessDenied    = errors.New("访问被拒绝")
	ErrBucketNotFound  = errors.New("存储桶不存在")
	ErrInvalidPath     = errors.New("路径无效")
	ErrUploadFailed    = errors.New("上传失败")
	ErrDownloadFailed  = errors.New("下载失败")
	ErrDeleteFailed    = errors.New("删除失败")
	ErrCopyFailed      = errors.New("复制失败")
	ErrMoveFailed      = errors.New("移动失败")
	ErrConfigInvalid   = errors.New("配置无效")
	ErrProviderClosed  = errors.New("存储提供商已关闭")
	ErrOperationTimeout = errors.New("操作超时")
)

// IsNotFoundError 检查错误是否为"未找到"错误
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsAccessDeniedError 检查错误是否为"访问被拒绝"错误
func IsAccessDeniedError(err error) bool {
	return errors.Is(err, ErrAccessDenied)
}

// OperationResult 操作结果
type OperationResult struct {
	// Success 是否成功
	Success bool
	
	// Message 结果消息
	Message string
	
	// Error 错误信息（如果有）
	Error error
	
	// Duration 操作耗时
	Duration time.Duration
	
	// BytesTransferred 传输的字节数
	BytesTransferred int64
}

// ProgressHandler 进度处理函数类型
type ProgressHandler func(current, total int64)

// WithProgress 返回一个包装了进度处理的读取器
type WithProgress struct {
	Reader   io.Reader
	Total    int64
	Current  int64
	Handler  ProgressHandler
}

// Read 实现io.Reader接口
func (wp *WithProgress) Read(p []byte) (int, error) {
	n, err := wp.Reader.Read(p)
	if n > 0 {
		wp.Current += int64(n)
		if wp.Handler != nil {
			wp.Handler(wp.Current, wp.Total)
		}
	}
	return n, err
}

// NewProgressReader 创建带进度处理的读取器
func NewProgressReader(reader io.Reader, total int64, handler ProgressHandler) io.Reader {
	return &WithProgress{
		Reader:  reader,
		Total:   total,
		Handler: handler,
	}
}

// ProviderFactory 提供商工厂接口
type ProviderFactory interface {
	// Create 创建存储提供商实例
	Create(config Config) (StorageProvider, error)
	
	// SupportedTypes 返回支持的提供商类型
	SupportedTypes() []ProviderType
}

// DefaultProviderFactory 默认提供商工厂
type DefaultProviderFactory struct{}

// Create 创建存储提供商实例
func (f *DefaultProviderFactory) Create(config Config) (StorageProvider, error) {
	return NewProvider(config)
}

// SupportedTypes 返回支持的提供商类型
func (f *DefaultProviderFactory) SupportedTypes() []ProviderType {
	return []ProviderType{
		ProviderTencentCOS,
		ProviderAliyunOSS,
		ProviderAWSS3,
	}
}

// GlobalProviderFactory 全局提供商工厂实例
var GlobalProviderFactory ProviderFactory = &DefaultProviderFactory{}

// RegisterProviderFactory 注册自定义提供商工厂
func RegisterProviderFactory(factory ProviderFactory) {
	GlobalProviderFactory = factory
}