package preprocessing

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"llm-cache/internal/domain/models"
	"llm-cache/internal/domain/services"
	"llm-cache/pkg/logger"
	"llm-cache/pkg/status"
)

// DefaultRequestPreprocessingService 默认请求预处理服务实现
type DefaultRequestPreprocessingService struct {
	config         *Config
	preprocessors  map[string]services.PreprocessorFunc
	processorOrder []string
	mutex          sync.RWMutex
	logger         logger.Logger
}

// NewDefaultRequestPreprocessingService 创建默认请求预处理服务
func NewDefaultRequestPreprocessingService(config *Config, log logger.Logger) services.RequestPreprocessingService {
	if config == nil {
		config = DefaultConfig()
	}

	service := &DefaultRequestPreprocessingService{
		config:         config,
		preprocessors:  make(map[string]services.PreprocessorFunc),
		processorOrder: make([]string, 0),
		logger:         log,
	}

	return service
}

// PreprocessQuery 预处理查询请求
func (s *DefaultRequestPreprocessingService) PreprocessQuery(ctx context.Context, request *models.CacheQuery) (*models.PreprocessedRequest, status.StatusCode, error) {
	startTime := time.Now()

	// 应用注册的预处理函数链
	processedText, err := s.applyPreprocessors(ctx, request.Question, request.UserType, make(map[string]interface{}))
	if err != nil {
		s.logger.Error("预处理函数执行失败", "error", err)
		return nil, status.ErrCodeInternal, fmt.Errorf("预处理函数执行失败: %w", err)
	}

	// 创建预处理结果
	result := &models.PreprocessedRequest{
		Original:          request,
		ProcessedQuestion: processedText,
		ProcessingTime:    float64(time.Since(startTime).Milliseconds()),
		Success:           true,
	}

	s.logger.Info("请求预处理完成",
		"user_type", request.UserType,
		"original_length", len(request.Question),
		"processed_length", len(processedText),
		"applied_preprocessors", len(s.processorOrder),
		"processing_time_ms", result.ProcessingTime,
		"original_question", request.Question,
		"processed_question", processedText)

	return result, status.CodeOK, nil
}

// RegisterPreprocessor 注册预处理函数
func (s *DefaultRequestPreprocessingService) RegisterPreprocessor(name string, processor services.PreprocessorFunc) error {
	if name == "" {
		return fmt.Errorf("预处理函数名称不能为空")
	}

	if processor == nil {
		return fmt.Errorf("预处理函数不能为空")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 检查是否已存在
	if _, exists := s.preprocessors[name]; exists {
		return fmt.Errorf("预处理函数 '%s' 已存在", name)
	}

	// 注册函数
	s.preprocessors[name] = processor
	s.processorOrder = append(s.processorOrder, name)

	s.logger.Info("预处理函数已注册", "name", name, "total_count", len(s.preprocessors))

	return nil
}

// UnregisterPreprocessor 取消注册预处理函数
func (s *DefaultRequestPreprocessingService) UnregisterPreprocessor(name string) error {
	if name == "" {
		return fmt.Errorf("预处理函数名称不能为空")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 检查是否存在
	if _, exists := s.preprocessors[name]; !exists {
		return fmt.Errorf("预处理函数 '%s' 不存在", name)
	}

	// 删除函数
	delete(s.preprocessors, name)

	// 从顺序列表中移除
	for i, orderName := range s.processorOrder {
		if orderName == name {
			s.processorOrder = append(s.processorOrder[:i], s.processorOrder[i+1:]...)
			break
		}
	}

	s.logger.Info("预处理函数已取消注册", "name", name, "remaining_count", len(s.preprocessors))

	return nil
}

// ListPreprocessors 列出所有已注册的预处理函数名称
func (s *DefaultRequestPreprocessingService) ListPreprocessors() []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 返回副本，避免外部修改
	result := make([]string, len(s.processorOrder))
	copy(result, s.processorOrder)

	return result
}

// applyPreprocessors 应用预处理函数链
func (s *DefaultRequestPreprocessingService) applyPreprocessors(ctx context.Context, text string, userType string, metadata map[string]interface{}) (string, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 如果没有注册任何预处理函数，直接返回原文本
	if len(s.preprocessors) == 0 {
		return text, nil
	}

	// 按注册顺序依次应用预处理函数
	processedText := text
	for _, name := range s.processorOrder {
		processor, exists := s.preprocessors[name]
		if !exists {
			continue // 跳过已删除的函数
		}

		// 应用预处理函数
		// 注意：PreprocessorFunc 签名为 func(text string, metadata map[string]interface{}) string
		processedText = processor(processedText, metadata)

		s.logger.Debug("预处理函数已应用", "name", name, "before_length", len(text), "after_length", len(processedText))
	}

	return processedText, nil
}

// getAppliedPreprocessors 获取已应用的预处理函数列表
func (s *DefaultRequestPreprocessingService) getAppliedPreprocessors() []string {
	// 返回当前注册的预处理函数顺序
	return s.ListPreprocessors()
}

// generateRequestID 生成请求ID
func (s *DefaultRequestPreprocessingService) generateRequestID(ctx context.Context) string {
	// 从context中获取现有的请求ID

	return ctx.Value("request_id").(string)
}

// basicTextCleaning 基础文本清理（当没有自定义预处理函数时使用）
func (s *DefaultRequestPreprocessingService) basicTextCleaning(text string) string {
	// 移除多余空白字符
	processed := strings.TrimSpace(text)

	// 标准化连续空白字符
	for strings.Contains(processed, "  ") {
		processed = strings.ReplaceAll(processed, "  ", " ")
	}

	return processed
}
