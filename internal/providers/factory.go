// Package providers 提供云存储提供商的工厂模式实现
//
// 工厂模式负责：
// 1. 提供商类型映射：将配置中的type映射到具体提供商
// 2. 配置验证：验证提供商配置的完整性
// 3. 实例创建：根据配置创建对应的提供商实例
// 4. 错误处理：处理各种错误情况
package providers


import (
	"github.com/zhangyf/cloud-storage-tool/internal/storage"
)
// ProviderType 提供商类型常量
const (
	// TypeTencentCOS 腾讯云COS提供商类型
	TypeTencentCOS = "tencent_cos"
	
	// TypeAWSS3 AWS S3提供商类型
	TypeAWSS3 = "aws_s3"
)

// FactoryErrorCode 工厂错误代码类型
type FactoryErrorCode string

// 工厂错误代码常量
const (
	// ErrCodeUnsupportedProvider 不支持的提供商类型
	ErrCodeUnsupportedProvider FactoryErrorCode = "UNSUPPORTED_PROVIDER"
	
	// ErrCodeConfigInvalid 配置无效
	ErrCodeConfigInvalid FactoryErrorCode = "CONFIG_INVALID"
	
	// ErrCodeConfigMissingField 配置缺少必需字段
	ErrCodeConfigMissingField FactoryErrorCode = "CONFIG_MISSING_FIELD"
	
	// ErrCodeProviderCreationFailed 提供商创建失败
	ErrCodeProviderCreationFailed FactoryErrorCode = "PROVIDER_CREATION_FAILED"
	
	// ErrCodeProviderInitializationFailed 提供商初始化失败
	ErrCodeProviderInitializationFailed FactoryErrorCode = "PROVIDER_INITIALIZATION_FAILED"
)

// FactoryError 工厂错误类型
type FactoryError struct {
	// Code 错误代码
	Code FactoryErrorCode
	
	// Message 错误消息
	Message string
	
	// ProviderType 提供商类型（如果相关）
	ProviderType string
	
	// ConfigField 配置字段（如果相关）
	ConfigField string
}

// Error 实现error接口
func (e *FactoryError) Error() string {
	if e.ProviderType != "" {
		return string(e.Code) + ": " + e.Message + " (provider: " + e.ProviderType + ")"
	}
	return string(e.Code) + ": " + e.Message
}

// ProviderFactory 提供商工厂接口
type ProviderFactory interface {
	// CreateProvider 创建存储提供商实例
	CreateProvider(config interface{}) (interface{}, error)
	
	// ValidateConfig 验证提供商配置
	ValidateConfig(config interface{}) error
	
	// SupportedProviderTypes 返回支持的提供商类型列表
	SupportedProviderTypes() []string
}

// NewFactoryError 创建新的工厂错误
func NewFactoryError(code FactoryErrorCode, message string) *FactoryError {
	return &FactoryError{
		Code:    code,
		Message: message,
	}
}

// DefaultProviderFactory 默认提供商工厂实现
type DefaultProviderFactory struct{}

// NewDefaultProviderFactory 创建默认提供商工厂实例
func NewDefaultProviderFactory() *DefaultProviderFactory {
	return &DefaultProviderFactory{}
}

// CreateProvider 创建存储提供商实例
func (f *DefaultProviderFactory) CreateProvider(config interface{}) (interface{}, error) {
	// 将配置转换为map[string]interface{}以便处理
	configMap, ok := config.(map[string]interface{})
	if !ok {
		return nil, NewFactoryError(ErrCodeConfigInvalid, "配置必须是map[string]interface{}类型")
	}
	
	// 获取提供商类型
	providerType, ok := configMap["type"].(string)
	if !ok {
		return nil, NewFactoryError(ErrCodeConfigMissingField, "缺少提供商类型(type)字段")
	}
	
	// 根据提供商类型创建实例
	switch providerType {
	case TypeTencentCOS:
		return f.createTencentCOS(configMap)
	case TypeAWSS3:
		return f.createAWSS3(configMap)
	default:
		return nil, &FactoryError{
			Code:         ErrCodeUnsupportedProvider,
			Message:      "不支持的提供商类型",
			ProviderType: providerType,
		}
	}
}

