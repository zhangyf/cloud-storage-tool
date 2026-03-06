// Package utils 提供工具函数和日志系统
package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// LogLevel 日志级别类型
type LogLevel int

const (
	// LevelDebug 调试级别，最详细的日志
	LevelDebug LogLevel = iota
	
	// LevelInfo 信息级别，常规操作日志
	LevelInfo
	
	// LevelWarn 警告级别，需要注意但不影响程序运行的问题
	LevelWarn
	
	// LevelError 错误级别，程序运行错误
	LevelError
	
	// LevelFatal 致命级别，程序无法继续运行
	LevelFatal
)

// String 返回日志级别的字符串表示
func (l LogLevel) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger 日志记录器
type Logger struct {
	mu        sync.Mutex
	level     LogLevel
	out       io.Writer
	callDepth int
	prefix    string
	colors    bool
}

// NewLogger 创建新的日志记录器
func NewLogger(level LogLevel, out io.Writer) *Logger {
	return &Logger{
		level:     level,
		out:       out,
		callDepth: 3, // 默认调用深度
		colors:    runtime.GOOS != "windows", // Windows默认不启用颜色
	}
}

// SetLevel 设置日志级别
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetOutput 设置输出目标
func (l *Logger) SetOutput(out io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.out = out
}

// SetCallDepth 设置调用深度
func (l *Logger) SetCallDepth(depth int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.callDepth = depth
}

// SetPrefix 设置日志前缀
func (l *Logger) SetPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = prefix
}

// EnableColors 启用颜色输出
func (l *Logger) EnableColors() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.colors = true
}

// DisableColors 禁用颜色输出
func (l *Logger) DisableColors() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.colors = false
}

// log 内部日志方法
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// 获取调用者信息
	_, file, line, ok := runtime.Caller(l.callDepth)
	if !ok {
		file = "???"
		line = 0
	}

	// 提取文件名（不含路径）
	fileName := filepath.Base(file)

	// 构建日志消息
	message := fmt.Sprintf(format, args...)
	
	// 构建完整日志行
	var logLine string
	if l.colors {
		logLine = l.colorizedLogLine(level, fileName, line, message)
	} else {
		logLine = l.plainLogLine(level, fileName, line, message)
	}

	// 添加前缀
	if l.prefix != "" {
		logLine = l.prefix + " " + logLine
	}

	// 输出日志
	fmt.Fprintln(l.out, logLine)
}

// colorizedLogLine 构建带颜色的日志行
func (l *Logger) colorizedLogLine(level LogLevel, file string, line int, message string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	
	// 根据级别选择颜色
	var levelColor, resetColor string
	switch level {
	case LevelDebug:
		levelColor = "\033[36m" // 青色
	case LevelInfo:
		levelColor = "\033[32m" // 绿色
	case LevelWarn:
		levelColor = "\033[33m" // 黄色
	case LevelError:
		levelColor = "\033[31m" // 红色
	case LevelFatal:
		levelColor = "\033[35m" // 紫色
	}
	resetColor = "\033[0m"
	
	return fmt.Sprintf("%s [%s%s%s] %s:%d - %s", 
		timestamp, levelColor, level.String(), resetColor, file, line, message)
}

// plainLogLine 构建普通日志行
func (l *Logger) plainLogLine(level LogLevel, file string, line int, message string) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	return fmt.Sprintf("%s [%s] %s:%d - %s", 
		timestamp, level.String(), file, line, message)
}

// Debug 记录调试级别日志
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

// Info 记录信息级别日志
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

// Warn 记录警告级别日志
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

// Error 记录错误级别日志
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// Fatal 记录致命级别日志并退出程序
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(LevelFatal, format, args...)
	os.Exit(1)
}

// WithField 创建带字段的日志记录器
func (l *Logger) WithField(key, value string) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	newLogger := *l
	if newLogger.prefix == "" {
		newLogger.prefix = fmt.Sprintf("[%s=%s]", key, value)
	} else {
		newLogger.prefix = fmt.Sprintf("%s [%s=%s]", newLogger.prefix, key, value)
	}
	
	return &newLogger
}

