// Package providers 提供阿里云 OSS 存储提供商实现
package providers

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/zhangyf/cloud-storage-tool/internal/storage"
	"github.com/zhangyf/cloud-storage-tool/internal/utils"
)

// AliyunOSSProvider 阿里云 OSS 存储提供商实现
type AliyunOSSProvider struct {
	config     storage.Config
	client     *oss.Client
	bucket     *oss.Bucket
	logger     *utils.Logger
	closed     bool
}

// NewAliyunOSSProvider 创建阿里云 OSS 存储提供商实例
func NewAliyunOSSProvider(config storage.Config) (storage.StorageProvider, error) {
	// 验证配置
	if config.Bucket == "" {
		return nil, utils.NewErrorBuilder(utils.ErrorTypeConfig).
			WithCode(utils.ErrCodeConfigInvalid).
			WithMessage("阿里云 OSS 存储桶名称不能为空").
			Build()
	}

	if config.Endpoint == "" {
		return nil, utils.NewErrorBuilder(utils.ErrorTypeConfig).
			WithCode(utils.ErrCodeConfigInvalid).
			WithMessage("阿里云 OSS 端点不能为空").
			Build()
	}

	if config.Credentials.AccessKeyID == "" || config.Credentials.AccessKeySecret == "" {
		return nil, utils.NewErrorBuilder(utils.ErrorTypeConfig).
			WithCode(utils.ErrCodeConfigInvalid).
			WithMessage("阿里云 OSS AccessKeyID 和 AccessKeySecret 不能为空").
			Build()
	}

	// 创建 OSS 客户端
	client, err := oss.New(config.Endpoint, config.Credentials.AccessKeyID, config.Credentials.AccessKeySecret)
	if err != nil {
		return nil, utils.Wrapf(err, utils.ErrorTypeConfig, utils.ErrCodeConfigParseFailed,
			"创建阿里云 OSS 客户端失败: %v", err)
	}

	// 获取存储桶
	bucket, err := client.Bucket(config.Bucket)
	if err != nil {
		return nil, utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
			"获取阿里云 OSS 存储桶失败: %s", config.Bucket)
	}

	// 创建日志记录器
	logger := utils.GetLogger().WithField("provider", "aliyun_oss")

	logger.Info("阿里云 OSS 提供商初始化成功，存储桶: %s, 端点: %s", config.Bucket, config.Endpoint)

	return &AliyunOSSProvider{
		config:   config,
		client:   client,
		bucket:   bucket,
		logger:   logger,
		closed:   false,
	}, nil
}

// Upload 上传本地文件到阿里云 OSS
func (p *AliyunOSSProvider) Upload(localPath, remotePath string) error {
	if p.closed {
		return utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "阿里云 OSS 提供商已关闭")
	}

	p.logger.Info("上传文件: %s -> %s", localPath, remotePath)

	// 上传文件
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	err := p.bucket.PutObjectFromFileWithContext(ctx, remotePath, localPath)
	if err != nil {
		return utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
			"上传到阿里云 OSS 失败: %s", remotePath)
	}

	p.logger.Info("上传完成: %s -> %s", localPath, remotePath)
	return nil
}

// UploadStream 从流上传数据到阿里云 OSS
func (p *AliyunOSSProvider) UploadStream(reader io.Reader, remotePath string, size int64) error {
	if p.closed {
		return utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "阿里云 OSS 提供商已关闭")
	}

	p.logger.Info("流式上传: %s", remotePath)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	err := p.bucket.PutObjectWithContext(ctx, remotePath, reader)
	if err != nil {
		return utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
			"流式上传失败: %s", remotePath)
	}

	p.logger.Debug("流式上传完成: %s", remotePath)
	return nil
}

// Download 从阿里云 OSS 下载文件到本地
func (p *AliyunOSSProvider) Download(remotePath, localPath string) error {
	if p.closed {
		return utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "阿里云 OSS 提供商已关闭")
	}

	p.logger.Info("下载文件: %s -> %s", remotePath, localPath)

	// 创建本地文件
	file, err := os.Create(localPath)
	if err != nil {
		return utils.Wrapf(err, utils.ErrorTypeIO, utils.ErrCodeFileWriteFailed,
			"创建本地文件失败: %s", localPath)
	}
	defer file.Close()

	// 下载对象
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	err = p.bucket.GetObjectToFileWithContext(ctx, remotePath, localPath)
	if err != nil {
		return utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageNotFound,
			"下载阿里云 OSS 对象失败: %s", remotePath)
	}

	p.logger.Info("下载完成: %s -> %s", remotePath, localPath)
	return nil
}

