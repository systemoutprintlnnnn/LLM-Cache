package handlers

import (
	"net/http"
	"strings"
	"time"

	"llm-cache/internal/app/middleware"
	"llm-cache/internal/domain/models"
	"llm-cache/internal/domain/services"
	"llm-cache/pkg/logger"
	"llm-cache/pkg/status"

	"github.com/gin-gonic/gin"
)

// CacheHandler 缓存处理器
// 负责处理缓存相关的HTTP请求，协调CacheService完成业务操作
type CacheHandler struct {
	cacheService services.CacheService // 缓存服务接口
	logger       logger.Logger         // 日志器
}

// NewCacheHandler 创建缓存处理器
func NewCacheHandler(cacheService services.CacheService, log logger.Logger) *CacheHandler {
	return &CacheHandler{
		cacheService: cacheService,
		logger:       log,
	}
}

// APIResponse 统一的API响应格式
type APIResponse struct {
	Success   bool        `json:"success"`              // 是否成功
	Code      int         `json:"code"`                 // 状态码
	Message   string      `json:"message"`              // 消息
	Data      interface{} `json:"data,omitempty"`       // 数据
	RequestID string      `json:"request_id,omitempty"` // 请求ID
	Timestamp int64       `json:"timestamp"`            // 时间戳
}

// ErrorDetail 错误详情
type ErrorDetail struct {
	Field   string `json:"field,omitempty"` // 错误字段
	Message string `json:"message"`         // 错误消息
	Code    string `json:"code,omitempty"`  // 错误码
}

// QueryCache 查询缓存
// GET /v1/cache/search
func (h *CacheHandler) QueryCache(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := middleware.GetRequestID(c)

	h.logger.InfoContext(ctx, "开始处理缓存查询请求", "request_id", requestID)

	// 解析请求参数
	var query models.CacheQuery
	if err := c.ShouldBindJSON(&query); err != nil {
		h.logger.ErrorContext(ctx, "缓存查询请求参数解析失败",
			"request_id", requestID,
			"error", err.Error())

		h.respondWithError(c, status.ErrCodeInvalidParam, "请求参数格式错误", err.Error())
		return
	}

	// 参数验证
	if err := h.validateCacheQuery(&query); err != nil {
		h.logger.ErrorContext(ctx, "缓存查询请求参数验证失败",
			"request_id", requestID,
			"error", err.Error())

		h.respondWithError(c, status.ErrCodeInvalidParam, "请求参数验证失败", err.Error())
		return
	}

	// 调用缓存服务查询
	startTime := time.Now()
	result, err := h.cacheService.QueryCache(ctx, &query)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		h.logger.ErrorContext(ctx, "缓存查询服务调用失败",
			"request_id", requestID,
			"duration_ms", duration,
			"error", err.Error())

		h.respondWithError(c, status.ErrCodeInternal, "缓存查询失败", err.Error())
		return
	}

	h.logger.InfoContext(ctx, "缓存查询请求处理完成",
		"request_id", requestID,
		"duration_ms", duration,
		"found", result.Found)

	// 返回成功响应
	h.respondWithSuccess(c, result, "缓存查询成功")
}

// StoreCache 存储缓存
// POST /v1/cache/store
func (h *CacheHandler) StoreCache(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := middleware.GetRequestID(c)

	h.logger.InfoContext(ctx, "开始处理缓存存储请求", "request_id", requestID)

	// 解析请求参数
	var request models.CacheWriteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.ErrorContext(ctx, "缓存存储请求参数解析失败",
			"request_id", requestID,
			"error", err.Error())

		h.respondWithError(c, status.ErrCodeInvalidParam, "请求参数格式错误", err.Error())
		return
	}

	// 参数验证
	if err := h.validateCacheWriteRequest(&request); err != nil {
		h.logger.ErrorContext(ctx, "缓存存储请求参数验证失败",
			"request_id", requestID,
			"error", err.Error())

		h.respondWithError(c, status.ErrCodeInvalidParam, "请求参数验证失败", err.Error())
		return
	}

	// 调用缓存服务存储
	startTime := time.Now()
	result, err := h.cacheService.StoreCache(ctx, &request)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		h.logger.ErrorContext(ctx, "缓存存储服务调用失败",
			"request_id", requestID,
			"duration_ms", duration,
			"error", err.Error())

		h.respondWithError(c, status.ErrCodeInternal, "缓存存储失败", err.Error())
		return
	}

	h.logger.InfoContext(ctx, "缓存存储请求处理完成",
		"request_id", requestID,
		"duration_ms", duration,
		"success", result.Success,
		"cache_id", result.CacheID)

	// 返回成功响应
	h.respondWithSuccess(c, result, "缓存存储成功")
}

