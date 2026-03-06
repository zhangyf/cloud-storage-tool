// Package providers 提供 AWS S3 存储提供商实现
package providers

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/zhangyf/cloud-storage-tool/internal/storage"
	"github.com/zhangyf/cloud-storage-tool/internal/utils"
)

// AWSS3Provider AWS S3 存储提供商实现
type AWSS3Provider struct {
	config   storage.Config
	s3Client *s3.Client
	logger   *utils.Logger
	closed   bool
}

// NewAWSS3Provider 创建 AWS S3 存储提供商实例
func NewAWSS3Provider(config storage.Config) (storage.StorageProvider, error) {
	// 验证配置
	if config.Bucket == "" {
		return nil, utils.NewErrorBuilder(utils.ErrorTypeConfig).
			WithCode(utils.ErrCodeConfigInvalid).
			WithMessage("AWS S3 存储桶名称不能为空").
			Build()
	}

	if config.Region == "" {
		return nil, utils.NewErrorBuilder(utils.ErrorTypeConfig).
			WithCode(utils.ErrCodeConfigInvalid).
			WithMessage("AWS S3 区域不能为空").
			Build()
	}

	if config.Credentials.AWSAccessKeyID == "" || config.Credentials.AWSSecretAccessKey == "" {
		return nil, utils.NewErrorBuilder(utils.ErrorTypeConfig).
			WithCode(utils.ErrCodeConfigInvalid).
			WithMessage("AWS S3 AccessKeyID 和 SecretAccessKey 不能为空").
			Build()
	}

	// 加载 AWS 配置
	awsConfig, err := config.LoadDefaultConfig()
	if err != nil {
		return nil, utils.Wrapf(err, utils.ErrorTypeConfig, utils.ErrCodeConfigParseFailed,
			"加载 AWS 配置失败: %v", err)
	}

	// 创建 S3 客户端
	s3Client := s3.NewFromConfig(awsConfig)

	// 创建日志记录器
	logger := utils.GetLogger().WithField("provider", "aws_s3")

	logger.Info("AWS S3 提供商初始化成功，存储桶: %s, 区域: %s", config.Bucket, config.Region)

	return &AWSS3Provider{
		config:   config,
		s3Client: s3Client,
		logger:   logger,
		closed:   false,
	}, nil
}

// Upload 上传本地文件到 AWS S3
func (p *AWSS3Provider) Upload(localPath, remotePath string) error {
	if p.closed {
		return utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "AWS S3 提供商已关闭")
	}

	p.logger.Info("上传文件: %s -> %s", localPath, remotePath)

	// 打开本地文件
	file, err := os.Open(localPath)
	if err != nil {
		return utils.Wrapf(err, utils.ErrorTypeIO, utils.ErrCodeFileNotFound,
			"打开本地文件失败: %s", localPath)
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		return utils.Wrapf(err, utils.ErrorTypeIO, utils.ErrCodeFileReadFailed,
			"获取文件信息失败: %s", localPath)
	}

	// 构建上传输入
	input := &s3.PutObjectInput{
		Bucket:        aws.String(p.config.Bucket),
		Key:           aws.String(remotePath),
		Body:          file,
		ContentLength: aws.Int64(fileInfo.Size()),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	_, err = p.s3Client.PutObject(ctx, input)
	if err != nil {
		return utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
			"上传到 AWS S3 失败: %s", remotePath)
	}

	p.logger.Info("上传完成: %s -> %s", localPath, remotePath)
	return nil
}

// UploadStream 从流上传数据到 AWS S3
func (p *AWSS3Provider) UploadStream(reader io.Reader, remotePath string, size int64) error {
	if p.closed {
		return utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "AWS S3 提供商已关闭")
	}

	p.logger.Info("流式上传: %s", remotePath)

	input := &s3.PutObjectInput{
		Bucket:        aws.String(p.config.Bucket),
		Key:           aws.String(remotePath),
		Body:          reader,
		ContentLength: aws.Int64(size),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	_, err := p.s3Client.PutObject(ctx, input)
	if err != nil {
		return utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
			"流式上传失败: %s", remotePath)
	}

	p.logger.Debug("流式上传完成: %s", remotePath)
	return nil
}

