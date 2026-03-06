// Package providers 提供腾讯云 COS 存储提供商实现
package providers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/zhangyf/cloud-storage-tool/internal/storage"
	"github.com/zhangyf/cloud-storage-tool/internal/utils"

	cos "github.com/tencentyun/cos-go-sdk-v5"
)

// TencentCOSProvider 腾讯云 COS 存储提供商实现
type TencentCOSProvider struct {
	config storage.Config
	client *cos.Client
	logger *utils.Logger
	closed bool
}

// NewTencentCOSProvider 创建腾讯云 COS 存储提供商实例
func NewTencentCOSProvider(config storage.Config) (storage.StorageProvider, error) {
	// 验证配置
	if config.Bucket == "" {
		return nil, utils.NewErrorBuilder(utils.ErrorTypeConfig).
			WithCode(utils.ErrCodeConfigInvalid).
			WithMessage("腾讯云 COS 存储桶名称不能为空").
			Build()
	}

	if config.Region == "" {
		return nil, utils.NewErrorBuilder(utils.ErrorTypeConfig).
			WithCode(utils.ErrCodeConfigInvalid).
			WithMessage("腾讯云 COS 区域不能为空").
			Build()
	}

	if config.Credentials.SecretID == "" || config.Credentials.SecretKey == "" {
		return nil, utils.NewErrorBuilder(utils.ErrorTypeConfig).
			WithCode(utils.ErrCodeConfigInvalid).
			WithMessage("腾讯云 COS SecretID 和 SecretKey 不能为空").
			Build()
	}

	// 创建 COS 客户端
	u := fmt.Sprintf("https://cos.%s.myqcloud.com", config.Region)
	p := cos.NewPipeline(
		cos.NewAuth(config.Credentials.SecretID, config.Credentials.SecretKey),
		time.Duration(config.Timeout)*time.Second,
	)
	client := cos.NewClient(u, p)

	// 创建日志记录器
	logger := utils.GetLogger().WithField("provider", "tencent_cos")

	logger.Info("腾讯云 COS 提供商初始化成功，存储桶: %s, 区域: %s", config.Bucket, config.Region)

	return &TencentCOSProvider{
		config: config,
		client: client,
		logger: logger,
		closed: false,
	}, nil
}

// Upload 上传本地文件到腾讯云 COS
func (p *TencentCOSProvider) Upload(localPath, remotePath string) error {
	if p.closed {
		return utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "腾讯云 COS 提供商已关闭")
	}

	p.logger.Info("上传文件: %s -> %s", localPath, remotePath)

	// 打开本地文件
	file, err := os.Open(localPath)
	if err != nil {
		return utils.Wrapf(err, utils.ErrorTypeIO, utils.ErrCodeFileNotFound,
			"打开本地文件失败: %s", localPath)
	}
	defer file.Close()

	// 上传对象
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	_, _, err = p.client.Object.Upload(ctx, remotePath, file, nil)
	if err != nil {
		return utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
			"上传到腾讯云 COS 失败: %s", remotePath)
	}

	p.logger.Info("上传完成: %s -> %s", localPath, remotePath)
	return nil
}

// UploadStream 从流上传数据到腾讯云 COS
func (p *TencentCOSProvider) UploadStream(reader io.Reader, remotePath string, size int64) error {
	if p.closed {
		return utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "腾讯云 COS 提供商已关闭")
	}

	p.logger.Info("流式上传: %s", remotePath)

	// 创建缓冲读取器
	readerWithCloser := struct {
		io.Reader
		io.Closer
	}{
		Reader: reader,
		Closer: ioutil.NopCloser(io.NopCloser(nil)),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	_, _, err := p.client.Object.Upload(ctx, remotePath, readerWithCloser, nil)
	if err != nil {
		return utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
			"流式上传失败: %s", remotePath)
	}

	p.logger.Debug("流式上传完成: %s", remotePath)
	return nil
}

