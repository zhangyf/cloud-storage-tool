// Package utils 提供错误处理工具
package utils

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// ErrorType 错误类型
type ErrorType string

const (
	// ErrorTypeConfig 配置错误
	ErrorTypeConfig ErrorType = "config"
	
	// ErrorTypeNetwork 网络错误
	ErrorTypeNetwork ErrorType = "network"
	
	// ErrorTypeStorage 存储错误
	ErrorTypeStorage ErrorType = "storage"
	
	// ErrorTypeIO 输入输出错误
	ErrorTypeIO ErrorType = "io"
	
	// ErrorTypeValidation 验证错误
	ErrorTypeValidation ErrorType = "validation"
	
	// ErrorTypeAuth 认证错误
	ErrorTypeAuth ErrorType = "auth"
	
	// ErrorTypeTimeout 超时错误
	ErrorTypeTimeout ErrorType = "timeout"
	
	// ErrorTypeUnknown 未知错误
	ErrorTypeUnknown ErrorType = "unknown"
)

// AppError 应用程序错误
type AppError struct {
	// Type 错误类型
	Type ErrorType
	
	// Code 错误代码
	Code string
	
	// Message 错误消息
	Message string
	
	// Operation 操作名称
	Operation string
	
	// Err 原始错误
	Err error
	
	// StackTrace 堆栈跟踪
	StackTrace string
	
	// Context 上下文信息
	Context map[string]interface{}
}

// NewAppError 创建新的应用程序错误
func NewAppError(errType ErrorType, code, message string) *AppError {
	return &AppError{
		Type:    errType,
		Code:    code,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// Wrap 包装错误
func Wrap(err error, errType ErrorType, code, message string) *AppError {
	appErr := NewAppError(errType, code, message)
	appErr.Err = err
	appErr.captureStackTrace()
	return appErr
}

// Wrapf 包装错误并格式化消息
func Wrapf(err error, errType ErrorType, code, format string, args ...interface{}) *AppError {
	return Wrap(err, errType, code, fmt.Sprintf(format, args...))
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	var parts []string
	
	if e.Code != "" {
		parts = append(parts, fmt.Sprintf("[%s]", e.Code))
	}
	
	if e.Type != "" {
		parts = append(parts, fmt.Sprintf("(%s)", e.Type))
	}
	
	if e.Message != "" {
		parts = append(parts, e.Message)
	}
	
	if e.Operation != "" {
		parts = append(parts, fmt.Sprintf("操作: %s", e.Operation))
	}
	
	if e.Err != nil {
		parts = append(parts, fmt.Sprintf("原因: %v", e.Err))
	}
	
	return strings.Join(parts, " ")
}

// Unwrap 实现 errors.Unwrap
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithOperation 设置操作名称
func (e *AppError) WithOperation(operation string) *AppError {
	e.Operation = operation
	return e
}

// WithContext 添加上下文信息
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithStackTrace 捕获堆栈跟踪
func (e *AppError) captureStackTrace() {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:]) // 跳过 captureStackTrace, New/Wrap, 和调用者
	frames := runtime.CallersFrames(pcs[:n])
	
	var stack []string
	for {
		frame, more := frames.Next()
		stack = append(stack, fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function))
		if !more {
			break
		}
	}
	
	e.StackTrace = strings.Join(stack, "\n")
}

// GetStackTrace 获取堆栈跟踪
func (e *AppError) GetStackTrace() string {
	return e.StackTrace
}

// Is 检查错误类型
func (e *AppError) Is(target error) bool {
	if target == nil {
		return false
	}
	
	// 检查是否是相同类型的 AppError
	var targetErr *AppError
	if errors.As(target, &targetErr) {
		return e.Type == targetErr.Type && e.Code == targetErr.Code
	}
	
	// 检查包装的错误
	return errors.Is(e.Err, target)
}

// ErrorBuilder 错误构建器
type ErrorBuilder struct {
	errType ErrorType
	code    string
	message string
	context map[string]interface{}
}

// NewErrorBuilder 创建错误构建器
func NewErrorBuilder(errType ErrorType) *ErrorBuilder {
	return &ErrorBuilder{
		errType: errType,
		context: make(map[string]interface{}),
	}
}

