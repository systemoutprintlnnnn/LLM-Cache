package vector

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"

	"llm-cache/internal/domain/models"
	"llm-cache/internal/domain/repositories"
	"llm-cache/internal/domain/services"
	"llm-cache/pkg/logger"
)

// DefaultVectorService 默认向量服务实现
// 组合EmbeddingService和VectorRepository完成语义搜索，提供文本相似度计算能力
type DefaultVectorService struct {
	embeddingService services.EmbeddingService     // 嵌入服务，负责文本向量化
	vectorRepository repositories.VectorRepository // 向量仓储，负责向量存储和搜索
	strategyFactory  *StrategyFactory              // 选择策略工厂
	logger           logger.Logger                 // 日志器
	config           *VectorServiceConfig          // 配置
}

// VectorServiceConfig 向量服务配置
type VectorServiceConfig struct {
	// DefaultCollectionName 默认集合名称
	DefaultCollectionName string `json:"default_collection_name" yaml:"default_collection_name"`

	// DefaultTopK 默认返回结果数量
	DefaultTopK int `json:"default_top_k" yaml:"default_top_k"`

	// DefaultSimilarityThreshold 默认相似度阈值
	DefaultSimilarityThreshold float64 `json:"default_similarity_threshold" yaml:"default_similarity_threshold"`

	// MaxBatchSize 最大批量处理大小
	MaxBatchSize int `json:"max_batch_size" yaml:"max_batch_size"`

	// RequestTimeout 请求超时时间（秒）
	RequestTimeout int `json:"request_timeout" yaml:"request_timeout"`

	// EnableNormalization 是否启用向量归一化
	EnableNormalization bool `json:"enable_normalization" yaml:"enable_normalization"`

	// DefaultSelectionStrategy 默认选择策略
	DefaultSelectionStrategy string `json:"default_selection_strategy" yaml:"default_selection_strategy"`

	// TemperatureSoftmaxConfig 温度softmax策略配置
	TemperatureSoftmaxConfig *TemperatureSoftmaxConfig `json:"temperature_softmax" yaml:"temperature_softmax"`
}

// DefaultVectorServiceConfig 默认向量服务配置
func DefaultVectorServiceConfig() *VectorServiceConfig {
	return &VectorServiceConfig{
		DefaultCollectionName:      "qa_cache",
		DefaultTopK:                5,
		DefaultSimilarityThreshold: 0.7,
		MaxBatchSize:               100,
		RequestTimeout:             30,
		EnableNormalization:        true,
		DefaultSelectionStrategy:   "highest_score",
		TemperatureSoftmaxConfig:   DefaultTemperatureSoftmaxConfig(),
	}
}

// NewDefaultVectorService 创建默认向量服务实例
func NewDefaultVectorService(
	embeddingService services.EmbeddingService,
	vectorRepository repositories.VectorRepository,
	config *VectorServiceConfig,
	log logger.Logger,
) services.VectorService {
	if config == nil {
		config = DefaultVectorServiceConfig()
	}

	// 创建选择策略工厂
	strategyFactory := NewStrategyFactory(log, config.TemperatureSoftmaxConfig)

	return &DefaultVectorService{
		embeddingService: embeddingService,
		vectorRepository: vectorRepository,
		strategyFactory:  strategyFactory,
		logger:           log,
		config:           config,
	}
}