// Download 从腾讯云 COS 下载文件到本地
func (p *TencentCOSProvider) Download(remotePath, localPath string) error {
	if p.closed {
		return utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "腾讯云 COS 提供商已关闭")
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

	resp, err := p.client.Object.Get(ctx, remotePath, nil)
	if err != nil {
		return utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageNotFound,
			"下载腾讯云 COS 对象失败: %s", remotePath)
	}
	defer resp.Body.Close()

	// 复制内容到文件
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return utils.Wrapf(err, utils.ErrorTypeIO, utils.ErrCodeFileWriteFailed,
			"写入本地文件失败: %s", localPath)
	}

	p.logger.Info("下载完成: %s -> %s", remotePath, localPath)
	return nil
}

// DownloadStream 从腾讯云 COS 下载数据到流
func (p *TencentCOSProvider) DownloadStream(remotePath string) (io.ReadCloser, error) {
	if p.closed {
		return nil, utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "腾讯云 COS 提供商已关闭")
	}

	p.logger.Debug("流式下载: %s", remotePath)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	resp, err := p.client.Object.Get(ctx, remotePath, nil)
	if err != nil {
		return nil, utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageNotFound,
			"获取腾讯云 COS 对象失败: %s", remotePath)
	}

	// 包装读取器以支持取消
	reader := &streamReaderWithCancel{
		ReadCloser: resp.Body,
		cancel:     cancel,
	}

	return reader, nil
}

// List 列出腾讯云 COS 中指定前缀的对象
func (p *TencentCOSProvider) List(prefix string) ([]storage.FileInfo, error) {
	if p.closed {
		return nil, utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "腾讯云 COS 提供商已关闭")
	}

	p.logger.Debug("列出对象: prefix=%s", prefix)

	// 构建列表参数
	opt := &cos.BucketGetOptions{
		Prefix: prefix,
	}

	// 默认每页1000个对象
	opt.MaxKeys = 1000

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	resp, err := p.client.Object.List(ctx, opt)
	if err != nil {
		return nil, utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
			"列出腾讯云 COS 对象失败")
	}

	// 解析结果
	var objects []storage.FileInfo
	for _, item := range resp.Objects {
		fileInfo := storage.FileInfo{
			Name:         item.Key,
			Size:         item.Size,
			LastModified: item.LastModified,
			ETag:         strings.Trim(item.ETag, "\""),
			StorageClass: item.StorageClass,
			IsDir:        false,
		}
		objects = append(objects, fileInfo)
	}

	p.logger.Debug("找到 %d 个对象", len(objects))
	return objects, nil
}

// Delete 删除腾讯云 COS 中的对象
func (p *TencentCOSProvider) Delete(path string) error {
	if p.closed {
		return utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "腾讯云 COS 提供商已关闭")
	}

	p.logger.Debug("删除对象: %s", path)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	_, err := p.client.Object.Delete(ctx, path)
	if err != nil {
		return utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
			"删除腾讯云 COS 对象失败: %s", path)
	}

	p.logger.Debug("删除完成: %s", path)
	return nil
}

// DeleteMultiple 批量删除腾讯云 COS 中的对象
func (p *TencentCOSProvider) DeleteMultiple(paths []string) error {
	if p.closed {
		return utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "腾讯云 COS 提供商已关闭")
	}

	p.logger.Debug("批量删除: %d 个对象", len(paths))

	// 构建删除对象列表
	objects := make([]cos.ObjectDelete, len(paths))
	for i, path := range paths {
		objects[i] = cos.ObjectDelete{
			Key: path,
		}
	}

	input := &cos.MultiDeleteOptions{
		ObjectDeleteList: objects,
		Quiet:            true, // 安静模式，不返回详细结果
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	_, err := p.client.Object.MultiDelete(ctx, input)
	if err != nil {
		return utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
			"批量删除腾讯云 COS 对象失败")
	}

	return nil
}

