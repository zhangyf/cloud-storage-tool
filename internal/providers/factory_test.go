package providers

import (
	"testing"
	
	"github.com/zhangyf/cloud-storage-tool/internal/storage"
)

// TestSupportedProviderTypes 测试支持的提供商类型列表
func TestSupportedProviderTypes(t *testing.T) {
	factory := NewDefaultProviderFactory()
	types := factory.SupportedProviderTypes()
	
	// 验证返回的类型数量
	if len(types) != 2 {
		t.Errorf("期望支持2个提供商类型，实际得到 %d 个", len(types))
	}
	
	// 验证包含所有预期的类型
	expectedTypes := map[string]bool{
		TypeTencentCOS: false,
		TypeAWSS3:      false,
	}
	
	for _, tpe := range types {
		if _, exists := expectedTypes[tpe]; exists {
			expectedTypes[tpe] = true
		} else {
			t.Errorf("发现未预期的提供商类型：%s", tpe)
		}
	}
	
	// 验证所有预期类型都存在
	for tpe, found := range expectedTypes {
		if !found {
			t.Errorf("未找到预期的提供商类型：%s", tpe)
		}
	}
}

// TestValidateConfigValid 测试有效的配置验证
func TestValidateConfigValid(t *testing.T) {
	factory := NewDefaultProviderFactory()
	
	// 测试腾讯云COS配置
	cosConfig := map[string]interface{}{
		"type":       TypeTencentCOS,
		"secret_id":  "test_secret_id",
		"secret_key": "test_secret_key",
		"region":     "ap-beijing",
		"bucket":     "test-bucket",
	}
	
	if err := factory.ValidateConfig(cosConfig); err != nil {
		t.Errorf("腾讯云COS配置验证失败：%v", err)
	}
	
	// 测试AWS S3配置
	s3Config := map[string]interface{}{
		"type":                TypeAWSS3,
		"access_key_id":       "test_access_key_id",
		"secret_access_key":   "test_secret_access_key",
		"region":              "us-east-1",
		"bucket":              "test-bucket",
	}
	
	if err := factory.ValidateConfig(s3Config); err != nil {
		t.Errorf("AWS S3配置验证失败：%v", err)
	}
}

// TestValidateConfigInvalidType 测试无效的配置类型
func TestValidateConfigInvalidType(t *testing.T) {
	factory := NewDefaultProviderFactory()
	
	// 测试非map类型的配置
	invalidConfig := "not a map"
	
	if err := factory.ValidateConfig(invalidConfig); err == nil {
		t.Error("期望验证非map类型配置失败，但验证通过了")
	} else {
		expectedCode := ErrCodeConfigInvalid
		if fe, ok := err.(*FactoryError); ok {
			if fe.Code != expectedCode {
				t.Errorf("期望错误代码 %s，实际得到 %s", expectedCode, fe.Code)
			}
		} else {
			t.Errorf("期望 FactoryError 类型错误，实际得到 %T", err)
		}
	}
}

// TestValidateConfigMissingType 测试缺少type字段的配置
func TestValidateConfigMissingType(t *testing.T) {
	factory := NewDefaultProviderFactory()
	
	config := map[string]interface{}{
		"secret_id":  "test_secret_id",
		"secret_key": "test_secret_key",
		// 缺少type字段
	}
	
	if err := factory.ValidateConfig(config); err == nil {
		t.Error("期望验证缺少type字段的配置失败，但验证通过了")
	} else {
		expectedCode := ErrCodeConfigMissingField
		if fe, ok := err.(*FactoryError); ok {
			if fe.Code != expectedCode {
				t.Errorf("期望错误代码 %s，实际得到 %s", expectedCode, fe.Code)
			}
			if fe.ConfigField != "type" {
				t.Errorf("期望配置字段为 'type'，实际得到 %s", fe.ConfigField)
			}
		} else {
			t.Errorf("期望错误类型为 *FactoryError，实际得到 %T", err)
		}
	}
}

// TestValidateConfigUnsupportedProvider 测试不支持的提供商类型
func TestValidateConfigUnsupportedProvider(t *testing.T) {
	factory := NewDefaultProviderFactory()
	
	config := map[string]interface{}{
		"type":       "unsupported_provider",
		"secret_id":  "test_secret_id",
		"secret_key": "test_secret_key",
	}
	
	if err := factory.ValidateConfig(config); err == nil {
		t.Error("期望验证不支持的提供商类型失败，但验证通过了")
	} else {
		expectedCode := ErrCodeUnsupportedProvider
		if fe, ok := err.(*FactoryError); ok {
			if fe.Code != expectedCode {
				t.Errorf("期望错误代码 %s，实际得到 %s", expectedCode, fe.Code)
			}
			if fe.ProviderType != "unsupported_provider" {
				t.Errorf("期望提供商类型为 'unsupported_provider'，实际得到 %s", fe.ProviderType)
			}
		}
	}
}

