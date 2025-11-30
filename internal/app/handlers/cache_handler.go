package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego/eino/compose"
	"github.com/gin-gonic/gin"

	"llm-cache/internal/app/middleware"
	"llm-cache/internal/eino/flows"
	"llm-cache/pkg/logger"
	"llm-cache/pkg/status"
)

// CacheHandler 缓存处理器
// 使用 Eino compose.Runnable 类型替代自定义 Service 接口
type CacheHandler struct {
	queryRunner   compose.Runnable[*flows.CacheQueryInput, *flows.CacheQueryOutput]
	storeRunner   compose.Runnable[*flows.CacheStoreInput, *flows.CacheStoreOutput]
	deleteService *flows.CacheDeleteService
	logger        logger.Logger
}

// NewCacheHandler 创建缓存处理器
func NewCacheHandler(
	queryRunner compose.Runnable[*flows.CacheQueryInput, *flows.CacheQueryOutput],
	storeRunner compose.Runnable[*flows.CacheStoreInput, *flows.CacheStoreOutput],
	deleteService *flows.CacheDeleteService,
	log logger.Logger,
) *CacheHandler {
	return &CacheHandler{
		queryRunner:   queryRunner,
		storeRunner:   storeRunner,
		deleteService: deleteService,
		logger:        log,
	}
}

