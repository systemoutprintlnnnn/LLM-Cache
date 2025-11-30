package qdrant

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"llm-cache/configs"
	"llm-cache/internal/domain/models"
	"llm-cache/internal/domain/repositories"

	"github.com/qdrant/go-client/qdrant"
)

// QdrantVectorStore Qdrant向量存储实现
// 实现 VectorRepository 接口，提供向量数据的持久化操作
type QdrantVectorStore struct {
	client *QdrantClient
	config *configs.QdrantConfig
	logger *slog.Logger
}

// NewQdrantVectorStore 创建新的Qdrant向量存储实例
func NewQdrantVectorStore(ctx context.Context, config *configs.QdrantConfig, logger *slog.Logger) (repositories.VectorRepository, error) {
	if config == nil {
		return nil, fmt.Errorf("qdrant config cannot be nil")
	}

	if logger == nil {
		logger = slog.Default()
	}

	// 创建Qdrant客户端
	client, err := NewQdrantClient(ctx, config, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create qdrant client: %w", err)
	}

	store := &QdrantVectorStore{
		client: client,
		config: config,
		logger: logger,
	}

	logger.InfoContext(ctx, "Qdrant向量存储初始化成功",
		"collection", config.CollectionName,
		"vector_size", config.VectorSize)

	return store, nil
}

// Store 存储单个向量数据 - 实现VectorRepository接口
func (v *QdrantVectorStore) Store(ctx context.Context, request *models.VectorStoreRequest) (*models.VectorStoreResponse, error) {
	startTime := time.Now()

	// 参数验证
	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	if request.ID == "" {
		return nil, fmt.Errorf("vector ID cannot be empty")
	}

	if len(request.Vector) == 0 {
		return nil, fmt.Errorf("vector cannot be empty")
	}

	// 验证向量维度
	if len(request.Vector) != v.config.VectorSize {
		v.logger.WarnContext(ctx, "向量维度不匹配",
			"vector_id", request.ID,
			"expected", v.config.VectorSize,
			"actual", len(request.Vector))
		return &models.VectorStoreResponse{
			Success:   false,
			VectorID:  request.ID,
			Message:   fmt.Sprintf("vector dimension mismatch: expected %d, got %d", v.config.VectorSize, len(request.Vector)),
			StoreTime: float64(time.Since(startTime).Milliseconds()),
		}, nil
	}

	// 丰富payload数据
	enrichedPayload := v.enrichPayload(request.Payload)

	// 执行单个向量存储
	err := v.client.UpsertPoint(ctx, request.ID, request.Vector, enrichedPayload)
	if err != nil {
		v.logger.ErrorContext(ctx, "单个向量存储失败",
			"vector_id", request.ID,
			"error", err)
		return &models.VectorStoreResponse{
			Success:   false,
			VectorID:  request.ID,
			Message:   fmt.Sprintf("vector storage failed: %v", err),
			StoreTime: float64(time.Since(startTime).Milliseconds()),
		}, nil
	}

	duration := time.Since(startTime)
	v.logger.InfoContext(ctx, "单个向量存储完成",
		"vector_id", request.ID,
		"duration_ms", duration.Milliseconds())

	return &models.VectorStoreResponse{
		Success:   true,
		VectorID:  request.ID,
		Message:   "vector stored successfully",
		StoreTime: float64(duration.Milliseconds()),
	}, nil
}