// TestValidateConfigMissingRequiredFields 测试缺少必需字段的配置
func TestValidateConfigMissingRequiredFields(t *testing.T) {
	factory := NewDefaultProviderFactory()
	
	testCases := []struct {
		name           string
		providerType   string
		config         map[string]interface{}
		expectedField  string
	}{
		{
			name:         "腾讯云COS缺少secret_id",
			providerType: TypeTencentCOS,
			config: map[string]interface{}{
				"type":       TypeTencentCOS,
				// 缺少secret_id
				"secret_key": "test_secret_key",
				"region":     "ap-beijing",
				"bucket":     "test-bucket",
			},
			expectedField: "secret_id",
		},
		{
			name:         "AWS S3缺少region",
			providerType: TypeAWSS3,
			config: map[string]interface{}{
				"type":              TypeAWSS3,
				"access_key_id":     "test_access_key_id",
				"secret_access_key": "test_secret_access_key",
				// 缺少region
				"bucket":            "test-bucket",
			},
			expectedField: "region",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if err := factory.ValidateConfig(tc.config); err == nil {
				t.Errorf("测试用例 '%s'：期望验证失败，但验证通过了", tc.name)
			} else {
				expectedCode := ErrCodeConfigMissingField
				if fe, ok := err.(*FactoryError); ok {
					if fe.Code != expectedCode {
						t.Errorf("测试用例 '%s'：期望错误代码 %s，实际得到 %s", tc.name, expectedCode, fe.Code)
					}
					if fe.ProviderType != tc.providerType {
						t.Errorf("测试用例 '%s'：期望提供商类型 %s，实际得到 %s", tc.name, tc.providerType, fe.ProviderType)
					}
					if fe.ConfigField != tc.expectedField {
						t.Errorf("测试用例 '%s'：期望配置字段 %s，实际得到 %s", tc.name, tc.expectedField, fe.ConfigField)
					}
				}
			}
		})
	}
}

// TestCreateProviderValidConfig 测试有效的配置创建提供商
func TestCreateProviderValidConfig(t *testing.T) {
	factory := NewDefaultProviderFactory()
	
	testCases := []struct {
		name         string
		providerType string
		config       map[string]interface{}
	}{
		{
			name:         "腾讯云COS",
			providerType: TypeTencentCOS,
			config: map[string]interface{}{
				"type":       TypeTencentCOS,
				"secret_id":  "test_secret_id",
				"secret_key": "test_secret_key",
				"region":     "ap-beijing",
				"bucket":     "test-bucket",
			},
		},
		{
			name:         "AWS S3",
			providerType: TypeAWSS3,
			config: map[string]interface{}{
				"type":                TypeAWSS3,
				"access_key_id":       "test_access_key_id",
				"secret_access_key":   "test_secret_access_key",
				"region":              "us-east-1",
				"bucket":              "test-bucket",
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// 现在工厂可以成功创建提供商（最小化版）
			provider, err := factory.CreateProvider(tc.config)
			
			// 验证没有错误
			if err != nil {
				t.Errorf("测试用例 '%s'：期望创建成功，但得到错误: %v", tc.name, err)
			}
			
			// 验证provider不为nil
			if provider == nil {
				t.Errorf("测试用例 '%s'：期望provider不为nil，但得到nil", tc.name)
			}
			
			// 验证provider实现了正确的接口
			if _, ok := provider.(storage.StorageProvider); !ok {
				t.Errorf("测试用例 '%s'：创建的provider没有实现StorageProvider接口", tc.name)
			}
		})
	}
}

// TestCreateProviderInvalidConfig 测试无效的配置创建提供商
func TestCreateProviderInvalidConfig(t *testing.T) {
	factory := NewDefaultProviderFactory()
	
	testCases := []struct {
		name         string
		config       interface{}
		expectedCode FactoryErrorCode
	}{
		{
			name:         "非map类型配置",
			config:       "invalid config",
			expectedCode: ErrCodeConfigInvalid,
		},
		{
			name: "缺少type字段",
			config: map[string]interface{}{
				"secret_id": "test_secret_id",
			},
			expectedCode: ErrCodeConfigMissingField,
		},
		{
			name: "不支持的提供商类型",
			config: map[string]interface{}{
				"type": "unknown_provider",
			},
			expectedCode: ErrCodeUnsupportedProvider,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider, err := factory.CreateProvider(tc.config)
			
			if err == nil {
				t.Errorf("测试用例 '%s'：期望返回错误，但错误为nil", tc.name)
			}
			
			if fe, ok := err.(*FactoryError); ok {
				if fe.Code != tc.expectedCode {
					t.Errorf("测试用例 '%s'：期望错误代码 %s，实际得到 %s", tc.name, tc.expectedCode, fe.Code)
				}
			}
			
			if provider != nil {
				t.Errorf("测试用例 '%s'：期望provider为nil，实际得到 %v", tc.name, provider)
			}
		})
	}
}

// TestFactoryError 测试工厂错误类型
func TestFactoryError(t *testing.T) {
	// 测试基本错误
	err := NewFactoryError(ErrCodeConfigInvalid, "配置无效")
	if err.Error() != string(ErrCodeConfigInvalid)+": 配置无效" {
		t.Errorf("期望错误消息 '%s: 配置无效'，实际得到 '%s'", ErrCodeConfigInvalid, err.Error())
	}
	
	// 测试带提供商类型的错误
	errWithProvider := &FactoryError{
		Code:         ErrCodeUnsupportedProvider,
		Message:      "不支持的提供商",
		ProviderType: "test_provider",
	}
	expectedMsg := string(ErrCodeUnsupportedProvider) + ": 不支持的提供商 (provider: test_provider)"
	if errWithProvider.Error() != expectedMsg {
		t.Errorf("期望错误消息 '%s'，实际得到 '%s'", expectedMsg, errWithProvider.Error())
	}
	
	// 测试带配置字段的错误
	errWithField := &FactoryError{
		Code:        ErrCodeConfigMissingField,
		Message:     "缺少字段",
		ConfigField: "secret_id",
	}
	if errWithField.ConfigField != "secret_id" {
		t.Errorf("期望配置字段 'secret_id'，实际得到 '%s'", errWithField.ConfigField)
	}
}