// Download 从 AWS S3 下载文件到本地
func (p *AWSS3Provider) Download(remotePath, localPath string) error {
	if p.closed {
		return utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "AWS S3 提供商已关闭")
	}

	p.logger.Info("下载文件: %s -> %s", remotePath, localPath)

	// 创建本地文件
	file, err := os.Create(localPath)
	if err != nil {
		return utils.Wrapf(err, utils.ErrorTypeIO, utils.ErrCodeFileWriteFailed,
			"创建本地文件失败: %s", localPath)
	}
	defer file.Close()

	// 构建下载输入
	input := &s3.GetObjectInput{
		Bucket: aws.String(p.config.Bucket),
		Key:    aws.String(remotePath),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	result, err := p.s3Client.GetObject(ctx, input)
	if err != nil {
		return utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageNotFound,
			"下载 AWS S3 对象失败: %s", remotePath)
	}
	defer result.Body.Close()

	// 复制内容到文件
	_, err = io.Copy(file, result.Body)
	if err != nil {
		return utils.Wrapf(err, utils.ErrorTypeIO, utils.ErrCodeFileWriteFailed,
			"写入本地文件失败: %s", localPath)
	}

	p.logger.Info("下载完成: %s -> %s", remotePath, localPath)
	return nil
}

// DownloadStream 从 AWS S3 下载数据到流
func (p *AWSS3Provider) DownloadStream(remotePath string) (io.ReadCloser, error) {
	if p.closed {
		return nil, utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "AWS S3 提供商已关闭")
	}

	p.logger.Debug("流式下载: %s", remotePath)

	input := &s3.GetObjectInput{
		Bucket: aws.String(p.config.Bucket),
		Key:    aws.String(remotePath),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	result, err := p.s3Client.GetObject(ctx, input)
	if err != nil {
		return nil, utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageNotFound,
			"获取 AWS S3 对象失败: %s", remotePath)
	}

	// 包装读取器以支持取消
	reader := &streamReaderWithCancel{
		ReadCloser: result.Body,
		cancel:     cancel,
	}

	return reader, nil
}

// List 列出 AWS S3 中指定前缀的对象
func (p *AWSS3Provider) List(prefix string) ([]storage.FileInfo, error) {
	if p.closed {
		return nil, utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "AWS S3 提供商已关闭")
	}

	p.logger.Debug("列出对象: prefix=%s", prefix)

	// 构建列表输入
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(p.config.Bucket),
		Prefix: aws.String(prefix),
	}

	var objects []storage.FileInfo
	cursor := ""

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	for {
		if cursor != "" {
			input.StartAfter = aws.String(cursor)
		}

		result, err := p.s3Client.ListObjectsV2(ctx, input)
		if err != nil {
			return nil, utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
				"列出 AWS S3 对象失败")
		}

		// 解析结果
		for _, obj := range result.Contents {
			fileInfo := storage.FileInfo{
				Name:         *obj.Key,
				Size:         *obj.Size,
				LastModified: *obj.LastModified,
				ETag:         strings.Trim(*obj.ETag, "\""),
				StorageClass: string(obj.StorageClass),
				IsDir:        false,
			}
			objects = append(objects, fileInfo)
		}

		// 检查是否还有更多对象
		if result.IsTruncated == nil || !*result.IsTruncated {
			break
		}

		if result.NextContinuationToken != nil {
			cursor = *result.NextContinuationToken
		} else if result.Contents != nil && len(result.Contents) > 0 {
			cursor = *result.Contents[len(result.Contents)-1].Key
		}
	}

	p.logger.Debug("找到 %d 个对象", len(objects))
	return objects, nil
}

// Delete 删除 AWS S3 中的对象
func (p *AWSS3Provider) Delete(path string) error {
	if p.closed {
		return utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "AWS S3 提供商已关闭")
	}

	p.logger.Debug("删除对象: %s", path)

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(p.config.Bucket),
		Key:    aws.String(path),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	_, err := p.s3Client.DeleteObject(ctx, input)
	if err != nil {
		return utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
			"删除 AWS S3 对象失败: %s", path)
	}

	p.logger.Debug("删除完成: %s", path)
	return nil
}

// DeleteMultiple 批量删除 AWS S3 中的对象
func (p *AWSS3Provider) DeleteMultiple(paths []string) error {
	if p.closed {
		return utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "AWS S3 提供商已关闭")
	}

	p.logger.Debug("批量删除: %d 个对象", len(paths))

	// AWS SDK 限制每次最多删除 1000 个对象
	const maxBatchSize = 1000
	for i := 0; i < len(paths); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(paths) {
			end = len(paths)
		}

		if err := p.deleteBatch(paths[i:end]); err != nil {
			return err
		}
	}

	return nil
}