// BatchStore 存储向量数据 - 实现VectorRepository接口
func (v *QdrantVectorStore) BatchStore(ctx context.Context, request *models.VectorBatchStoreRequest) (*models.VectorBatchStoreResponse, error) {
	startTime := time.Now()

	// 参数验证
	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	if len(request.Vectors) == 0 {
		return &models.VectorBatchStoreResponse{
			Success:     true,
			StoredCount: 0,
			Message:     "no vectors to store",
			StoreTime:   0,
		}, nil
	}

	// 转换为批量点格式
	batchPoints := make([]BatchPoint, 0, len(request.Vectors))
	failedIDs := make([]string, 0)

	for _, item := range request.Vectors {
		// 验证向量维度
		if len(item.Vector) != v.config.VectorSize {
			failedIDs = append(failedIDs, item.ID)
			v.logger.WarnContext(ctx, "向量维度不匹配",
				"vector_id", item.ID,
				"expected", v.config.VectorSize,
				"actual", len(item.Vector))
			continue
		}

		// 丰富payload数据
		enrichedPayload := v.enrichPayload(item.Payload)

		batchPoint := BatchPoint{
			ID:      item.ID,
			Vector:  item.Vector,
			Payload: enrichedPayload,
		}

		batchPoints = append(batchPoints, batchPoint)
	}

	// 执行批量存储
	err := v.client.UpsertBatch(ctx, batchPoints)
	if err != nil {
		v.logger.ErrorContext(ctx, "批量存储向量失败",
			"total_count", len(request.Vectors),
			"valid_count", len(batchPoints),
			"error", err)
		return &models.VectorBatchStoreResponse{
			Success:     false,
			StoredCount: 0,
			FailedCount: len(request.Vectors),
			FailedIDs:   v.extractIDs(request.Vectors),
			Message:     fmt.Sprintf("batch storage failed: %v", err),
			StoreTime:   float64(time.Since(startTime).Milliseconds()),
		}, nil
	}

	duration := time.Since(startTime)
	storedCount := len(batchPoints)
	failedCount := len(failedIDs)

	v.logger.InfoContext(ctx, "批量向量存储完成",
		"total_count", len(request.Vectors),
		"stored_count", storedCount,
		"failed_count", failedCount,
		"duration_ms", duration.Milliseconds())

	return &models.VectorBatchStoreResponse{
		Success:     true,
		StoredCount: storedCount,
		FailedCount: failedCount,
		FailedIDs:   failedIDs,
		Message:     fmt.Sprintf("successfully stored %d vectors", storedCount),
		StoreTime:   float64(duration.Milliseconds()),
	}, nil
}

// Search 向量相似性搜索 - 实现VectorRepository接口
func (v *QdrantVectorStore) Search(ctx context.Context, request *models.VectorSearchRequest) (*models.VectorSearchResponse, error) {
	startTime := time.Now()

	// 参数验证
	if request == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	if len(request.QueryVector) == 0 && request.QueryID == "" {
		return nil, fmt.Errorf("either query_vector or query_id must be provided")
	}

	// 构建过滤器
	filter := v.buildFilter(request.UserType, request.Filters)

	// 构建搜索参数
	limit := uint64(request.TopK)
	var scoreThreshold *float32
	if request.SimilarityThreshold > 0 {
		threshold := float32(request.SimilarityThreshold)
		scoreThreshold = &threshold
	}

	// 获取查询向量：优先使用QueryVector，其次使用QueryID
	var queryVector []float32
	if len(request.QueryVector) > 0 {
		// 直接使用提供的向量
		queryVector = request.QueryVector
		v.logger.InfoContext(ctx, "使用提供的查询向量",
			"vector_dimension", len(queryVector),
			"query_text", request.QueryText)
	} else if request.QueryID != "" {
		// 根据ID获取向量
		point, err := v.client.GetPoint(ctx, request.QueryID, false, true)
		if err != nil {
			v.logger.ErrorContext(ctx, "获取查询向量失败",
				"query_id", request.QueryID,
				"error", err)
			return nil, fmt.Errorf("failed to get query vector: %w", err)
		}
		if point == nil {
			return nil, fmt.Errorf("query vector not found: %s", request.QueryID)
		}
		queryVector = point.Vector
		v.logger.InfoContext(ctx, "通过ID获取查询向量",
			"query_id", request.QueryID,
			"vector_dimension", len(queryVector))
	}

	// 执行搜索
	results, err := v.client.SearchPoints(ctx, queryVector, limit, scoreThreshold, filter)
	if err != nil {
		v.logger.ErrorContext(ctx, "向量搜索失败",
			"user_type", request.UserType,
			"top_k", request.TopK,
			"threshold", request.SimilarityThreshold,
			"error", err)
		return nil, fmt.Errorf("vector search failed: %w", err)
	}

	// 转换搜索结果
	searchResults := make([]models.VectorSearchResult, 0, len(results))
	for _, result := range results {
		searchResult := models.VectorSearchResult{
			ID:      result.ID,
			Score:   float64(result.Score),
			Payload: result.Payload,
		}

		// 如果需要，可以包含向量数据
		if result.Vector != nil {
			vector := models.NewVector(result.ID, result.Vector)
			searchResult.Vector = vector
		}

		searchResults = append(searchResults, searchResult)
	}

	duration := time.Since(startTime)
	v.logger.InfoContext(ctx, "向量搜索完成",
		"user_type", request.UserType,
		"result_count", len(searchResults),
		"duration_ms", duration.Milliseconds())

	return &models.VectorSearchResponse{
		Results:    searchResults,
		TotalCount: len(searchResults),
		SearchTime: float64(duration.Milliseconds()),
		QueryInfo: models.VectorQueryInfo{
			Dimension:     len(queryVector),
			FilterApplied: len(request.Filters) > 0 || request.UserType != "",
		},
	}, nil
}

