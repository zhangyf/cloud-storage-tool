// Package providers 提供腾讯云 COS 存储提供商实现（最小化版）
package providers

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/zhangyf/cloud-storage-tool/internal/storage"
)

// TencentCOSProvider 腾讯云 COS 存储提供商实现（最小化版）
type TencentCOSProvider struct {
	config storage.Config
	closed bool
}

// NewTencentCOSProvider 创建腾讯云 COS 存储提供商实例（最小化版）
func NewTencentCOSProvider(config storage.Config) (storage.StorageProvider, error) {
	// 验证配置
	if err := validateTencentCOSConfig(config); err != nil {
		return nil, fmt.Errorf("腾讯云 COS 配置验证失败: %w", err)
	}

	// 创建最小化版的提供商
	provider := &TencentCOSProvider{
		config: config,
		closed: false,
	}

	return provider, nil
}

// validateTencentCOSConfig 验证腾讯云 COS 配置
func validateTencentCOSConfig(config storage.Config) error {
	if config.Bucket == "" {
		return errors.New("腾讯云 COS 存储桶不能为空")
	}
	if config.Region == "" {
		return errors.New("腾讯云 COS 区域不能为空")
	}
	if config.Credentials.SecretID == "" || config.Credentials.SecretKey == "" {
		return errors.New("腾讯云 COS SecretID 和 SecretKey 不能为空")
	}
	return nil
}

// Upload 上传本地文件到腾讯云 COS（最小化版）
func (p *TencentCOSProvider) Upload(localPath, remotePath string) error {
	if p.closed {
		return errors.New("腾讯云 COS 提供商已关闭")
	}
	return errors.New("腾讯云 COS 上传功能未实现（最小化版）")
}

// UploadStream 从流上传数据到腾讯云 COS（最小化版）
func (p *TencentCOSProvider) UploadStream(reader io.Reader, remotePath string, size int64) error {
	if p.closed {
		return errors.New("腾讯云 COS 提供商已关闭")
	}
	return errors.New("腾讯云 COS 流式上传功能未实现（最小化版）")
}

// Download 从腾讯云 COS 下载文件到本地（最小化版）
func (p *TencentCOSProvider) Download(remotePath, localPath string) error {
	if p.closed {
		return errors.New("腾讯云 COS 提供商已关闭")
	}
	return errors.New("腾讯云 COS 下载功能未实现（最小化版）")
}

// DownloadStream 从腾讯云 COS 下载数据到流（最小化版）
func (p *TencentCOSProvider) DownloadStream(remotePath string) (io.ReadCloser, error) {
	if p.closed {
		return nil, errors.New("腾讯云 COS 提供商已关闭")
	}
	return nil, errors.New("腾讯云 COS 流式下载功能未实现（最小化版）")
}

// List 列出指定前缀的对象（最小化版）
func (p *TencentCOSProvider) List(prefix string) ([]storage.FileInfo, error) {
	if p.closed {
		return nil, errors.New("腾讯云 COS 提供商已关闭")
	}
	return []storage.FileInfo{}, nil
}

// Delete 删除腾讯云 COS 中的对象（最小化版）
func (p *TencentCOSProvider) Delete(remotePath string) error {
	if p.closed {
		return errors.New("腾讯云 COS 提供商已关闭")
	}
	return errors.New("腾讯云 COS 删除功能未实现（最小化版）")
}

// DeleteBatch 批量删除腾讯云 COS 中的对象（最小化版）
func (p *TencentCOSProvider) DeleteBatch(paths []string) error {
	if p.closed {
		return errors.New("腾讯云 COS 提供商已关闭")
	}
	return errors.New("腾讯云 COS 批量删除功能未实现（最小化版）")
}

// DeleteMultiple 批量删除腾讯云 COS 中的对象（最小化版）- 别名方法
func (p *TencentCOSProvider) DeleteMultiple(paths []string) error {
	return p.DeleteBatch(paths)
}

// Stat 获取对象元数据（最小化版）
func (p *TencentCOSProvider) Stat(remotePath string) (storage.FileInfo, error) {
	if p.closed {
		return storage.FileInfo{}, errors.New("腾讯云 COS 提供商已关闭")
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
func (p *TencentCOSProvider) Exists(remotePath string) (bool, error) {
	if p.closed {
		return false, errors.New("腾讯云 COS 提供商已关闭")
	}
	return false, nil
}

// GeneratePresignedURL 生成预签名 URL（最小化版）
func (p *TencentCOSProvider) GeneratePresignedURL(remotePath string, expires time.Duration) (string, error) {
	if p.closed {
		return "", errors.New("腾讯云 COS 提供商已关闭")
	}
	return "", errors.New("腾讯云 COS 预签名URL功能未实现（最小化版）")
}

// GetURL 获取对象URL（最小化版）
func (p *TencentCOSProvider) GetURL(remotePath string, expires time.Duration) (string, error) {
	if p.closed {
		return "", errors.New("腾讯云 COS 提供商已关闭")
	}
	return "", errors.New("腾讯云 COS 获取URL功能未实现（最小化版）")
}

// Copy 复制文件（最小化版）
func (p *TencentCOSProvider) Copy(srcPath, dstPath string) error {
	if p.closed {
		return errors.New("腾讯云 COS 提供商已关闭")
	}
	return errors.New("腾讯云 COS 复制功能未实现（最小化版）")
}

// Move 移动文件（最小化版）
func (p *TencentCOSProvider) Move(srcPath, dstPath string) error {
	if p.closed {
		return errors.New("腾讯云 COS 提供商已关闭")
	}
	return errors.New("腾讯云 COS 移动功能未实现（最小化版）")
}

// Close 关闭腾讯云 COS 提供商（最小化版）
func (p *TencentCOSProvider) Close() error {
	if p.closed {
		return nil
	}
	p.closed = true
	return nil
}

// GetConfig 获取配置信息（最小化版）
func (p *TencentCOSProvider) GetConfig() storage.Config {
	return p.config
}

// ProviderName 获取提供商名称（最小化版）
func (p *TencentCOSProvider) ProviderName() string {
	return "tencent_cos"
}