// WithCode 设置错误代码
func (b *ErrorBuilder) WithCode(code string) *ErrorBuilder {
	b.code = code
	return b
}

// WithMessage 设置错误消息
func (b *ErrorBuilder) WithMessage(message string) *ErrorBuilder {
	b.message = message
	return b
}

// WithMessagef 格式化错误消息
func (b *ErrorBuilder) WithMessagef(format string, args ...interface{}) *ErrorBuilder {
	b.message = fmt.Sprintf(format, args...)
	return b
}

// WithContext 添加上下文
func (b *ErrorBuilder) WithContext(key string, value interface{}) *ErrorBuilder {
	b.context[key] = value
	return b
}

// Build 构建错误
func (b *ErrorBuilder) Build() *AppError {
	err := NewAppError(b.errType, b.code, b.message)
	for key, value := range b.context {
		err.Context[key] = value
	}
	return err
}

// Wrap 包装错误
func (b *ErrorBuilder) Wrap(err error) *AppError {
	appErr := b.Build()
	appErr.Err = err
	appErr.captureStackTrace()
	return appErr
}

// Error codes
const (
	// 配置错误代码
	ErrCodeConfigNotFound    = "CONFIG_NOT_FOUND"
	ErrCodeConfigInvalid     = "CONFIG_INVALID"
	ErrCodeConfigParseFailed = "CONFIG_PARSE_FAILED"
	
	// 网络错误代码
	ErrCodeNetworkTimeout    = "NETWORK_TIMEOUT"
	ErrCodeNetworkUnreachable = "NETWORK_UNREACHABLE"
	ErrCodeConnectionFailed  = "CONNECTION_FAILED"
	
	// 存储错误代码
	ErrCodeStorageNotFound   = "STORAGE_NOT_FOUND"
	ErrCodeStorageAccessDenied = "STORAGE_ACCESS_DENIED"
	ErrCodeStorageQuotaExceeded = "STORAGE_QUOTA_EXCEEDED"
	ErrCodeStorageCorrupted  = "STORAGE_CORRUPTED"
	
	// IO错误代码
	ErrCodeFileNotFound      = "FILE_NOT_FOUND"
	ErrCodeFileReadFailed    = "FILE_READ_FAILED"
	ErrCodeFileWriteFailed   = "FILE_WRITE_FAILED"
	ErrCodeFilePermissionDenied = "FILE_PERMISSION_DENIED"
	
	// 验证错误代码
	ErrCodeValidationFailed  = "VALIDATION_FAILED"
	ErrCodeInvalidArgument   = "INVALID_ARGUMENT"
	ErrCodeMissingRequired   = "MISSING_REQUIRED"
	
	// 认证错误代码
	ErrCodeAuthFailed        = "AUTH_FAILED"
	ErrCodeTokenExpired      = "TOKEN_EXPIRED"
	ErrCodeInvalidCredentials = "INVALID_CREDENTIALS"
	
	// 超时错误代码
	ErrCodeOperationTimeout  = "OPERATION_TIMEOUT"
	ErrCodeRequestTimeout    = "REQUEST_TIMEOUT"
	
	// 未知错误代码
	ErrCodeUnknown           = "UNKNOWN_ERROR"
)