// DownloadStream 从阿里云 OSS 下载数据到流
func (p *AliyunOSSProvider) DownloadStream(remotePath string) (io.ReadCloser, error) {
	if p.closed {
		return nil, utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "阿里云 OSS 提供商已关闭")
	}

	p.logger.Debug("流式下载: %s", remotePath)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	reader, err := p.bucket.GetObjectWithContext(ctx, remotePath)
	if err != nil {
		return nil, utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageNotFound,
			"获取阿里云 OSS 对象失败: %s", remotePath)
	}

	// 包装读取器以支持取消
	streamReader := &streamReaderWithCancel{
		ReadCloser: reader,
		cancel:     cancel,
	}

	return streamReader, nil
}

// List 列出阿里云 OSS 中指定前缀的对象
func (p *AliyunOSSProvider) List(prefix string) ([]storage.FileInfo, error) {
	if p.closed {
		return nil, utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "阿里云 OSS 提供商已关闭")
	}

	p.logger.Debug("列出对象: prefix=%s", prefix)

	// 列出对象
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	marker := ""
	var objects []storage.FileInfo

	for {
		result, err := p.bucket.ListObjectsV2(oss.Prefix(prefix), oss.Marker(marker))
		if err != nil {
			return nil, utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
				"列出阿里云 OSS 对象失败")
		}

		// 解析结果
		for _, obj := range result.Objects {
			fileInfo := storage.FileInfo{
				Name:         obj.Key,
				Size:         obj.Size,
				LastModified: obj.LastModified,
				ETag:         strings.Trim(obj.ETag, "\""),
				StorageClass: string(obj.StorageClass),
				IsDir:        false,
			}
			objects = append(objects, fileInfo)
		}

		// 检查是否还有更多对象
		if !result.IsTruncated {
			break
		}

		marker = result.NextMarker
	}

	p.logger.Debug("找到 %d 个对象", len(objects))
	return objects, nil
}

// Delete 删除阿里云 OSS 中的对象
func (p *AliyunOSSProvider) Delete(path string) error {
	if p.closed {
		return utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "阿里云 OSS 提供商已关闭")
	}

	p.logger.Debug("删除对象: %s", path)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	err := p.bucket.DeleteObjectWithContext(ctx, path)
	if err != nil {
		return utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
			"删除阿里云 OSS 对象失败: %s", path)
	}

	p.logger.Debug("删除完成: %s", path)
	return nil
}

// DeleteMultiple 批量删除阿里云 OSS 中的对象
func (p *AliyunOSSProvider) DeleteMultiple(paths []string) error {
	if p.closed {
		return utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "阿里云 OSS 提供商已关闭")
	}

	p.logger.Debug("批量删除: %d 个对象", len(paths))

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	_, err := p.bucket.DeleteObjectsWithContext(ctx, paths)
	if err != nil {
		return utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
			"批量删除阿里云 OSS 对象失败")
	}

	return nil
}

// Stat 获取阿里云 OSS 中对象的元数据信息
func (p *AliyunOSSProvider) Stat(path string) (storage.FileInfo, error) {
	if p.closed {
		return storage.FileInfo{}, utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "阿里云 OSS 提供商已关闭")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	// 获取对象属性
	prop, err := p.bucket.GetObjectDetailedMetaWithContext(ctx, path)
	if err != nil {
		return storage.FileInfo{}, utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
			"获取对象元数据失败: %s", path)
	}

	// 获取对象大小
	prop2, err := p.bucket.GetObjectMeta(path)
	if err != nil {
		return storage.FileInfo{}, utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
			"获取对象大小失败: %s", path)
	}

	size, _ := prop2.ContentLength()

	fileInfo := storage.FileInfo{
		Name:         path,
		Size:         size,
		LastModified: prop.LastModified,
		ETag:         strings.Trim(prop.ETag, "\""),
		StorageClass: string(prop.StorageClass),
		IsDir:        false,
	}

	p.logger.Debug("获取对象信息: %s (大小: %d, 修改时间: %v)", path, size, prop.LastModified)
	return fileInfo, nil
}

