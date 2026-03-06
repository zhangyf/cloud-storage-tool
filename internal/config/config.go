// Package config 提供配置文件管理功能
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config 表示整个应用程序的配置
type Config struct {
	// DefaultProvider 默认使用的云存储提供商
	DefaultProvider string `yaml:"default_provider"`
	
	// Providers 云存储提供商配置
	Providers map[string]ProviderConfig `yaml:"providers"`
}

// ProviderConfig 表示单个云存储提供商的配置
type ProviderConfig struct {
	// Type 提供商类型，如 "tencent_cos", "aliyun_oss", "aws_s3"
	Type string `yaml:"type"`
	
	// Bucket 存储桶名称
	Bucket string `yaml:"bucket"`
	
	// Region 区域，对于腾讯云COS和AWS S3
	Region string `yaml:"region,omitempty"`
	
	// Endpoint 端点，对于阿里云OSS
	Endpoint string `yaml:"endpoint,omitempty"`
	
	// SecretID 或 AccessKeyID
	SecretID string `yaml:"secret_id,omitempty"`
	AccessKeyID string `yaml:"access_key_id,omitempty"`
	
	// SecretKey 或 AccessKeySecret
	SecretKey string `yaml:"secret_key,omitempty"`
	AccessKeySecret string `yaml:"access_key_secret,omitempty"`
	
	// SecretAccessKey 对于AWS S3
	SecretAccessKey string `yaml:"secret_access_key,omitempty"`
	
	// SessionToken 临时凭证的会话令牌
	SessionToken string `yaml:"session_token,omitempty"`
	
	// Timeout 操作超时时间（秒）
	Timeout int `yaml:"timeout,omitempty"`
	
	// MaxRetries 最大重试次数
	MaxRetries int `yaml:"max_retries,omitempty"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		DefaultProvider: "tencent_cos",
		Providers: map[string]ProviderConfig{
			"tencent_cos": {
				Type:     "tencent_cos",
				Bucket:   "",
				Region:   "ap-beijing",
				SecretID: "",
				SecretKey: "",
				Timeout:   30,
				MaxRetries: 3,
			},
			"aliyun_oss": {
				Type:            "aliyun_oss",
				Bucket:          "",
				Endpoint:        "oss-cn-beijing.aliyuncs.com",
				AccessKeyID:     "",
				AccessKeySecret: "",
				Timeout:         30,
				MaxRetries:      3,
			},
			"aws_s3": {
				Type:              "aws_s3",
				Bucket:            "",
				Region:            "us-east-1",
				AccessKeyID:       "",
				SecretAccessKey:   "",
				SessionToken:      "",
				Timeout:           30,
				MaxRetries:        3,
			},
		},
	}
}

// Load 从文件加载配置
func Load(configPath string) (*Config, error) {
	if configPath == "" {
		// 使用默认配置文件路径
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("获取用户主目录失败: %w", err)
		}
		configPath = filepath.Join(homeDir, ".cloud-storage", "config.yaml")
	}

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", configPath)
	}

	// 读取文件内容
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return &config, nil
}

// Save 保存配置到文件
func (c *Config) Save(configPath string) error {
	if configPath == "" {
		// 使用默认配置文件路径
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("获取用户主目录失败: %w", err)
		}
		configDir := filepath.Join(homeDir, ".cloud-storage")
		configPath = filepath.Join(configDir, "config.yaml")
		
		// 创建配置目录
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("创建配置目录失败: %w", err)
		}
	}

	// 验证配置
	if err := c.Validate(); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}

	// 转换为YAML
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	// 写入文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	if c.DefaultProvider == "" {
		return errors.New("默认提供商不能为空")
	}

	// 检查默认提供商是否存在
	if _, exists := c.Providers[c.DefaultProvider]; !exists {
		return fmt.Errorf("默认提供商 '%s' 未在配置中定义", c.DefaultProvider)
	}

	// 验证每个提供商的配置
	for name, provider := range c.Providers {
		if err := provider.Validate(); err != nil {
			return fmt.Errorf("提供商 '%s' 配置无效: %w", name, err)
		}
	}

	return nil
}

// Validate 验证提供商配置
func (p *ProviderConfig) Validate() error {
	if p.Type == "" {
		return errors.New("提供商类型不能为空")
	}

	if p.Bucket == "" {
		return errors.New("存储桶名称不能为空")
	}

	// 根据提供商类型验证必要字段
	switch p.Type {
	case "tencent_cos":
		if p.Region == "" {
			return errors.New("腾讯云COS区域不能为空")
		}
		if p.SecretID == "" {
			return errors.New("腾讯云COS SecretID不能为空")
		}
		if p.SecretKey == "" {
			return errors.New("腾讯云COS SecretKey不能为空")
		}
		
	case "aliyun_oss":
		if p.Endpoint == "" {
			return errors.New("阿里云OSS端点不能为空")
		}
		if p.AccessKeyID == "" {
			return errors.New("阿里云OSS AccessKeyID不能为空")
		}
		if p.AccessKeySecret == "" {
			return errors.New("阿里云OSS AccessKeySecret不能为空")
		}
		
	case "aws_s3":
		if p.Region == "" {
			return errors.New("AWS S3区域不能为空")
		}
		if p.AccessKeyID == "" {
			return errors.New("AWS S3 AccessKeyID不能为空")
		}
		if p.SecretAccessKey == "" {
			return errors.New("AWS S3 SecretAccessKey不能为空")
		}
		
	default:
		return fmt.Errorf("不支持的提供商类型: %s", p.Type)
	}

	// 验证超时设置
	if p.Timeout <= 0 {
		p.Timeout = 30
	}

	// 验证重试次数
	if p.MaxRetries < 0 {
		p.MaxRetries = 3
	}

	return nil
}

// GetProvider 获取指定提供商的配置
func (c *Config) GetProvider(name string) (*ProviderConfig, error) {
	provider, exists := c.Providers[name]
	if !exists {
		return nil, fmt.Errorf("提供商 '%s' 未在配置中定义", name)
	}
	return &provider, nil
}

// GetDefaultProvider 获取默认提供商的配置
func (c *Config) GetDefaultProvider() (*ProviderConfig, error) {
	return c.GetProvider(c.DefaultProvider)
}