// deleteBatch 删除一批对象
func (p *AWSS3Provider) deleteBatch(paths []string) error {
	// 构建删除对象列表
	objects := make([]*s3.ObjectIdentifier, len(paths))
	for i, path := range paths {
		objects[i] = &s3.ObjectIdentifier{
			Key: aws.String(path),
		}
	}

	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(p.config.Bucket),
		Delete: &s3.Delete{
			Objects: objects,
			Quiet:   aws.Bool(true), // 安静模式，不返回详细结果
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	_, err := p.s3Client.DeleteObjects(ctx, input)
	if err != nil {
		return utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
			"批量删除 AWS S3 对象失败")
	}

	return nil
}

// Stat 获取 AWS S3 中对象的元数据信息
func (p *AWSS3Provider) Stat(path string) (storage.FileInfo, error) {
	if p.closed {
		return storage.FileInfo{}, utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "AWS S3 提供商已关闭")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	result, err := p.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(p.config.Bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		// 检查是否是404错误
		var apiErr *types.ResourceNotFoundException
		if errors.As(err, &apiErr) {
			return storage.FileInfo{}, utils.Wrap(storage.ErrNotFound, utils.ErrorTypeStorage,
				utils.ErrCodeStorageNotFound, "对象不存在: %s", path)
		}
		return storage.FileInfo{}, utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
			"获取对象元数据失败: %s", path)
	}

	fileInfo := storage.FileInfo{
		Name:         path,
		Size:         *result.ContentLength,
		LastModified: *result.LastModified,
		ETag:         strings.Trim(*result.ETag, "\""),
		StorageClass: string(result.StorageClass),
		IsDir:        false,
	}

	p.logger.Debug("获取对象信息: %s (大小: %d, 修改时间: %v)", path, *result.ContentLength, *result.LastModified)
	return fileInfo, nil
}

// Copy 复制 AWS S3 中的对象
func (p *AWSS3Provider) Copy(srcPath, dstPath string) error {
	if p.closed {
		return utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "AWS S3 提供商已关闭")
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

	// 构建源对象引用
	source := fmt.Sprintf("%s/%s", p.config.Bucket, srcPath)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(p.config.Timeout)*time.Second)
	defer cancel()

	_, err = p.s3Client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(p.config.Bucket),
		CopySource: aws.String(source),
		Key:        aws.String(dstPath),
	})
	if err != nil {
		return utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
			"复制 AWS S3 对象失败: %s -> %s", srcPath, dstPath)
	}

	p.logger.Info("复制完成: %s -> %s", srcPath, dstPath)
	return nil
}

// Move 移动 AWS S3 中的对象
func (p *AWSS3Provider) Move(srcPath, dstPath string) error {
	if p.closed {
		return utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "AWS S3 提供商已关闭")
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

// Exists 检查 AWS S3 中的对象是否存在
func (p *AWSS3Provider) Exists(path string) (bool, error) {
	if p.closed {
		return false, utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "AWS S3 提供商已关闭")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := p.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(p.config.Bucket),
		Key:    aws.String(path),
	})
	if err != nil {
		var apiErr *types.ResourceNotFoundException
		if errors.As(err, &apiErr) {
			return false, nil
		}
		return false, utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
			"检查对象是否存在失败: %s", path)
	}

	return true, nil
}

// GetURL 获取 AWS S3 中对象的访问URL
func (p *AWSS3Provider) GetURL(path string, expires time.Duration) (string, error) {
	if p.closed {
		return "", utils.Wrap(storage.ErrProviderClosed, utils.ErrorTypeStorage,
			utils.ErrCodeStorageAccessDenied, "AWS S3 提供商已关闭")
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

	req, _ := p.s3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(p.config.Bucket),
		Key:    aws.String(path),
	})

	urlStr, err := req.Presign(expires)
	if err != nil {
		return "", utils.Wrapf(err, utils.ErrorTypeStorage, utils.ErrCodeStorageAccessDenied,
			"生成预签名URL失败: %s", path)
	}

	p.logger.Debug("生成预签名URL: %s (有效期: %v)", path, expires)
	return urlStr, nil
}

// Close 关闭 AWS S3 提供商连接
func (p *AWSS3Provider) Close() error {
	if p.closed {
		return nil
	}

	p.logger.Info("关闭 AWS S3 提供商连接")
	p.closed = true

	// AWS SDK 没有显式的关闭方法
	// 这里主要是标记为已关闭状态
	return nil
}

// ProviderName 返回提供商名称
func (p *AWSS3Provider) ProviderName() string {
	return "AWS S3"
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
