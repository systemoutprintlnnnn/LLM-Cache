// Package logger 提供基于 logrus 的日志记录器接口和实现。
// 支持字段注入、字段提取、JSON 格式输出和日志文件轮转。
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

// Fields 定义日志字段类型，用于结构化日志记录。
type Fields map[string]interface{}

// Logger 定义通用的日志记录器接口。
// 提供了一组标准的方法来记录不同级别（Debug, Info, Warn, Error）的日志，
// 并支持携带上下文信息（Context）以便于链路追踪。
// 还提供字段注入和提取功能，支持链式调用。
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

	// 字段注入方法 - 注入后返回新的 Logger 实例，后续日志自动包含这些字段
	WithFields(fields Fields) Logger
	WithField(key string, value interface{}) Logger

	// 字段提取方法 - 获取当前已注入的所有字段
	GetFields() Fields

	// 获取底层 logrus.Entry，用于与现有基础设施兼容
	LogrusEntry() *logrus.Entry
}

// Level 定义日志级别类型。
type Level int

const (
	// DebugLevel 调试级别，记录最详细的日志信息。
	DebugLevel Level = iota
	// InfoLevel 信息级别，记录常规操作信息。
	InfoLevel
	// WarnLevel 警告级别，记录潜在问题。
	WarnLevel
	// ErrorLevel 错误级别，记录错误信息。
	ErrorLevel
)

// Config 定义日志系统的配置参数。
// 包含日志级别、输出目标（标准输出或文件）、格式以及文件轮转设置。
type Config struct {
	Level      Level  // 日志级别
	Output     string // 输出：stdout、stderr、file
	FilePath   string // 文件路径（当Output为file时）
	Format     string // 格式：text、json
	MaxSize    int    // 单个日志文件最大大小(MB)
	MaxBackups int    // 保留的旧日志文件数量
	MaxAge     int    // 日志文件保留天数
	Compress   bool   // 是否压缩旧日志文件
}

// appLogger 日志器接口的具体实现，封装了 logrus.Entry。
type appLogger struct {
	entry  *logrus.Entry  // logrus Entry (带预设字段)
	logger *logrus.Logger // 底层 logrus Logger
	fields Fields         // 存储已注入的字段，用于提取
}

// Default 创建并返回一个使用默认配置的 Logger 实例。
// 默认配置为：输出到 stdout，级别为 Info，格式为 text。
func Default() Logger {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339,
	})

	return &appLogger{
		entry:  logrus.NewEntry(logger),
		logger: logger,
		fields: make(Fields),
	}
}

// New 根据提供的配置创建一个新的 Logger 实例。
// 如果配置了文件输出，会自动创建目录和文件，并支持日志轮转。
// 参数 config: 日志配置。
// 返回: Logger 接口实例。
func New(config Config) Logger {
	logger := logrus.New()

	// 设置日志级别
	logger.SetLevel(convertLevel(config.Level))

	// 设置输出格式
	logger.SetFormatter(createFormatter(config.Format))

	// 设置输出目标
	logger.SetOutput(getWriter(config))

	return &appLogger{
		entry:  logrus.NewEntry(logger),
		logger: logger,
		fields: make(Fields),
	}
}

// convertLevel 将自定义 Level 转换为 logrus.Level。
func convertLevel(level Level) logrus.Level {
	switch level {
	case DebugLevel:
		return logrus.DebugLevel
	case InfoLevel:
		return logrus.InfoLevel
	case WarnLevel:
		return logrus.WarnLevel
	case ErrorLevel:
		return logrus.ErrorLevel
	default:
		return logrus.InfoLevel
	}
}

// createFormatter 根据配置创建日志格式化器。
func createFormatter(format string) logrus.Formatter {
	switch format {
	case "json":
		return &logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
			},
		}
	default: // "text" 或其他
		return &logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		}
	}
}

// getWriter 内部辅助函数：获取日志输出的 Writer。
// 支持 stdout、stderr 和带轮转的文件输出。
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

		// 使用 lumberjack 进行日志轮转
		return &lumberjack.Logger{
			Filename:   config.FilePath,
			MaxSize:    getMaxSize(config.MaxSize),
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		}
	default:
		return os.Stdout
	}
}

// getMaxSize 获取日志文件最大大小，默认为 100MB。
func getMaxSize(size int) int {
	if size <= 0 {
		return 100 // 默认 100MB
	}
	return size
}

// argsToFields 将 key-value 对参数转换为 Fields。
// 参数格式为：key1, value1, key2, value2, ...
func argsToFields(args []interface{}) logrus.Fields {
	fields := make(logrus.Fields)
	for i := 0; i < len(args)-1; i += 2 {
		if key, ok := args[i].(string); ok {
			fields[key] = args[i+1]
		}
	}
	return fields
}

