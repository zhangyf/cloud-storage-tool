// Package providers 提供 AWS S3 存储提供商实现（最小化版）
package providers

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/zhangyf/cloud-storage-tool/internal/storage"
)

// AWSS3Provider AWS S3 存储提供商实现（最小化版）
type AWSS3Provider struct {
	config storage.Config
	closed bool
}

// NewAWSS3Provider 创建 AWS S3 存储提供商实例（最小化版）
func NewAWSS3Provider(cfg storage.Config) (storage.StorageProvider, error) {
	// 验证配置
	if err := validateAWSS3Config(cfg); err != nil {
		return nil, fmt.Errorf("AWS S3 配置验证失败: %w", err)
	}

	// 创建最小化版的提供商
	provider := &AWSS3Provider{
		config: cfg,
		closed: false,
	}

	return provider, nil
}

// validateAWSS3Config 验证 AWS S3 配置
func validateAWSS3Config(config storage.Config) error {
	if config.Bucket == "" {
		return errors.New("AWS S3 存储桶不能为空")
	}
	if config.Region == "" {
		return errors.New("AWS S3 区域不能为空")
	}
	if config.Credentials.AWSAccessKeyID == "" || config.Credentials.AWSSecretAccessKey == "" {
		return errors.New("AWS S3 AccessKeyID 和 SecretAccessKey 不能为空")
	}
	return nil
}

// Upload 上传本地文件到 AWS S3（最小化版）
func (p *AWSS3Provider) Upload(localPath, remotePath string) error {
	if p.closed {
		return errors.New("AWS S3 提供商已关闭")
	}
	return errors.New("AWS S3 上传功能未实现（最小化版）")
}

// UploadStream 从流上传数据到 AWS S3（最小化版）
func (p *AWSS3Provider) UploadStream(reader io.Reader, remotePath string, size int64) error {
	if p.closed {
		return errors.New("AWS S3 提供商已关闭")
	}
	return errors.New("AWS S3 流式上传功能未实现（最小化版）")
}

// Download 从 AWS S3 下载文件到本地（最小化版）
func (p *AWSS3Provider) Download(remotePath, localPath string) error {
	if p.closed {
		return errors.New("AWS S3 提供商已关闭")
	}
	return errors.New("AWS S3 下载功能未实现（最小化版）")
}

// DownloadStream 从 AWS S3 下载数据到流（最小化版）
func (p *AWSS3Provider) DownloadStream(remotePath string) (io.ReadCloser, error) {
	if p.closed {
		return nil, errors.New("AWS S3 提供商已关闭")
	}
	return nil, errors.New("AWS S3 流式下载功能未实现（最小化版）")
}

// List 列出指定前缀的对象（最小化版）
func (p *AWSS3Provider) List(prefix string) ([]storage.FileInfo, error) {
	if p.closed {
		return nil, errors.New("AWS S3 提供商已关闭")
	}
	return []storage.FileInfo{}, nil
}

// Delete 删除 AWS S3 中的对象（最小化版）
func (p *AWSS3Provider) Delete(remotePath string) error {
	if p.closed {
		return errors.New("AWS S3 提供商已关闭")
	}
	return errors.New("AWS S3 删除功能未实现（最小化版）")
}

// DeleteBatch 批量删除 AWS S3 中的对象（最小化版）
func (p *AWSS3Provider) DeleteBatch(paths []string) error {
	if p.closed {
		return errors.New("AWS S3 提供商已关闭")
	}
	return errors.New("AWS S3 批量删除功能未实现（最小化版）")
}

// DeleteMultiple 批量删除 AWS S3 中的对象（最小化版）- 别名方法
func (p *AWSS3Provider) DeleteMultiple(paths []string) error {
	return p.DeleteBatch(paths)
}

// Stat 获取对象元数据（最小化版）
func (p *AWSS3Provider) Stat(remotePath string) (storage.FileInfo, error) {
	if p.closed {
		return storage.FileInfo{}, errors.New("AWS S3 提供商已关闭")
	}
	return storage.FileInfo{
		Name:         remotePath,
		Size:         0,
		LastModified: time.Now(),
		ETag:         "",
		StorageClass: "STANDARD",
		IsDir:        false,
	}, nil
}

// Exists 检查对象是否存在（最小化版）
func (p *AWSS3Provider) Exists(remotePath string) (bool, error) {
	if p.closed {
		return false, errors.New("AWS S3 提供商已关闭")
	}
	return false, nil
}

// GeneratePresignedURL 生成预签名 URL（最小化版）
func (p *AWSS3Provider) GeneratePresignedURL(remotePath string, expires time.Duration) (string, error) {
	if p.closed {
		return "", errors.New("AWS S3 提供商已关闭")
	}
	return "", errors.New("AWS S3 预签名URL功能未实现（最小化版）")
}

// GetURL 获取对象URL（最小化版）
func (p *AWSS3Provider) GetURL(remotePath string, expires time.Duration) (string, error) {
	if p.closed {
		return "", errors.New("AWS S3 提供商已关闭")
	}
	return "", errors.New("AWS S3 获取URL功能未实现（最小化版）")
}

// Copy 复制文件（最小化版）- 新增方法
func (p *AWSS3Provider) Copy(srcPath, dstPath string) error {
	if p.closed {
		return errors.New("AWS S3 提供商已关闭")
	}
	return errors.New("AWS S3 复制功能未实现（最小化版）")
}

// Move 移动文件（最小化版）- 新增方法
func (p *AWSS3Provider) Move(srcPath, dstPath string) error {
	if p.closed {
		return errors.New("AWS S3 提供商已关闭")
	}
	return errors.New("AWS S3 移动功能未实现（最小化版）")
}

// Close 关闭 AWS S3 提供商（最小化版）
func (p *AWSS3Provider) Close() error {
	if p.closed {
		return nil
	}
	p.closed = true
	return nil
}

// GetConfig 获取配置信息（最小化版）
func (p *AWSS3Provider) GetConfig() storage.Config {
	return p.config
}

// ProviderName 获取提供商名称（最小化版）
func (p *AWSS3Provider) ProviderName() string {
	return "aws_s3"
}