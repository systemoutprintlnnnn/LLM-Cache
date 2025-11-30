package middleware

import (
	"context"
	"llm-cache/pkg/logger"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDKey 请求ID在Context中的键名
const RequestIDKey = "request_id"

// LoggingConfig 日志中间件配置
type LoggingConfig struct {
	// SkipPaths 跳过日志记录的路径（如健康检查接口）
	SkipPaths []string
	// IncludeRequestBody 是否记录请求体
	IncludeRequestBody bool
	// IncludeResponseBody 是否记录响应体
	IncludeResponseBody bool
	// Logger 日志器实例
	Logger logger.Logger
}

// LoggingMiddleware 返回HTTP日志记录中间件
// config: 中间件配置，如果为nil则使用默认配置
func LoggingMiddleware(config *LoggingConfig) gin.HandlerFunc {
	// 使用默认配置
	if config == nil {
		config = &LoggingConfig{
			SkipPaths:           []string{"/health", "/metrics"},
			IncludeRequestBody:  false,
			IncludeResponseBody: false,
			Logger:              logger.GetDefault(),
		}
	}

	// 如果没有指定Logger，使用默认Logger
	if config.Logger == nil {
		config.Logger = logger.GetDefault()
	}

	return func(c *gin.Context) {
		// 生成请求ID
		requestID := generateRequestID()
		c.Set(RequestIDKey, requestID)

		// 检查是否需要跳过日志记录
		if shouldSkipPath(c.Request.URL.Path, config.SkipPaths) {
			c.Next()
			return
		}

		// 记录请求开始时间
		startTime := time.Now()

		// 提取请求信息
		requestInfo := extractRequestInfo(c, requestID)

		// 创建带有请求ID的context
		ctx := context.WithValue(c.Request.Context(), RequestIDKey, requestID)
		c.Request = c.Request.WithContext(ctx)

		// 记录请求开始日志
		config.Logger.InfoContext(ctx, "HTTP请求开始",
			"request_id", requestID,
			"method", requestInfo.Method,
			"path", requestInfo.Path,
			"client_ip", requestInfo.ClientIP,
			"user_agent", requestInfo.UserAgent,
			"content_length", requestInfo.ContentLength,
			"query_params", requestInfo.QueryParams,
		)

		// 创建自定义ResponseWriter来捕获响应信息
		responseWriter := &responseWriter{
			ResponseWriter: c.Writer,
			body:           make([]byte, 0),
		}
		c.Writer = responseWriter

		// 执行请求处理
		c.Next()

		// 计算处理时间
		duration := time.Since(startTime)
		statusCode := c.Writer.Status()

		// 提取响应信息
		responseInfo := extractResponseInfo(c, responseWriter, duration)

		config.Logger.InfoContext(ctx, "HTTP请求完成",
			"request_id", requestID,
			"method", requestInfo.Method,
			"path", requestInfo.Path,
			"status_code", statusCode,
			"duration_ms", responseInfo.DurationMs,
			"response_size", responseInfo.ResponseSize,
			"client_ip", requestInfo.ClientIP,
		)

		// 如果有错误，记录详细错误信息
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				config.Logger.ErrorContext(ctx, "HTTP请求处理错误",
					"request_id", requestID,
					"error", err.Error(),
					"error_type", err.Type,
				)
			}
		}
	}
}

// RequestInfo HTTP请求信息
type RequestInfo struct {
	Method        string            `json:"method"`
	Path          string            `json:"path"`
	ClientIP      string            `json:"client_ip"`
	UserAgent     string            `json:"user_agent"`
	ContentLength int64             `json:"content_length"`
	QueryParams   map[string]string `json:"query_params"`
	Headers       map[string]string `json:"headers"`
}

// ResponseInfo HTTP响应信息
type ResponseInfo struct {
	StatusCode   int     `json:"status_code"`
	DurationMs   float64 `json:"duration_ms"`
	ResponseSize int     `json:"response_size"`
}

// responseWriter 自定义ResponseWriter，用于捕获响应数据
type responseWriter struct {
	gin.ResponseWriter
	body []byte
}

// Write 重写Write方法，捕获响应体
func (w *responseWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return w.ResponseWriter.Write(b)
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return uuid.New().String()
}

// extractRequestInfo 提取请求信息
func extractRequestInfo(c *gin.Context, requestID string) *RequestInfo {
	// 提取查询参数
	queryParams := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			queryParams[key] = values[0] // 只取第一个值
		}
	}

	// 提取重要的请求头
	headers := make(map[string]string)
	importantHeaders := []string{"Content-Type", "Accept", "Authorization", "X-Forwarded-For"}
	for _, header := range importantHeaders {
		if value := c.GetHeader(header); value != "" {
			headers[header] = value
		}
	}

	return &RequestInfo{
		Method:        c.Request.Method,
		Path:          c.Request.URL.Path,
		ClientIP:      c.ClientIP(),
		UserAgent:     c.GetHeader("User-Agent"),
		ContentLength: c.Request.ContentLength,
		QueryParams:   queryParams,
		Headers:       headers,
	}
}

// extractResponseInfo 提取响应信息
func extractResponseInfo(c *gin.Context, rw *responseWriter, duration time.Duration) *ResponseInfo {
	return &ResponseInfo{
		StatusCode:   c.Writer.Status(),
		DurationMs:   float64(duration.Nanoseconds()) / 1e6, // 转换为毫秒
		ResponseSize: len(rw.body),
	}
}

// shouldSkipPath 检查是否应该跳过某个路径的日志记录
func shouldSkipPath(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// GetRequestID 从Context中获取请求ID
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDKey); exists {
		return requestID.(string)
	}
	return ""
}