// APIResponse 统一的API响应格式
type APIResponse struct {
	Success   bool        `json:"success"`
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// ErrorDetail 错误详情
type ErrorDetail struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// QueryRequest 查询请求
type QueryRequest struct {
	Question            string  `json:"question" binding:"required"`
	UserType            string  `json:"user_type" binding:"required"`
	TopK                int     `json:"top_k,omitempty"`
	SimilarityThreshold float64 `json:"similarity_threshold,omitempty"`
}

// StoreRequest 存储请求
type StoreRequest struct {
	Question   string         `json:"question" binding:"required"`
	Answer     string         `json:"answer" binding:"required"`
	UserType   string         `json:"user_type" binding:"required"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	ForceWrite bool           `json:"force_write,omitempty"`
}

// DeleteRequest 删除请求
type DeleteRequest struct {
	CacheIDs []string `json:"cache_ids" binding:"required"`
	UserType string   `json:"user_type" binding:"required"`
	Force    bool     `json:"force,omitempty"`
}

// QueryCache 查询缓存
// POST /v1/cache/search
func (h *CacheHandler) QueryCache(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := middleware.GetRequestID(c)

	h.logger.InfoContext(ctx, "开始处理缓存查询请求", "request_id", requestID)

	// 解析请求参数
	var req QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.ErrorContext(ctx, "缓存查询请求参数解析失败",
			"request_id", requestID,
			"error", err.Error())
		h.respondWithError(c, status.ErrCodeInvalidParam, "请求参数格式错误", err.Error())
		return
	}

	// 参数验证
	if err := h.validateQueryRequest(&req); err != nil {
		h.logger.ErrorContext(ctx, "缓存查询请求参数验证失败",
			"request_id", requestID,
			"error", err.Error())
		h.respondWithError(c, status.ErrCodeInvalidParam, "请求参数验证失败", err.Error())
		return
	}

	// 构建 Graph 输入
	input := &flows.CacheQueryInput{
		Query:          req.Question,
		UserType:       req.UserType,
		TopK:           req.TopK,
		ScoreThreshold: req.SimilarityThreshold,
	}

	// 调用 Eino Runnable
	startTime := time.Now()
	result, err := h.queryRunner.Invoke(ctx, input)
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
		"found", result.Hit)

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
	var req StoreRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.ErrorContext(ctx, "缓存存储请求参数解析失败",
			"request_id", requestID,
			"error", err.Error())
		h.respondWithError(c, status.ErrCodeInvalidParam, "请求参数格式错误", err.Error())
		return
	}

	// 参数验证
	if err := h.validateStoreRequest(&req); err != nil {
		h.logger.ErrorContext(ctx, "缓存存储请求参数验证失败",
			"request_id", requestID,
			"error", err.Error())
		h.respondWithError(c, status.ErrCodeInvalidParam, "请求参数验证失败", err.Error())
		return
	}

	// 构建 Graph 输入
	input := &flows.CacheStoreInput{
		Question:   req.Question,
		Answer:     req.Answer,
		UserType:   req.UserType,
		Metadata:   req.Metadata,
		ForceWrite: req.ForceWrite,
	}

	// 调用 Eino Runnable
	startTime := time.Now()
	result, err := h.storeRunner.Invoke(ctx, input)
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

	// 调用删除服务
	startTime := time.Now()
	input := &flows.CacheDeleteInput{
		CacheIDs: []string{cacheID},
		UserType: userType,
		Force:    c.Query("force") == "true",
	}

	result, err := h.deleteService.Delete(ctx, input)
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
	var req DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.ErrorContext(ctx, "批量缓存删除请求参数解析失败",
			"request_id", requestID,
			"error", err.Error())
		h.respondWithError(c, status.ErrCodeInvalidParam, "请求参数格式错误", err.Error())
		return
	}

	// 参数验证
	if len(req.CacheIDs) == 0 {
		h.logger.ErrorContext(ctx, "批量缓存删除请求缺少cache_ids", "request_id", requestID)
		h.respondWithError(c, status.ErrCodeInvalidParam, "缺少要删除的缓存ID", "")
		return
	}

	if req.UserType == "" {
		h.logger.ErrorContext(ctx, "批量缓存删除请求缺少user_type", "request_id", requestID)
		h.respondWithError(c, status.ErrCodeInvalidParam, "缺少user_type参数", "")
		return
	}

	// 调用删除服务
	startTime := time.Now()
	input := &flows.CacheDeleteInput{
		CacheIDs: req.CacheIDs,
		UserType: req.UserType,
		Force:    req.Force,
	}

	result, err := h.deleteService.Delete(ctx, input)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		h.logger.ErrorContext(ctx, "批量缓存删除服务调用失败",
			"request_id", requestID,
			"cache_count", len(req.CacheIDs),
			"duration_ms", duration,
			"error", err.Error())
		h.respondWithError(c, status.ErrCodeInternal, "批量缓存删除失败", err.Error())
		return
	}

	h.logger.InfoContext(ctx, "批量缓存删除请求处理完成",
		"request_id", requestID,
		"cache_count", len(req.CacheIDs),
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

	// 调用获取服务
	startTime := time.Now()
	cacheItem, err := h.deleteService.GetByID(ctx, cacheID)
	duration := time.Since(startTime).Milliseconds()

	if err != nil {
		h.logger.ErrorContext(ctx, "缓存查询服务调用失败",
			"request_id", requestID,
			"cache_id", cacheID,
			"duration_ms", duration,
			"error", err.Error())

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

	// 统计功能暂未在 Eino 中实现，返回基本信息
	statistics := map[string]interface{}{
		"status": "running",
		"time":   time.Now().Unix(),
	}

	h.logger.InfoContext(ctx, "缓存统计查询请求处理完成", "request_id", requestID)

	h.respondWithSuccess(c, statistics, "缓存统计查询成功")
}

// HealthCheck 健康检查
// GET /v1/cache/health
func (h *CacheHandler) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()
	requestID := middleware.GetRequestID(c)

	h.logger.InfoContext(ctx, "开始处理健康检查请求", "request_id", requestID)

	healthInfo := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Unix(),
	}

	h.logger.InfoContext(ctx, "健康检查请求处理完成", "request_id", requestID)

	h.respondWithSuccess(c, healthInfo, "服务正常")
}

// 私有方法：参数验证

// validateQueryRequest 验证查询请求
func (h *CacheHandler) validateQueryRequest(req *QueryRequest) error {
	if strings.TrimSpace(req.Question) == "" {
		return &ValidationError{Field: "question", Message: "问题不能为空"}
	}

	if strings.TrimSpace(req.UserType) == "" {
		return &ValidationError{Field: "user_type", Message: "用户类型不能为空"}
	}

	if req.SimilarityThreshold != 0 && (req.SimilarityThreshold < 0 || req.SimilarityThreshold > 1) {
		return &ValidationError{Field: "similarity_threshold", Message: "相似度阈值必须在0-1之间"}
	}

	if req.TopK != 0 && (req.TopK < 1 || req.TopK > 100) {
		return &ValidationError{Field: "top_k", Message: "TopK值必须在1-100之间"}
	}

	return nil
}

// validateStoreRequest 验证存储请求
func (h *CacheHandler) validateStoreRequest(req *StoreRequest) error {
	if strings.TrimSpace(req.Question) == "" {
		return &ValidationError{Field: "question", Message: "问题不能为空"}
	}

	if len(req.Question) > 1000 {
		return &ValidationError{Field: "question", Message: "问题长度不能超过1000字符"}
	}

	if strings.TrimSpace(req.Answer) == "" {
		return &ValidationError{Field: "answer", Message: "答案不能为空"}
	}

	if len(req.Answer) > 10000 {
		return &ValidationError{Field: "answer", Message: "答案长度不能超过10000字符"}
	}

	if strings.TrimSpace(req.UserType) == "" {
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

	if detail != "" {
		response.Data = ErrorDetail{
			Message: detail,
			Code:    code.String(),
		}
	}

	c.JSON(http.StatusOK, response)
}