// ValidateConfig 验证提供商配置
func (f *DefaultProviderFactory) ValidateConfig(config interface{}) error {
	configMap, ok := config.(map[string]interface{})
	if !ok {
		return NewFactoryError(ErrCodeConfigInvalid, "配置必须是map[string]interface{}类型")
	}
	
	// 检查提供商类型
	providerType, ok := configMap["type"].(string)
	if !ok {
		return &FactoryError{
			Code:        ErrCodeConfigMissingField,
			Message:     "缺少提供商类型(type)字段",
			ConfigField: "type",
		}
	}
	
	// 验证提供商类型是否支持
	supportedTypes := f.SupportedProviderTypes()
	found := false
	for _, t := range supportedTypes {
		if t == providerType {
			found = true
			break
		}
	}
	
	if !found {
		return &FactoryError{
			Code:         ErrCodeUnsupportedProvider,
			Message:      "不支持的提供商类型",
			ProviderType: providerType,
		}
	}
	
	// 根据提供商类型验证具体配置
	switch providerType {
	case TypeTencentCOS:
		return f.validateTencentCOSConfig(configMap)
	case TypeAWSS3:
		return f.validateAWSS3Config(configMap)
	}
	
	return nil
}

// SupportedProviderTypes 返回支持的提供商类型列表
func (f *DefaultProviderFactory) SupportedProviderTypes() []string {
	return []string{
		TypeTencentCOS,
		TypeAWSS3,
	}
}

// createTencentCOS 创建腾讯云COS提供商实例
func (f *DefaultProviderFactory) createTencentCOS(config map[string]interface{}) (interface{}, error) {
	// 创建简单的 storage.Config
	storageConfig := storage.Config{}
	
	// 设置基本字段
	if typeVal, ok := config["type"].(string); ok {
		storageConfig.Type = storage.ProviderType(typeVal)
	}
	
	if bucket, ok := config["bucket"].(string); ok {
		storageConfig.Bucket = bucket
	}
	
	if region, ok := config["region"].(string); ok {
		storageConfig.Region = region
	}
	
	// 设置认证信息
	if secretID, ok := config["secret_id"].(string); ok {
		storageConfig.Credentials.SecretID = secretID
	}
	
	if secretKey, ok := config["secret_key"].(string); ok {
		storageConfig.Credentials.SecretKey = secretKey
	}
	
	// 调用最小化版的构造函数
	return NewTencentCOSProvider(storageConfig)
}

// createAWSS3 创建AWS S3提供商实例
func (f *DefaultProviderFactory) createAWSS3(config map[string]interface{}) (interface{}, error) {
	// 创建简单的 storage.Config
	storageConfig := storage.Config{}
	
	// 设置基本字段
	if typeVal, ok := config["type"].(string); ok {
		storageConfig.Type = storage.ProviderType(typeVal)
	}
	
	if bucket, ok := config["bucket"].(string); ok {
		storageConfig.Bucket = bucket
	}
	
	if region, ok := config["region"].(string); ok {
		storageConfig.Region = region
	}
	
	// 设置认证信息
	if accessKeyID, ok := config["access_key_id"].(string); ok {
		storageConfig.Credentials.AWSAccessKeyID = accessKeyID
	}
	
	if secretAccessKey, ok := config["secret_access_key"].(string); ok {
		storageConfig.Credentials.AWSSecretAccessKey = secretAccessKey
	}
	
	// 调用简化版的构造函数
	return NewAWSS3Provider(storageConfig)
}

// validateTencentCOSConfig 验证腾讯云COS配置
func (f *DefaultProviderFactory) validateTencentCOSConfig(config map[string]interface{}) error {
	requiredFields := []string{"secret_id", "secret_key", "region", "bucket"}
	for _, field := range requiredFields {
		if _, ok := config[field]; !ok {
			return &FactoryError{
				Code:         ErrCodeConfigMissingField,
				Message:      "缺少必需字段",
				ProviderType: TypeTencentCOS,
				ConfigField:  field,
			}
		}
	}
	return nil
}

// validateAWSS3Config 验证AWS S3配置
func (f *DefaultProviderFactory) validateAWSS3Config(config map[string]interface{}) error {
	requiredFields := []string{"access_key_id", "secret_access_key", "region", "bucket"}
	for _, field := range requiredFields {
		if _, ok := config[field]; !ok {
			return &FactoryError{
				Code:         ErrCodeConfigMissingField,
				Message:      "缺少必需字段",
				ProviderType: TypeAWSS3,
				ConfigField:  field,
			}
		}
	}
	return nil
}