// Stat 获取腾讯云 COS 中对象的元数据信息
func (p *TencentCOSProvider) Stat(path string) (storage.FileInfo, error) {
	if p.closed {
		return storage.FileInfo{}, utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "腾讯云 COS 提供商已关闭")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	resp, err := p.client.Object.Head(ctx, path, nil)
	if err != nil {
		// 检查是否是404错误
		var cosErr *cosErrorResponse
		if errors.As(err, &cosErr) && cosErr.StatusCode == 404 {
			return storage.FileInfo{}, utils.Wrap(storage.ErrNotFound, utils.ErrorTypeStorage,
				utils.ErrCodeStorageNotFound, "对象不存在: %s", path)
		}
		return storage.FileInfo{}, utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
			"获取对象元数据失败: %s", path)
	}

	// 解析响应头
	size := resp.ContentLength
	lastModified, _ := time.Parse(time.RFC1123, resp.Header.Get("Last-Modified"))
	etag := strings.Trim(resp.Header.Get("ETag"), "\"")
	storageClass := resp.Header.Get("x-cos-storage-class")

	fileInfo := storage.FileInfo{
		Name:         path,
		Size:         size,
		LastModified: lastModified,
		ETag:         etag,
		StorageClass: storageClass,
		IsDir:        false,
	}

	p.logger.Debug("获取对象信息: %s (大小: %d, 修改时间: %v)", path, size, lastModified)
	return fileInfo, nil
}

// Copy 复制腾讯云 COS 中的对象
func (p *TencentCOSProvider) Copy(srcPath, dstPath string) error {
	if p.closed {
		return utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "腾讯云 COS 提供商已关闭")
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

	// 构建源URL
	sourceURL := fmt.Sprintf("%s/%s", p.client.BaseURL.BucketURL.String(), srcPath)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	_, _, err = p.client.Object.Copy(ctx, dstPath, sourceURL, nil)
	if err != nil {
		return utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
			"复制腾讯云 COS 对象失败: %s -> %s", srcPath, dstPath)
	}

	p.logger.Info("复制完成: %s -> %s", srcPath, dstPath)
	return nil
}

// Move 移动腾讯云 COS 中的对象
func (p *TencentCOSProvider) Move(srcPath, dstPath string) error {
	if p.closed {
		return utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "腾讯云 COS 提供商已关闭")
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

// Exists 检查腾讯云 COS 中的对象是否存在
func (p *TencentCOSProvider) Exists(path string) (bool, error) {
	if p.closed {
		return false, utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "腾讯云 COS 提供商已关闭")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := p.client.Object.Head(ctx, path, nil)
	if err != nil {
		// 检查是否是404错误
		var cosErr *cosErrorResponse
		if errors.As(err, &cosErr) && cosErr.StatusCode == 404 {
			return false, nil
		}
		return false, utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
			"检查对象是否存在失败: %s", path)
	}

	return true, nil
}

// GetURL 获取腾讯云 COS 中对象的访问URL
func (p *TencentCOSProvider) GetURL(path string, expires time.Duration) (string, error) {
	if p.closed {
		return "", utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "腾讯云 COS 提供商已关闭")
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

	ctx := context.Background()
	authTime := &cos.AuthTime{
		SignStartTime: time.Now(),
		SignEndTime:   time.Now().Add(expires),
	}

	// 获取预签名URL
	presignedURL, err := p.client.Object.GetPresignedURL(ctx, http.MethodGet, path,
		p.config.Credentials.SecretID, p.config.Credentials.SecretKey, expires, authTime)
	if err != nil {
		return "", utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
			"生成预签名URL失败: %s", path)
	}

	p.logger.Debug("生成预签名URL: %s (有效期: %v)", path, expires)
	return presignedURL.String(), nil
}

// Close 关闭腾讯云 COS 提供商连接
func (p *TencentCOSProvider) Close() error {
	if p.closed {
		return nil
	}

	p.logger.Info("关闭腾讯云 COS 提供商连接")
	p.closed = true

	// 腾讯云 COS 客户端没有显式的关闭方法
	// 这里主要是标记为已关闭状态
	return nil
}

// ProviderName 返回提供商名称
func (p *TencentCOSProvider) ProviderName() string {
	return "腾讯云 COS"
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

// cosErrorResponse 腾讯云 COS 错误响应
type cosErrorResponse struct {
	StatusCode int
	Code       string
	Message    string
	RequestID  string
	HostID     string
}
