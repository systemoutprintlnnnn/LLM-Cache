// 日志接口
package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// Logger 日志器接口
type Logger interface {
	// 基础日志方法
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})

	// 带上下文的日志方法
	DebugContext(ctx context.Context, msg string, args ...interface{})
	InfoContext(ctx context.Context, msg string, args ...interface{})
	WarnContext(ctx context.Context, msg string, args ...interface{})
	ErrorContext(ctx context.Context, msg string, args ...interface{})

	// 获取底层slog.Logger，用于与现有基础设施兼容
	SlogLogger() *slog.Logger
}

// Config 简化的日志配置
type Config struct {
	Level    slog.Level // 日志级别
	Output   string     // 输出：stdout、stderr、file
	FilePath string     // 文件路径（当Output为file时）
}

// appLogger 日志器实现
type appLogger struct {
	logger *slog.Logger
}

// Default 创建默认日志器 - 使用Go的默认配置，不做任何修改
func Default() Logger {
	return &appLogger{
		logger: slog.Default(),
	}
}

// New 根据配置创建日志器
func New(config Config) Logger {
	handler := createHandler(config)
	return &appLogger{
		logger: slog.New(handler),
	}
}

// createHandler 创建简单的Handler
func createHandler(config Config) slog.Handler {
	// 基本配置 - 与默认slog保持一致
	opts := &slog.HandlerOptions{
		Level:       config.Level,
		AddSource:   false,
		ReplaceAttr: nil,
	}

	// 获取输出Writer
	writer := getWriter(config)

	// 统一使用TextHandler
	return slog.NewTextHandler(writer, opts)
}

// getWriter 获取输出Writer
func getWriter(config Config) io.Writer {
	switch config.Output {
	case "stdout":
		return os.Stdout
	case "stderr":
		return os.Stderr
	case "file":
		if config.FilePath == "" {
			return os.Stdout
		}

		// 确保目录存在
		dir := filepath.Dir(config.FilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "创建日志目录失败: %v\n", err)
			return os.Stdout
		}

		// 打开文件
		file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "打开日志文件失败: %v\n", err)
			return os.Stdout
		}
		return file
	default:
		return os.Stdout
	}
}

// Debug 记录Debug级别日志
func (l *appLogger) Debug(msg string, args ...interface{}) {
	l.logger.Debug(msg, args...)
}

// Info 记录Info级别日志
func (l *appLogger) Info(msg string, args ...interface{}) {
	l.logger.Info(msg, args...)
}

// Warn 记录Warn级别日志
func (l *appLogger) Warn(msg string, args ...interface{}) {
	l.logger.Warn(msg, args...)
}

// Error 记录Error级别日志
func (l *appLogger) Error(msg string, args ...interface{}) {
	l.logger.Error(msg, args...)
}

// DebugContext 记录带上下文的Debug级别日志
func (l *appLogger) DebugContext(ctx context.Context, msg string, args ...interface{}) {
	l.logger.DebugContext(ctx, msg, args...)
}

// InfoContext 记录带上下文的Info级别日志
func (l *appLogger) InfoContext(ctx context.Context, msg string, args ...interface{}) {
	l.logger.InfoContext(ctx, msg, args...)
}

// WarnContext 记录带上下文的Warn级别日志
func (l *appLogger) WarnContext(ctx context.Context, msg string, args ...interface{}) {
	l.logger.WarnContext(ctx, msg, args...)
}

// ErrorContext 记录带上下文的Error级别日志
func (l *appLogger) ErrorContext(ctx context.Context, msg string, args ...interface{}) {
	l.logger.ErrorContext(ctx, msg, args...)
}

// SlogLogger 获取底层slog.Logger
func (l *appLogger) SlogLogger() *slog.Logger {
	return l.logger
}

// 全局默认日志器
var defaultLogger Logger = Default()

// GetDefault 获取默认日志器
func GetDefault() Logger {
	return defaultLogger
}