// SearchCache 搜索语义缓存
// 完整的语义缓存查询流程：文本向量化 + 相似度搜索
func (s *DefaultVectorService) SearchCache(ctx context.Context, query *models.CacheQuery) (*models.CacheResult, error) {
	startTime := time.Now()

	s.logger.InfoContext(ctx, "开始搜索语义缓存",
		"question", query.Question,
		"user_type", query.UserType,
		"similarity_threshold", query.SimilarityThreshold,
		"top_k", query.TopK,
	)

	// 1. 文本向量化
	vectorRequest := &models.VectorProcessingRequest{
		Text:      query.Question,
		Normalize: s.config.EnableNormalization,
	}

	vectorResult, err := s.embeddingService.GenerateEmbedding(ctx, vectorRequest)
	if err != nil {
		s.logger.ErrorContext(ctx, "文本向量化失败", "error", err)
		return &models.CacheResult{
			Found:        false,
			ResponseTime: float64(time.Since(startTime).Nanoseconds()) / 1e6,
		}, fmt.Errorf("文本向量化失败: %w", err)
	}

	if !vectorResult.Success {
		s.logger.ErrorContext(ctx, "向量化处理失败", "error", vectorResult.Error)
		return &models.CacheResult{
			Found:        false,
			ResponseTime: float64(time.Since(startTime).Nanoseconds()) / 1e6,
		}, fmt.Errorf("向量化处理失败: %s", vectorResult.Error)
	}

	// 2. 构建向量搜索请求
	topK, similarityThreshold := func() (int, float64) {
		t := s.config.DefaultTopK
		st := s.config.DefaultSimilarityThreshold
		if query.TopK > 0 {
			t = query.TopK
		}
		if query.SimilarityThreshold >= 0 {
			st = query.SimilarityThreshold
		}
		return t, st
	}()

	searchRequest := &models.VectorSearchRequest{
		QueryText:           query.Question,
		QueryVector:         vectorResult.Vector.Values,
		TopK:                topK,
		SimilarityThreshold: similarityThreshold,
		UserType:            query.UserType,
		Filters:             query.Filters,
	}

	// 3. 执行向量相似度搜索
	s.logger.InfoContext(ctx, "开始执行向量搜索",
		"vector_dimension", len(vectorResult.Vector.Values),
		"top_k", topK,
		"similarity_threshold", similarityThreshold,
	)

	searchResponse, err := s.vectorRepository.Search(ctx, searchRequest)
	if err != nil {
		s.logger.ErrorContext(ctx, "向量搜索失败", "error", err)
		return &models.CacheResult{
			Found:        false,
			ResponseTime: float64(time.Since(startTime).Nanoseconds()) / 1e6,
		}, fmt.Errorf("向量搜索失败: %w", err)
	}

	// 4. 处理搜索结果
	responseTime := float64(time.Since(startTime).Nanoseconds()) / 1e6

	if len(searchResponse.Results) == 0 {
		s.logger.InfoContext(ctx, "未找到匹配的缓存", "response_time", responseTime)
		return &models.CacheResult{
			Found:        false,
			ResponseTime: responseTime,
		}, nil
	}

	// 5. 使用选择策略选择最佳结果
	strategy := s.config.DefaultSelectionStrategy

	// 转换结果类型
	resultPointers := make([]*models.VectorSearchResult, len(searchResponse.Results))
	for i := range searchResponse.Results {
		resultPointers[i] = &searchResponse.Results[i]
	}

	bestResult, err := s.SelectBestResult(ctx, resultPointers, query, strategy)
	if err != nil {
		s.logger.ErrorContext(ctx, "选择最优结果失败", "error", err)
		// 回退到第一个结果
		bestResult = &searchResponse.Results[0]
	}

	// 从payload中提取答案
	answer := ""
	if answerValue, exists := bestResult.Payload["answer"]; exists {
		if answerStr, ok := answerValue.(string); ok {
			answer = answerStr
		}
	}

	// 从payload中提取元数据
	var metadata *models.CacheMetadata
	if metadataValue, exists := bestResult.Payload["metadata"]; exists {
		if metadataMap, ok := metadataValue.(map[string]interface{}); ok {
			metadata = s.extractMetadataFromPayload(metadataMap)
		}
	}

	// 从payload中提取统计信息
	var statistics *models.CacheStatistics
	// 默认总是提取统计信息
	if statsValue, exists := bestResult.Payload["statistics"]; exists {
		if statsMap, ok := statsValue.(map[string]interface{}); ok {
			statistics = s.extractStatisticsFromPayload(statsMap)
		}
	}

	result := &models.CacheResult{
		Found:        true,
		CacheID:      bestResult.ID,
		Answer:       answer,
		Similarity:   bestResult.Score,
		ResponseTime: responseTime,
		Metadata:     metadata,
		Statistics:   statistics,
	}

	s.logger.InfoContext(ctx, "成功找到缓存匹配",
		"cache_id", result.CacheID,
		"similarity", result.Similarity,
		"response_time", responseTime,
		"selection_strategy", strategy,
	)

	return result, nil
}