// Debug 记录 Debug 级别的日志。
func (l *appLogger) Debug(msg string, args ...interface{}) {
	if len(args) > 0 {
		l.entry.WithFields(argsToFields(args)).Debug(msg)
	} else {
		l.entry.Debug(msg)
	}
}

// Info 记录 Info 级别的日志。
func (l *appLogger) Info(msg string, args ...interface{}) {
	if len(args) > 0 {
		l.entry.WithFields(argsToFields(args)).Info(msg)
	} else {
		l.entry.Info(msg)
	}
}

// Warn 记录 Warn 级别的日志。
func (l *appLogger) Warn(msg string, args ...interface{}) {
	if len(args) > 0 {
		l.entry.WithFields(argsToFields(args)).Warn(msg)
	} else {
		l.entry.Warn(msg)
	}
}

// Error 记录 Error 级别的日志。
func (l *appLogger) Error(msg string, args ...interface{}) {
	if len(args) > 0 {
		l.entry.WithFields(argsToFields(args)).Error(msg)
	} else {
		l.entry.Error(msg)
	}
}

// DebugContext 记录带上下文的 Debug 级别日志。
// 自动从 context 中提取 request_id 和 trace_id。
func (l *appLogger) DebugContext(ctx context.Context, msg string, args ...interface{}) {
	entry := l.entryWithContext(ctx)
	if len(args) > 0 {
		entry.WithFields(argsToFields(args)).Debug(msg)
	} else {
		entry.Debug(msg)
	}
}

// InfoContext 记录带上下文的 Info 级别日志。
// 自动从 context 中提取 request_id 和 trace_id。
func (l *appLogger) InfoContext(ctx context.Context, msg string, args ...interface{}) {
	entry := l.entryWithContext(ctx)
	if len(args) > 0 {
		entry.WithFields(argsToFields(args)).Info(msg)
	} else {
		entry.Info(msg)
	}
}

// WarnContext 记录带上下文的 Warn 级别日志。
// 自动从 context 中提取 request_id 和 trace_id。
func (l *appLogger) WarnContext(ctx context.Context, msg string, args ...interface{}) {
	entry := l.entryWithContext(ctx)
	if len(args) > 0 {
		entry.WithFields(argsToFields(args)).Warn(msg)
	} else {
		entry.Warn(msg)
	}
}

// ErrorContext 记录带上下文的 Error 级别日志。
// 自动从 context 中提取 request_id 和 trace_id。
func (l *appLogger) ErrorContext(ctx context.Context, msg string, args ...interface{}) {
	entry := l.entryWithContext(ctx)
	if len(args) > 0 {
		entry.WithFields(argsToFields(args)).Error(msg)
	} else {
		entry.Error(msg)
	}
}

// entryWithContext 从 context 中提取常用字段并返回带这些字段的 entry。
func (l *appLogger) entryWithContext(ctx context.Context) *logrus.Entry {
	entry := l.entry

	// 提取 request_id
	if requestID := ctx.Value("request_id"); requestID != nil {
		entry = entry.WithField("request_id", requestID)
	}

	// 提取 trace_id
	if traceID := ctx.Value("trace_id"); traceID != nil {
		entry = entry.WithField("trace_id", traceID)
	}

	// 提取 span_id
	if spanID := ctx.Value("span_id"); spanID != nil {
		entry = entry.WithField("span_id", spanID)
	}

	return entry
}

// WithFields 注入多个字段，返回新的 Logger 实例。
// 注入后，该实例的所有后续日志都会自动包含这些字段。
func (l *appLogger) WithFields(fields Fields) Logger {
	// 合并现有字段和新字段
	mergedFields := make(Fields)
	for k, v := range l.fields {
		mergedFields[k] = v
	}
	for k, v := range fields {
		mergedFields[k] = v
	}

	return &appLogger{
		entry:  l.entry.WithFields(logrus.Fields(fields)),
		logger: l.logger,
		fields: mergedFields,
	}
}

// WithField 注入单个字段，返回新的 Logger 实例。
func (l *appLogger) WithField(key string, value interface{}) Logger {
	return l.WithFields(Fields{key: value})
}

// GetFields 提取当前已注入的所有字段。
// 返回字段的副本，防止外部修改。
func (l *appLogger) GetFields() Fields {
	result := make(Fields)
	for k, v := range l.fields {
		result[k] = v
	}
	return result
}

// LogrusEntry 获取底层的 logrus.Entry 实例。
func (l *appLogger) LogrusEntry() *logrus.Entry {
	return l.entry
}

// 全局默认日志器
var defaultLogger Logger = Default()

// GetDefault 获取全局唯一的默认 Logger 实例。
// 该实例通常在应用程序启动初期使用，或者作为 fallback。
func GetDefault() Logger {
	return defaultLogger
}

// SetDefault 设置全局默认 Logger 实例。
func SetDefault(l Logger) {
	defaultLogger = l
}
