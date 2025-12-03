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

// Logger 定义通用的日志记录器接口。
// 提供了一组标准的方法来记录不同级别（Debug, Info, Warn, Error）的日志，
// 并支持携带上下文信息（Context）以便于链路追踪。
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

// Config 定义日志系统的配置参数。
// 包含日志级别、输出目标（标准输出或文件）以及文件路径。
type Config struct {
	Level    slog.Level // 日志级别
	Output   string     // 输出：stdout、stderr、file
	FilePath string     // 文件路径（当Output为file时）
}

// appLogger 日志器接口的具体实现，封装了 slog.Logger。
type appLogger struct {
	logger *slog.Logger
}

// Default 创建并返回一个使用默认配置的 Logger 实例。
// 默认配置为：输出到 stdout，级别为 Info。
func Default() Logger {
	return &appLogger{
		logger: slog.Default(),
	}
}

// New 根据提供的配置创建一个新的 Logger 实例。
// 如果配置了文件输出，会自动创建目录和文件。
// 参数 config: 日志配置。
// 返回: Logger 接口实例。
func New(config Config) Logger {
	handler := createHandler(config)
	return &appLogger{
		logger: slog.New(handler),
	}
}

// createHandler 内部辅助函数：根据配置创建 slog.Handler。
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

// getWriter 内部辅助函数：获取日志输出的 Writer。
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

// Debug 记录 Debug 级别的日志。
func (l *appLogger) Debug(msg string, args ...interface{}) {
	l.logger.Debug(msg, args...)
}

// Info 记录 Info 级别的日志。
func (l *appLogger) Info(msg string, args ...interface{}) {
	l.logger.Info(msg, args...)
}

// Warn 记录 Warn 级别的日志。
func (l *appLogger) Warn(msg string, args ...interface{}) {
	l.logger.Warn(msg, args...)
}

// Error 记录 Error 级别的日志。
func (l *appLogger) Error(msg string, args ...interface{}) {
	l.logger.Error(msg, args...)
}

// DebugContext 记录带上下文的 Debug 级别日志。
func (l *appLogger) DebugContext(ctx context.Context, msg string, args ...interface{}) {
	l.logger.DebugContext(ctx, msg, args...)
}

// InfoContext 记录带上下文的 Info 级别日志。
func (l *appLogger) InfoContext(ctx context.Context, msg string, args ...interface{}) {
	l.logger.InfoContext(ctx, msg, args...)
}

// WarnContext 记录带上下文的 Warn 级别日志。
func (l *appLogger) WarnContext(ctx context.Context, msg string, args ...interface{}) {
	l.logger.WarnContext(ctx, msg, args...)
}

// ErrorContext 记录带上下文的 Error 级别日志。
func (l *appLogger) ErrorContext(ctx context.Context, msg string, args ...interface{}) {
	l.logger.ErrorContext(ctx, msg, args...)
}

// SlogLogger 获取底层的 slog.Logger 实例。
func (l *appLogger) SlogLogger() *slog.Logger {
	return l.logger
}

// 全局默认日志器
var defaultLogger Logger = Default()

// GetDefault 获取全局唯一的默认 Logger 实例。
// 该实例通常在应用程序启动初期使用，或者作为 fallback。
func GetDefault() Logger {
	return defaultLogger
}
