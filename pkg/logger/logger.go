// 日志接口与实现（基于 logrus，支持上下文字段注入与 JSON 输出）
package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Fields 是日志字段的统一类型。
type Fields map[string]interface{}

// Logger 定义通用的日志记录器接口。
// 提供了一组标准方法来记录不同级别的日志，并支持上下文字段自动合并。
type Logger interface {
	// 基础日志方法（不依赖 context）
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})

	// 带上下文的日志方法，会自动合并通过 InjectFields 注入的字段
	DebugContext(ctx context.Context, msg string, args ...interface{})
	InfoContext(ctx context.Context, msg string, args ...interface{})
	WarnContext(ctx context.Context, msg string, args ...interface{})
	ErrorContext(ctx context.Context, msg string, args ...interface{})
}

// Config 定义日志配置。
// 兼容现有的 logging 配置，并新增 JSON/滚动参数支持。
type Config struct {
	Level        logrus.Level
	Output       string // stdout、stderr、file
	FilePath     string
	MaxSize      int // MB
	MaxBackups   int
	MaxAge       int // days
	Compress     bool
	ReportCaller bool
	JSONFormat   bool // 默认 true
}

// appLogger 是 Logger 的具体实现，封装 logrus。
type appLogger struct {
	base *logrus.Logger
}

// Default 返回默认 Logger（stdout、info、JSON）。
func Default() Logger {
	return New(Config{
		Level:      logrus.InfoLevel,
		Output:     "stdout",
		JSONFormat: true,
	})
}

// New 根据配置创建 Logger。
func New(config Config) Logger {
	l := logrus.New()
	l.SetLevel(config.Level)
	l.SetReportCaller(config.ReportCaller)
	l.SetOutput(resolveWriter(config))

	if config.JSONFormat {
		l.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
		})
	} else {
		l.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339Nano,
		})
	}

	return &appLogger{base: l}
}

// resolveWriter 根据配置选择输出。
func resolveWriter(config Config) io.Writer {
	switch config.Output {
	case "stderr":
		return os.Stderr
	case "file":
		if config.FilePath == "" {
			fmt.Fprintf(os.Stderr, "日志文件路径为空，回退到 stdout\n")
			return os.Stdout
		}
		dir := filepath.Dir(config.FilePath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "创建日志目录失败: %v，回退到 stdout\n", err)
			return os.Stdout
		}
		return &lumberjack.Logger{
			Filename:   config.FilePath,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		}
	default:
		return os.Stdout
	}
}

// Debug 记录 Debug 级别日志。
func (l *appLogger) Debug(msg string, args ...interface{}) {
	l.logWithContext(context.Background(), logrus.DebugLevel, msg, args...)
}

// Info 记录 Info 级别日志。
func (l *appLogger) Info(msg string, args ...interface{}) {
	l.logWithContext(context.Background(), logrus.InfoLevel, msg, args...)
}

// Warn 记录 Warn 级别日志。
func (l *appLogger) Warn(msg string, args ...interface{}) {
	l.logWithContext(context.Background(), logrus.WarnLevel, msg, args...)
}

// Error 记录 Error 级别日志。
func (l *appLogger) Error(msg string, args ...interface{}) {
	l.logWithContext(context.Background(), logrus.ErrorLevel, msg, args...)
}

// DebugContext 记录 Debug 级别日志，自动合并上下文字段。
func (l *appLogger) DebugContext(ctx context.Context, msg string, args ...interface{}) {
	l.logWithContext(ctx, logrus.DebugLevel, msg, args...)
}

// InfoContext 记录 Info 级别日志，自动合并上下文字段。
func (l *appLogger) InfoContext(ctx context.Context, msg string, args ...interface{}) {
	l.logWithContext(ctx, logrus.InfoLevel, msg, args...)
}

// WarnContext 记录 Warn 级别日志，自动合并上下文字段。
func (l *appLogger) WarnContext(ctx context.Context, msg string, args ...interface{}) {
	l.logWithContext(ctx, logrus.WarnLevel, msg, args...)
}

// ErrorContext 记录 Error 级别日志，自动合并上下文字段。
func (l *appLogger) ErrorContext(ctx context.Context, msg string, args ...interface{}) {
	l.logWithContext(ctx, logrus.ErrorLevel, msg, args...)
}

// logWithContext 将上下文字段与显式字段合并后输出。
func (l *appLogger) logWithContext(ctx context.Context, level logrus.Level, msg string, args ...interface{}) {
	entry := l.base.WithContext(ctx)
	fields := mergeFields(ExtractFields(ctx), parseArgs(args))
	if len(fields) > 0 {
		entry = entry.WithFields(logrus.Fields(fields))
	}
	entry.Log(level, msg)
}

// parseArgs 将可变参数解析为字段，期望 key,value 形式。
func parseArgs(args []interface{}) Fields {
	fields := Fields{}
	length := len(args)
	if length == 0 {
		return fields
	}

	if length%2 != 0 {
		fields["logger_args_error"] = "expected even number of arguments"
		length-- // 尝试解析前面的偶数部分
	}

	for i := 0; i < length; i += 2 {
		key, ok := args[i].(string)
		if !ok {
			fields[fmt.Sprintf("arg_%d", i)] = args[i+1]
			continue
		}
		fields[key] = args[i+1]
	}
	return fields
}

// mergeFields 合并上下文字段与本次日志字段，后者覆盖同名键。
func mergeFields(ctxFields Fields, logFields Fields) Fields {
	if len(ctxFields) == 0 && len(logFields) == 0 {
		return nil
	}
	result := Fields{}
	for k, v := range ctxFields {
		result[k] = v
	}
	for k, v := range logFields {
		result[k] = v
	}
	return result
}

// InjectFields 将字段注入到 context，后续日志会自动带上。
// 如果上下文为空，会使用 context.Background()。
func InjectFields(ctx context.Context, fields Fields) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if len(fields) == 0 {
		return ctx
	}
	current := ExtractFields(ctx)
	merged := Fields{}
	for k, v := range current {
		merged[k] = v
	}
	for k, v := range fields {
		merged[k] = v
	}
	return context.WithValue(ctx, fieldsContextKey, merged)
}

// ExtractFields 从 context 中提取已注入的字段。
// 返回的 map 是副本，可安全修改。
func ExtractFields(ctx context.Context) Fields {
	if ctx == nil {
		return Fields{}
	}
	if v := ctx.Value(fieldsContextKey); v != nil {
		if fields, ok := v.(Fields); ok && fields != nil {
			copyFields := Fields{}
			for k, val := range fields {
				copyFields[k] = val
			}
			return copyFields
		}
	}
	return Fields{}
}

// ToArgs 将字段转换为可用于日志方法的 key,value 形式。
func (f Fields) ToArgs() []interface{} {
	args := make([]interface{}, 0, len(f)*2)
	for k, v := range f {
		args = append(args, k, v)
	}
	return args
}

// 上下文字段使用的 key，避免冲突
type contextKey string

const fieldsContextKey contextKey = "logger_fields"

// 全局默认日志器
var defaultLogger Logger = Default()

// GetDefault 获取全局唯一的默认 Logger 实例。
// 该实例通常在应用程序启动初期使用，或者作为 fallback。
func GetDefault() Logger {
	return defaultLogger
}