// DeleteCache 删除单个缓存
// DELETE /v1/cache/:cache_id
func (h *CacheHandler) DeleteCache(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := middleware.GetRequestID(c)

	h.logger.InfoContext(ctx, "开始处理缓存删除请求", "request_id", requestID)

	// 获取缓存ID
	cacheID := c.Param("cache_id")
	if cacheID == "" {
		h.logger.ErrorContext(ctx, "缓存删除请求缺少cache_id参数", "request_id", requestID)
		h.respondWithError(c, status.ErrCodeInvalidParam, "缺少cache_id参数", "")
		return
	}

	// 获取用户类型
	userType := c.Query("user_type")
	if userType == "" {
		h.logger.ErrorContext(ctx, "缓存删除请求缺少user_type参数", "request_id", requestID)
		h.respondWithError(c, status.ErrCodeInvalidParam, "缺少user_type参数", "")
		return
	}

	// 构建删除请求
	deleteRequest := &models.CacheDeleteRequest{
		CacheIDs: []string{cacheID},
		UserType: userType,
		Force:    c.Query("force") == "true",
	}

	// 调用缓存服务删除
	startTime := time.Now()
	result, err := h.cacheService.DeleteCache(ctx, deleteRequest)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		h.logger.ErrorContext(ctx, "缓存删除服务调用失败",
			"request_id", requestID,
			"cache_id", cacheID,
			"duration_ms", duration,
			"error", err.Error())

		h.respondWithError(c, status.ErrCodeInternal, "缓存删除失败", err.Error())
		return
	}

	h.logger.InfoContext(ctx, "缓存删除请求处理完成",
		"request_id", requestID,
		"cache_id", cacheID,
		"duration_ms", duration,
		"success", result.Success)

	// 返回成功响应
	h.respondWithSuccess(c, result, "缓存删除成功")
}

// BatchDeleteCache 批量删除缓存
// DELETE /v1/cache/batch
func (h *CacheHandler) BatchDeleteCache(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := middleware.GetRequestID(c)

	h.logger.InfoContext(ctx, "开始处理批量缓存删除请求", "request_id", requestID)

	// 解析请求参数
	var request models.CacheDeleteRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.logger.ErrorContext(ctx, "批量缓存删除请求参数解析失败",
			"request_id", requestID,
			"error", err.Error())

		h.respondWithError(c, status.ErrCodeInvalidParam, "请求参数格式错误", err.Error())
		return
	}

	// 参数验证
	if len(request.CacheIDs) == 0 {
		h.logger.ErrorContext(ctx, "批量缓存删除请求缺少cache_ids", "request_id", requestID)
		h.respondWithError(c, status.ErrCodeInvalidParam, "缺少要删除的缓存ID", "")
		return
	}

	if request.UserType == "" {
		h.logger.ErrorContext(ctx, "批量缓存删除请求缺少user_type", "request_id", requestID)
		h.respondWithError(c, status.ErrCodeInvalidParam, "缺少user_type参数", "")
		return
	}

	// 调用缓存服务删除
	startTime := time.Now()
	result, err := h.cacheService.DeleteCache(ctx, &request)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		h.logger.ErrorContext(ctx, "批量缓存删除服务调用失败",
			"request_id", requestID,
			"cache_count", len(request.CacheIDs),
			"duration_ms", duration,
			"error", err.Error())

		h.respondWithError(c, status.ErrCodeInternal, "批量缓存删除失败", err.Error())
		return
	}

	h.logger.InfoContext(ctx, "批量缓存删除请求处理完成",
		"request_id", requestID,
		"cache_count", len(request.CacheIDs),
		"deleted_count", result.DeletedCount,
		"duration_ms", duration,
		"success", result.Success)

	// 返回成功响应
	h.respondWithSuccess(c, result, "批量缓存删除成功")
}

// GetCacheByID 根据ID获取缓存项
// GET /v1/cache/:cache_id
func (h *CacheHandler) GetCacheByID(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := middleware.GetRequestID(c)

	h.logger.InfoContext(ctx, "开始处理缓存查询请求", "request_id", requestID)

	// 获取缓存ID
	cacheID := c.Param("cache_id")
	if cacheID == "" {
		h.logger.ErrorContext(ctx, "缓存查询请求缺少cache_id参数", "request_id", requestID)
		h.respondWithError(c, status.ErrCodeInvalidParam, "缺少cache_id参数", "")
		return
	}

	// 获取用户类型
	userType := c.Query("user_type")
	if userType == "" {
		h.logger.ErrorContext(ctx, "缓存查询请求缺少user_type参数", "request_id", requestID)
		h.respondWithError(c, status.ErrCodeInvalidParam, "缺少user_type参数", "")
		return
	}

	// 解析是否包含统计信息
	// 调用缓存服务查询
	startTime := time.Now()
	cacheItem, err := h.cacheService.GetCacheByID(ctx, cacheID, userType)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		h.logger.ErrorContext(ctx, "缓存查询服务调用失败",
			"request_id", requestID,
			"cache_id", cacheID,
			"duration_ms", duration,
			"error", err.Error())

		// 检查是否为资源不存在错误
		if strings.Contains(err.Error(), "not found") {
			h.respondWithError(c, status.ErrCodeNotFound, "缓存项不存在", err.Error())
		} else {
			h.respondWithError(c, status.ErrCodeInternal, "缓存查询失败", err.Error())
		}
		return
	}

	h.logger.InfoContext(ctx, "缓存查询请求处理完成",
		"request_id", requestID,
		"cache_id", cacheID,
		"duration_ms", duration)

	// 返回成功响应
	h.respondWithSuccess(c, cacheItem, "缓存查询成功")
}