// Delete 删除向量数据 - 实现VectorRepository接口
func (v *QdrantVectorStore) Delete(ctx context.Context, ids []string, userType string) (*models.CacheDeleteResult, error) {
	if len(ids) == 0 {
		return &models.CacheDeleteResult{
			Success:      true,
			DeletedCount: 0,
			Message:      "没有要删除的向量",
		}, nil
	}

	// 如果指定了用户类型，需要先验证这些向量是否属于该用户类型
	if userType != "" {
		// 为安全起见，先获取这些向量的信息进行验证
		for _, id := range ids {
			point, err := v.client.GetPoint(ctx, id, true, false)
			if err != nil {
				v.logger.WarnContext(ctx, "获取向量信息失败，跳过删除",
					"vector_id", id,
					"error", err)
				continue
			}
			if point == nil {
				v.logger.WarnContext(ctx, "向量不存在，跳过删除", "vector_id", id)
				continue
			}

			// 检查用户类型
			if payloadUserType, ok := point.Payload["user_type"].(string); ok {
				if payloadUserType != userType {
					v.logger.WarnContext(ctx, "用户类型不匹配，跳过删除",
						"vector_id", id,
						"expected_user_type", userType,
						"actual_user_type", payloadUserType)
					continue
				}
			}
		}
	}

	err := v.client.DeleteBatch(ctx, ids)
	if err != nil {
		v.logger.ErrorContext(ctx, "批量删除向量失败",
			"count", len(ids),
			"user_type", userType,
			"error", err)
		return &models.CacheDeleteResult{
			Success:      false,
			DeletedCount: 0,
			FailedIDs:    ids,
			Message:      fmt.Sprintf("批量删除向量失败: %v", err),
		}, fmt.Errorf("failed to delete vectors: %w", err)
	}

	v.logger.InfoContext(ctx, "批量向量删除成功",
		"count", len(ids),
		"user_type", userType)
	return &models.CacheDeleteResult{
		Success:      true,
		DeletedCount: len(ids),
		Message:      "删除成功",
	}, nil
}

// GetByID 根据ID获取向量 - 实现VectorRepository接口
func (v *QdrantVectorStore) GetByID(ctx context.Context, id string) (*models.Vector, error) {
	if id == "" {
		return nil, fmt.Errorf("vector ID cannot be empty")
	}

	// 获取向量点
	point, err := v.client.GetPoint(ctx, id, false, true)
	if err != nil {
		v.logger.ErrorContext(ctx, "获取向量失败",
			"vector_id", id,
			"error", err)
		return nil, fmt.Errorf("failed to get vector: %w", err)
	}

	if point == nil {
		return nil, nil // 未找到
	}

	// 构建向量对象
	vector := models.NewVector(point.ID, point.Vector)

	v.logger.DebugContext(ctx, "向量获取成功",
		"vector_id", id,
		"dimension", len(point.Vector))

	return vector, nil
}

// 以下是附加的辅助方法，不是接口的一部分

// buildFilter 构建Qdrant过滤器
func (v *QdrantVectorStore) buildFilter(userType string, filters map[string]interface{}) *qdrant.Filter {
	conditions := make([]*qdrant.Condition, 0)

	// 添加用户类型过滤
	if userType != "" {
		conditions = append(conditions, qdrant.NewMatch("user_type", userType))
	}

	// 添加其他过滤条件
	for key, value := range filters {
		switch v := value.(type) {
		case string:
			conditions = append(conditions, qdrant.NewMatch(key, v))
		}
	}

	if len(conditions) == 0 {
		return nil
	}

	return &qdrant.Filter{
		Must: conditions,
	}
}

// enrichPayload 丰富payload数据，可选择性添加向量元数据
func (v *QdrantVectorStore) enrichPayload(originalPayload map[string]interface{}) map[string]interface{} {
	payload := make(map[string]interface{})

	// 复制原始payload
	for k, v := range originalPayload {
		payload[k] = v
	}

	// 添加时间戳
	payload["created_at"] = time.Now().Format(time.RFC3339)

	return payload
}

// extractIDs 提取向量ID列表
func (v *QdrantVectorStore) extractIDs(vectors []models.VectorStoreItem) []string {
	ids := make([]string, len(vectors))
	for i, vector := range vectors {
		ids[i] = vector.ID
	}
	return ids
}

// Close 关闭向量存储连接
func (v *QdrantVectorStore) Close() error {
	if v.client != nil {
		return v.client.Close()
	}
	return nil
}