// StoreCache 存储查询和响应到缓存
// 将用户查询和LLM响应存储到向量缓存中
func (s *DefaultVectorService) StoreCache(ctx context.Context, request *models.CacheWriteRequest) (*models.CacheWriteResult, error) {
	startTime := time.Now()

	s.logger.InfoContext(ctx, "开始存储缓存",
		"question_length", len(request.Question),
		"answer_length", len(request.Answer),
		"user_type", request.UserType,
		"force_write", request.ForceWrite,
	)

	// 1. 文本向量化
	vectorRequest := &models.VectorProcessingRequest{
		Text:      request.Question,
		Normalize: s.config.EnableNormalization,
	}

	vectorResult, err := s.embeddingService.GenerateEmbedding(ctx, vectorRequest)
	if err != nil {
		s.logger.ErrorContext(ctx, "文本向量化失败", "error", err)
		return &models.CacheWriteResult{
			Success: false,
			Message: fmt.Sprintf("文本向量化失败: 向量化错误: %v", err),
		}, fmt.Errorf("文本向量化失败: %w", err)
	}

	if !vectorResult.Success {
		s.logger.ErrorContext(ctx, "向量化处理失败", "error", vectorResult.Error)
		return &models.CacheWriteResult{
			Success: false,
			Message: fmt.Sprintf("向量化处理失败: %s", vectorResult.Error),
		}, fmt.Errorf("向量化处理失败: %s", vectorResult.Error)
	}

	// 2. 生成缓存ID
	cacheID := s.generateCacheID()

	// 3. 构建向量存储请求
	payload := map[string]interface{}{
		"question":  request.Question,
		"answer":    request.Answer,
		"user_type": request.UserType,
		"timestamp": time.Now().Unix(),
	}

	// 添加元数据
	if request.Metadata != nil {
		payload["metadata"] = map[string]interface{}{
			"source":        request.Metadata.Source,
			"quality_score": request.Metadata.QualityScore,
			"version":       request.Metadata.Version,
		}
	}

	// 添加初始统计信息
	payload["statistics"] = map[string]interface{}{
		"hit_count":     0,
		"like_count":    0,
		"dislike_count": 0,
		"response_time": 0.0,
	}

	storeRequest := &models.VectorStoreRequest{
		ID:             cacheID,
		Vector:         vectorResult.Vector.Values,
		CollectionName: s.config.DefaultCollectionName,
		Payload:        payload,
		UpsertMode:     true, // 允许更新已存在的记录
	}

	// 4. 执行存储操作
	storeResponse, err := s.vectorRepository.Store(ctx, storeRequest)
	if err != nil {
		s.logger.ErrorContext(ctx, "向量存储失败", "error", err)
		return &models.CacheWriteResult{
			Success: false,
			Message: fmt.Sprintf("向量存储失败: 存储错误: %v", err),
		}, fmt.Errorf("向量存储失败: %w", err)
	}

	if !storeResponse.Success {
		s.logger.ErrorContext(ctx, "向量存储操作失败", "message", storeResponse.Message)
		return &models.CacheWriteResult{
			Success: false,
			Message: fmt.Sprintf("向量存储操作失败: %s", storeResponse.Message),
		}, nil
	}

	responseTime := float64(time.Since(startTime).Nanoseconds()) / 1e6

	result := &models.CacheWriteResult{
		Success: true,
		CacheID: cacheID,
		Message: "缓存存储成功",
	}

	// 设置质量分数（如果有元数据）
	if request.Metadata != nil {
		result.QualityScore = request.Metadata.QualityScore
	}

	s.logger.InfoContext(ctx, "成功存储缓存",
		"cache_id", cacheID,
		"vector_id", storeResponse.VectorID,
		"store_time", storeResponse.StoreTime,
		"response_time", responseTime,
	)

	return result, nil
}

// DeleteCache 删除缓存项
// 从向量缓存中删除指定的缓存项
func (s *DefaultVectorService) DeleteCache(ctx context.Context, request *models.CacheDeleteRequest) (*models.CacheDeleteResult, error) {
	startTime := time.Now()

	s.logger.InfoContext(ctx, "开始删除缓存",
		"cache_ids", request.CacheIDs,
		"user_type", request.UserType,
		"force", request.Force,
	)

	if len(request.CacheIDs) == 0 {
		return &models.CacheDeleteResult{
			Success:      false,
			DeletedCount: 0,
			Message:      "未提供要删除的缓存ID",
		}, nil
	}

	// 执行删除操作
	deleteResult, err := s.vectorRepository.Delete(ctx, request.CacheIDs, request.UserType)
	if err != nil {
		s.logger.ErrorContext(ctx, "向量删除失败", "error", err)
		return &models.CacheDeleteResult{
			Success:      false,
			DeletedCount: 0,
			FailedIDs:    request.CacheIDs,
			Message:      fmt.Sprintf("删除操作失败: %v", err),
		}, fmt.Errorf("向量删除失败: %w", err)
	}

	responseTime := float64(time.Since(startTime).Nanoseconds()) / 1e6

	s.logger.InfoContext(ctx, "成功删除缓存",
		"deleted_count", deleteResult.DeletedCount,
		"response_time", responseTime,
	)

	return deleteResult, nil
}