// Common error constructors
var (
	// Config errors
	ConfigNotFoundError = func(path string) *AppError {
		return NewErrorBuilder(ErrorTypeConfig).
			WithCode(ErrCodeConfigNotFound).
			WithMessagef("配置文件未找到: %s", path).
			WithContext("path", path).
			Build()
	}
	
	ConfigInvalidError = func(reason string) *AppError {
		return NewErrorBuilder(ErrorTypeConfig).
			WithCode(ErrCodeConfigInvalid).
			WithMessagef("配置文件无效: %s", reason).
			WithContext("reason", reason).
			Build()
	}
	
	// Network errors
	NetworkTimeoutError = func(operation string, timeout int) *AppError {
		return NewErrorBuilder(ErrorTypeNetwork).
			WithCode(ErrCodeNetworkTimeout).
			WithMessagef("网络操作超时: %s", operation).
			WithContext("operation", operation).
			WithContext("timeout_seconds", timeout).
			Build()
	}
	
	// Storage errors
	StorageNotFoundError = func(path string) *AppError {
		return NewErrorBuilder(ErrorTypeStorage).
			WithCode(ErrCodeStorageNotFound).
			WithMessagef("存储对象未找到: %s", path).
			WithContext("path", path).
			Build()
	}
	
	StorageAccessDeniedError = func(path string) *AppError {
		return NewErrorBuilder(ErrorTypeStorage).
			WithCode(ErrCodeStorageAccessDenied).
			WithMessagef("存储访问被拒绝: %s", path).
			WithContext("path", path).
			Build()
	}
	
	// IO errors
	FileNotFoundError = func(path string) *AppError {
		return NewErrorBuilder(ErrorTypeIO).
			WithCode(ErrCodeFileNotFound).
			WithMessagef("文件未找到: %s", path).
			WithContext("path", path).
			Build()
	}
	
	// Validation errors
	ValidationError = func(field, reason string) *AppError {
		return NewErrorBuilder(ErrorTypeValidation).
			WithCode(ErrCodeValidationFailed).
			WithMessagef("字段验证失败: %s - %s", field, reason).
			WithContext("field", field).
			WithContext("reason", reason).
			Build()
	}
	
	// Auth errors
	AuthFailedError = func(reason string) *AppError {
		return NewErrorBuilder(ErrorTypeAuth).
			WithCode(ErrCodeAuthFailed).
			WithMessagef("认证失败: %s", reason).
			WithContext("reason", reason).
			Build()
	}
	
	// Timeout errors
	OperationTimeoutError = func(operation string, duration int) *AppError {
		return NewErrorBuilder(ErrorTypeTimeout).
			WithCode(ErrCodeOperationTimeout).
			WithMessagef("操作超时: %s (%d秒)", operation, duration).
			WithContext("operation", operation).
			WithContext("duration_seconds", duration).
			Build()
	}
)

// Error handling utilities

// IsConfigError 检查是否为配置错误
func IsConfigError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeConfig
	}
	return false
}

// IsNetworkError 检查是否为网络错误
func IsNetworkError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeNetwork
	}
	return false
}

// IsStorageError 检查是否为存储错误
func IsStorageError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeStorage
	}
	return false
}

// IsIOError 检查是否为IO错误
func IsIOError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeIO
	}
	return false
}

// IsAuthError 检查是否为认证错误
func IsAuthError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeAuth
	}
	return false
}

// IsTimeoutError 检查是否为超时错误
func IsTimeoutError(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Type == ErrorTypeTimeout
	}
	return false
}

// ShouldRetry 检查错误是否应该重试
func ShouldRetry(err error) bool {
	// 网络错误通常可以重试
	if IsNetworkError(err) {
		return true
	}
	
	// 超时错误可以重试
	if IsTimeoutError(err) {
		return true
	}
	
	// 特定的存储错误可以重试
	var appErr *AppError
	if errors.As(err, &appErr) {
		switch appErr.Code {
		case ErrCodeStorageQuotaExceeded:
			// 配额超限，等待后重试
			return true
		case ErrCodeConnectionFailed:
			// 连接失败，可以重试
			return true
		}
	}
	
	return false
}

// GetErrorContext 获取错误的上下文信息
func GetErrorContext(err error) map[string]interface{} {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Context
	}
	return nil
}

// FormatError 格式化错误信息
func FormatError(err error) string {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Error()
	}
	return err.Error()
}

// LogError 记录错误日志
func LogError(err error, logger *Logger) {
	if logger == nil {
		logger = GetLogger()
	}
	
	var appErr *AppError
	if errors.As(err, &appErr) {
		// 记录详细的错误信息
		logger.Error("应用程序错误: %s", appErr.Error())
		
		// 记录上下文信息
		if len(appErr.Context) > 0 {
			for key, value := range appErr.Context {
				logger.Debug("错误上下文: %s = %v", key, value)
			}
		}
		
		// 记录堆栈跟踪（调试级别）
		if appErr.StackTrace != "" {
			logger.Debug("堆栈跟踪:\n%s", appErr.StackTrace)
		}
		
		// 记录原始错误
		if appErr.Err != nil {
			logger.Debug("原始错误: %v", appErr.Err)
		}
	} else {
		// 普通错误
		logger.Error("错误: %v", err)
	}
}