// GetCacheStatistics 获取缓存统计信息
// GET /v1/cache/statistics
func (h *CacheHandler) GetCacheStatistics(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := middleware.GetRequestID(c)

	h.logger.InfoContext(ctx, "开始处理缓存统计查询请求", "request_id", requestID)

	// 获取查询参数
	userType := c.Query("user_type")
	timeRange := c.Query("time_range")
	if timeRange == "" {
		timeRange = "24h" // 默认24小时
	}

	// 调用缓存服务查询统计信息
	startTime := time.Now()
	statistics, err := h.cacheService.GetCacheStatistics(ctx)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		h.logger.ErrorContext(ctx, "缓存统计查询服务调用失败",
			"request_id", requestID,
			"user_type", userType,
			"time_range", timeRange,
			"duration_ms", duration,
			"error", err.Error())

		h.respondWithError(c, status.ErrCodeInternal, "缓存统计查询失败", err.Error())
		return
	}

	h.logger.InfoContext(ctx, "缓存统计查询请求处理完成",
		"request_id", requestID,
		"user_type", userType,
		"time_range", timeRange,
		"duration_ms", duration)

	// 返回成功响应
	h.respondWithSuccess(c, statistics, "缓存统计查询成功")
}

// HealthCheck 健康检查
// GET /v1/cache/health
func (h *CacheHandler) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := middleware.GetRequestID(c)

	h.logger.InfoContext(ctx, "开始处理健康检查请求", "request_id", requestID)

	// 调用缓存服务健康检查
	startTime := time.Now()
	healthInfo, err := h.cacheService.GetCacheHealth(ctx)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		h.logger.ErrorContext(ctx, "健康检查服务调用失败",
			"request_id", requestID,
			"duration_ms", duration,
			"error", err.Error())

		h.respondWithError(c, status.ErrCodeUnavailable, "服务不可用", err.Error())
		return
	}

	h.logger.InfoContext(ctx, "健康检查请求处理完成",
		"request_id", requestID,
		"duration_ms", duration)

	// 返回成功响应
	h.respondWithSuccess(c, healthInfo, "服务正常")
}

// 私有方法：参数验证

// validateCacheQuery 验证缓存查询请求
func (h *CacheHandler) validateCacheQuery(query *models.CacheQuery) error {
	if strings.TrimSpace(query.Question) == "" {
		return &ValidationError{Field: "question", Message: "问题不能为空"}
	}

	if strings.TrimSpace(query.UserType) == "" {
		return &ValidationError{Field: "user_type", Message: "用户类型不能为空"}
	}

	if query.SimilarityThreshold != 0 && (query.SimilarityThreshold < 0 || query.SimilarityThreshold > 1) {
		return &ValidationError{Field: "similarity_threshold", Message: "相似度阈值必须在0-1之间"}
	}

	if query.TopK != 0 && (query.TopK < 1 || query.TopK > 100) {
		return &ValidationError{Field: "top_k", Message: "TopK值必须在1-100之间"}
	}

	return nil
}

// validateCacheWriteRequest 验证缓存写入请求
func (h *CacheHandler) validateCacheWriteRequest(request *models.CacheWriteRequest) error {
	if strings.TrimSpace(request.Question) == "" {
		return &ValidationError{Field: "question", Message: "问题不能为空"}
	}

	if len(request.Question) > 1000 {
		return &ValidationError{Field: "question", Message: "问题长度不能超过1000字符"}
	}

	if strings.TrimSpace(request.Answer) == "" {
		return &ValidationError{Field: "answer", Message: "答案不能为空"}
	}

	if len(request.Answer) > 10000 {
		return &ValidationError{Field: "answer", Message: "答案长度不能超过10000字符"}
	}

	if strings.TrimSpace(request.UserType) == "" {
		return &ValidationError{Field: "user_type", Message: "用户类型不能为空"}
	}

	return nil
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// 私有方法：响应处理

// respondWithSuccess 返回成功响应
func (h *CacheHandler) respondWithSuccess(c *gin.Context, data interface{}, message string) {
	response := APIResponse{
		Success:   true,
		Code:      int(status.CodeOK),
		Message:   message,
		Data:      data,
		RequestID: middleware.GetRequestID(c),
		Timestamp: time.Now().Unix(),
	}

	c.JSON(http.StatusOK, response)
}

// respondWithError 返回错误响应
func (h *CacheHandler) respondWithError(c *gin.Context, code status.StatusCode, message, detail string) {
	response := APIResponse{
		Success:   false,
		Code:      int(code),
		Message:   message,
		RequestID: middleware.GetRequestID(c),
		Timestamp: time.Now().Unix(),
	}

	// 如果有详细错误信息，添加到data中
	if detail != "" {
		response.Data = ErrorDetail{
			Message: detail,
			Code:    code.String(),
		}
	}

	// 返回200的HTTP状态码
	c.JSON(http.StatusOK, response)
}