// generateCacheID 生成缓存ID
func (s *DefaultVectorService) generateCacheID() string {
	// 暂时直接使用UUID
	return uuid.New().String()
}

// calculateCosineSimilarity 计算余弦相似度
func (s *DefaultVectorService) calculateCosineSimilarity(vector1, vector2 []float32) (float64, error) {
	if len(vector1) != len(vector2) {
		return 0.0, fmt.Errorf("向量维度不匹配: %d vs %d", len(vector1), len(vector2))
	}

	if len(vector1) == 0 {
		return 0.0, fmt.Errorf("向量不能为空")
	}

	var dotProduct, norm1, norm2 float64

	for i := 0; i < len(vector1); i++ {
		dotProduct += float64(vector1[i]) * float64(vector2[i])
		norm1 += float64(vector1[i]) * float64(vector1[i])
		norm2 += float64(vector2[i]) * float64(vector2[i])
	}

	norm1 = math.Sqrt(norm1)
	norm2 = math.Sqrt(norm2)

	if norm1 == 0.0 || norm2 == 0.0 {
		return 0.0, nil
	}

	similarity := dotProduct / (norm1 * norm2)

	// 确保相似度在[0, 1]范围内
	if similarity < 0 {
		similarity = 0
	} else if similarity > 1 {
		similarity = 1
	}

	return similarity, nil
}

// extractMetadataFromPayload 从payload中提取元数据
func (s *DefaultVectorService) extractMetadataFromPayload(payload map[string]interface{}) *models.CacheMetadata {
	metadata := &models.CacheMetadata{}

	if source, ok := payload["source"].(string); ok {
		metadata.Source = source
	}

	// Tags字段已从CacheMetadata中移除，跳过提取

	if qualityScore, ok := payload["quality_score"].(float64); ok {
		metadata.QualityScore = qualityScore
	}

	if version, ok := payload["version"].(int); ok {
		metadata.Version = version
	}

	return metadata
}

// extractStatisticsFromPayload 从payload中提取统计信息
func (s *DefaultVectorService) extractStatisticsFromPayload(payload map[string]interface{}) *models.CacheStatistics {
	statistics := &models.CacheStatistics{}

	if hitCount, ok := payload["hit_count"].(int64); ok {
		statistics.HitCount = hitCount
	}

	if likeCount, ok := payload["like_count"].(int64); ok {
		statistics.LikeCount = likeCount
	}

	// DislikeCount 和 ResponseTime 字段已从CacheStatistics中移除
	// 跳过这些字段的提取

	if lastHitTimeUnix, ok := payload["last_hit_time"].(int64); ok {
		if lastHitTimeUnix > 0 {
			lastHitTime := time.Unix(lastHitTimeUnix, 0)
			statistics.LastHitTime = &lastHitTime
		}
	}

	return statistics
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SelectBestResult 选择最优结果
// 从候选结果中选择最符合查询意图的单个结果
func (s *DefaultVectorService) SelectBestResult(ctx context.Context, results []*models.VectorSearchResult, query *models.CacheQuery, strategy string) (*models.VectorSearchResult, error) {
	s.logger.InfoContext(ctx, "开始选择最优结果",
		"results_count", len(results),
		"strategy", strategy,
		"user_type", query.UserType)

	// 验证输入
	if len(results) == 0 {
		s.logger.WarnContext(ctx, "没有候选结果可供选择")
		return nil, fmt.Errorf("没有候选结果可供选择")
	}

	if query == nil {
		s.logger.ErrorContext(ctx, "查询请求不能为空")
		return nil, fmt.Errorf("查询请求不能为空")
	}

	// 使用默认策略如果未指定
	if strategy == "" {
		strategy = s.config.DefaultSelectionStrategy
	}

	// 创建策略实例
	selectionStrategy, err := s.strategyFactory.CreateStrategy(strategy)
	if err != nil {
		s.logger.ErrorContext(ctx, "创建选择策略失败",
			"strategy", strategy,
			"error", err)
		return nil, fmt.Errorf("创建选择策略失败: %w", err)
	}

	// 执行选择
	selectedResult, err := selectionStrategy.Select(ctx, results, query, make(map[string]interface{}))
	if err != nil {
		s.logger.ErrorContext(ctx, "选择结果失败",
			"strategy", strategy,
			"error", err)
		return nil, fmt.Errorf("选择结果失败: %w", err)
	}

	s.logger.InfoContext(ctx, "成功选择最优结果",
		"selected_id", selectedResult.ID,
		"selected_score", selectedResult.Score,
		"strategy", strategy)

	return selectedResult, nil
}