// WithFields 创建带多个字段的日志记录器
func (l *Logger) WithFields(fields map[string]string) *Logger {
	l.mu.Lock()
	defer l.mu.Unlock()
	
	newLogger := *l
	var fieldStrs []string
	for key, value := range fields {
		fieldStrs = append(fieldStrs, fmt.Sprintf("%s=%s", key, value))
	}
	
	if len(fieldStrs) > 0 {
		prefix := "[" + strings.Join(fieldStrs, " ") + "]"
		if newLogger.prefix == "" {
			newLogger.prefix = prefix
		} else {
			newLogger.prefix = fmt.Sprintf("%s %s", newLogger.prefix, prefix)
		}
	}
	
	return &newLogger
}

// FileLogger 文件日志记录器
type FileLogger struct {
	*Logger
	file     *os.File
	filePath string
	maxSize  int64 // 最大文件大小（字节）
	maxFiles int   // 最大文件数量
}

// NewFileLogger 创建文件日志记录器
func NewFileLogger(filePath string, level LogLevel, maxSize int64, maxFiles int) (*FileLogger, error) {
	// 创建日志目录
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}
	
	// 打开日志文件
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("打开日志文件失败: %w", err)
	}
	
	logger := NewLogger(level, file)
	logger.DisableColors() // 文件日志不需要颜色
	
	return &FileLogger{
		Logger:   logger,
		file:     file,
		filePath: filePath,
		maxSize:  maxSize,
		maxFiles: maxFiles,
	}, nil
}

// rotate 日志轮转
func (fl *FileLogger) rotate() error {
	// 检查文件大小
	info, err := fl.file.Stat()
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %w", err)
	}
	
	if info.Size() < fl.maxSize {
		return nil // 不需要轮转
	}
	
	// 关闭当前文件
	if err := fl.file.Close(); err != nil {
		return fmt.Errorf("关闭日志文件失败: %w", err)
	}
	
	// 轮转旧文件
	for i := fl.maxFiles - 1; i > 0; i-- {
		oldPath := fmt.Sprintf("%s.%d", fl.filePath, i)
		newPath := fmt.Sprintf("%s.%d", fl.filePath, i+1)
		
		if _, err := os.Stat(oldPath); err == nil {
			if err := os.Rename(oldPath, newPath); err != nil {
				return fmt.Errorf("重命名日志文件失败: %w", err)
			}
		}
	}
	
	// 重命名当前文件
	backupPath := fmt.Sprintf("%s.1", fl.filePath)
	if err := os.Rename(fl.filePath, backupPath); err != nil {
		return fmt.Errorf("备份日志文件失败: %w", err)
	}
	
	// 重新打开日志文件
	file, err := os.OpenFile(fl.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("重新打开日志文件失败: %w", err)
	}
	
	fl.file = file
	fl.SetOutput(file)
	
	return nil
}

// log 重写日志方法以支持轮转
func (fl *FileLogger) log(level LogLevel, format string, args ...interface{}) {
	// 检查是否需要轮转
	if err := fl.rotate(); err != nil {
		// 轮转失败，输出到标准错误
		fmt.Fprintf(os.Stderr, "日志轮转失败: %v\n", err)
	}
	
	// 调用父类的日志方法
	fl.Logger.log(level, format, args...)
}

// Close 关闭文件日志记录器
func (fl *FileLogger) Close() error {
	return fl.file.Close()
}

// GlobalLogger 全局日志记录器实例
var GlobalLogger *Logger

// InitGlobalLogger 初始化全局日志记录器
func InitGlobalLogger(level LogLevel) {
	GlobalLogger = NewLogger(level, os.Stdout)
}

// GetLogger 获取全局日志记录器
func GetLogger() *Logger {
	if GlobalLogger == nil {
		InitGlobalLogger(LevelInfo)
	}
	return GlobalLogger
}

// Debug 全局调试日志
func Debug(format string, args ...interface{}) {
	GetLogger().Debug(format, args...)
}

// Info 全局信息日志
func Info(format string, args ...interface{}) {
	GetLogger().Info(format, args...)
}

// Warn 全局警告日志
func Warn(format string, args ...interface{}) {
	GetLogger().Warn(format, args...)
}

// Error 全局错误日志
func Error(format string, args ...interface{}) {
	GetLogger().Error(format, args...)
}

// Fatal 全局致命日志
func Fatal(format string, args ...interface{}) {
	GetLogger().Fatal(format, args...)
}

// StdLogger 返回标准库的日志记录器
func StdLogger() *log.Logger {
	return log.New(GetLogger().out, "", 0)
}