// Copy 复制阿里云 OSS 中的对象
func (p *AliyunOSSProvider) Copy(srcPath, dstPath string) error {
	if p.closed {
		return utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "阿里云 OSS 提供商已关闭")
	}

	// 检查源对象是否存在
	exists, err := p.Exists(srcPath)
	if err != nil {
		return utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageNotFound,
			"检查源对象是否存在失败: %s", srcPath)
	}

	if !exists {
		return utils.NewErrorBuilder(utils.ErrorTypeStorage).
			WithCode(utils.ErrCodeStorageNotFound).
			WithMessage("源对象不存在").
			WithContext("src_path", srcPath).
			Build()
	}

	p.logger.Info("复制对象: %s -> %s", srcPath, dstPath)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	// 在阿里云 OSS 中，复制是通过 CopyObject 实现的
	err = p.bucket.CopyObjectFrom(srcPath, p.config.Bucket, dstPath)
	if err != nil {
		return utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
			"复制阿里云 OSS 对象失败: %s -> %s", srcPath, dstPath)
	}

	p.logger.Info("复制完成: %s -> %s", srcPath, dstPath)
	return nil
}

// Move 移动阿里云 OSS 中的对象
func (p *AliyunOSSProvider) Move(srcPath, dstPath string) error {
	if p.closed {
		return utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "阿里云 OSS 提供商已关闭")
	}

	// 先复制
	if err := p.Copy(srcPath, dstPath); err != nil {
		return utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
			"移动对象失败（复制阶段）: %s -> %s", srcPath, dstPath)
	}

	// 再删除源
	if err := p.Delete(srcPath); err != nil {
		// 如果删除失败，尝试回滚（删除目标）
		p.logger.Error("移动对象失败（删除源阶段），尝试回滚: %s -> %s", srcPath, dstPath)
		if deleteErr := p.Delete(dstPath); deleteErr != nil {
			p.logger.Error("回滚失败: %v", deleteErr)
		}
		return utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
			"移动对象失败（删除源阶段）: %s -> %s", srcPath, dstPath)
	}

	p.logger.Info("移动完成: %s -> %s", srcPath, dstPath)
	return nil
}

// Exists 检查阿里云 OSS 中的对象是否存在
func (p *AliyunOSSProvider) Exists(path string) (bool, error) {
	if p.closed {
		return false, utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "阿里云 OSS 提供商已关闭")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 尝试获取对象属性
	_, err := p.bucket.GetObjectDetailedMetaWithContext(ctx, path)
	if err != nil {
		// 检查是否是404错误
		if strings.Contains(err.Error(), "NotFound") || strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
			"检查对象是否存在失败: %s", path)
	}

	return true, nil
}

// GetURL 获取阿里云 OSS 中对象的访问URL
func (p *AliyunOSSProvider) GetURL(path string, expires time.Duration) (string, error) {
	if p.closed {
		return "", utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "阿里云 OSS 提供商已关闭")
	}

	// 检查对象是否存在
	exists, err := p.Exists(path)
	if err != nil {
		return "", utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageNotFound,
			"检查对象是否存在失败: %s", path)
	}

	if !exists {
		return "", utils.NewErrorBuilder(utils.ErrorTypeStorage).
			WithCode(utils.ErrCodeStorageNotFound).
			WithMessage("对象不存在").
			WithContext("path", path).
			Build()
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	// 生成签名URL
	url, err := p.bucket.SignURL(path, oss.HTTPGet, int(expires.Seconds()))
	if err != nil {
		return "", utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
			"生成预签名URL失败: %s", path)
	}

	p.logger.Debug("生成预签名URL: %s (有效期: %v)", path, expires)
	return url, nil
}

// Close 关闭阿里云 OSS 提供商连接
func (p *AliyunOSSProvider) Close() error {
	if p.closed {
		return nil
	}

	p.logger.Info("关闭阿里云 OSS 提供商连接")
	p.closed = true

	// OSS 客户端没有显式的关闭方法
	// 这里主要是标记为已关闭状态
	return nil
}

// ProviderName 返回提供商名称
func (p *AliyunOSSProvider) ProviderName() string {
	return "阿里云 OSS"
}

// streamReaderWithCancel 包装读取器，在关闭时取消上下文
type streamReaderWithCancel struct {
	io.ReadCloser
	cancel context.CancelFunc
}

// Close 关闭读取器并取消上下文
func (src *streamReaderWithCancel) Close() error {
	defer src.cancel()
	return src.ReadCloser.Close()
}
