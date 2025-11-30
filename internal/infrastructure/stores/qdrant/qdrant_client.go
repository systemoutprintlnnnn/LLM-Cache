package qdrant

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"llm-cache/configs"

	"github.com/qdrant/go-client/qdrant"
)

// QdrantClient Qdrant客户端封装
// 提供向量数据库的基础操作能力，封装gRPC连接管理和错误处理
type QdrantClient struct {
	client     *qdrant.Client
	config     *configs.QdrantConfig
	logger     *slog.Logger
	collection string
}

// QdrantClientConfig Qdrant客户端配置
type QdrantClientConfig struct {
	Host           string        `json:"host"`
	Port           int           `json:"port"`
	APIKey         string        `json:"api_key,omitempty"`
	UseTLS         bool          `json:"use_tls"`
	CollectionName string        `json:"collection_name"`
	VectorSize     int           `json:"vector_size"`
	Distance       string        `json:"distance"`
	Timeout        time.Duration `json:"timeout"`
	MaxRetries     int           `json:"max_retries"`
	RetryDelay     time.Duration `json:"retry_delay"`
}

// NewQdrantClient 创建新的Qdrant客户端实例
func NewQdrantClient(ctx context.Context, config *configs.QdrantConfig, logger *slog.Logger) (*QdrantClient, error) {
	if config == nil {
		return nil, fmt.Errorf("qdrant config cannot be nil")
	}

	if logger == nil {
		logger = slog.Default()
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		logger.ErrorContext(ctx, "Qdrant配置验证失败", "error", err)
		return nil, fmt.Errorf("invalid qdrant config: %w", err)
	}

	// 创建Qdrant客户端配置
	clientConfig := &qdrant.Config{
		Host:   config.Host,
		Port:   config.Port,
		APIKey: config.APIKey,
		UseTLS: config.APIKey != "", // 如果有API Key则使用TLS
	}

	// 创建客户端
	client, err := qdrant.NewClient(clientConfig)
	if err != nil {
		logger.ErrorContext(ctx, "创建Qdrant客户端失败",
			"host", config.Host,
			"port", config.Port,
			"error", err)
		return nil, fmt.Errorf("failed to create qdrant client: %w", err)
	}

	qdrantClient := &QdrantClient{
		client:     client,
		config:     config,
		logger:     logger,
		collection: config.CollectionName,
	}

	// 测试连接并确保集合存在
	if err := qdrantClient.ensureCollection(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure collection: %w", err)
	}

	logger.InfoContext(ctx, "Qdrant客户端初始化成功",
		"host", config.Host,
		"port", config.Port,
		"collection", config.CollectionName)

	return qdrantClient, nil
}

// ensureCollection 确保集合存在，如果不存在则创建
func (q *QdrantClient) ensureCollection(ctx context.Context) error {
	// 检查集合是否存在
	exists, err := q.collectionExists(ctx)
	if err != nil {
		return fmt.Errorf("failed to check collection existence: %w", err)
	}

	if exists {
		return nil
	}

	// 创建集合
	return q.createCollection(ctx)
}

// collectionExists 检查集合是否存在
func (q *QdrantClient) collectionExists(ctx context.Context) (bool, error) {
	collections, err := q.client.ListCollections(ctx)
	if err != nil {
		q.logger.ErrorContext(ctx, "获取集合列表失败", "error", err)
		return false, err
	}

	for _, collectionName := range collections {
		if collectionName == q.collection {
			return true, nil
		}
	}

	return false, nil
}

// createCollection 创建向量集合
func (q *QdrantClient) createCollection(ctx context.Context) error {
	// 转换距离类型
	distance, err := q.parseDistance()
	if err != nil {
		return err
	}

	// 创建集合配置
	vectorsConfig := qdrant.NewVectorsConfig(&qdrant.VectorParams{
		Size:     uint64(q.config.VectorSize),
		Distance: distance,
	})

	// 创建集合
	err = q.client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: q.collection,
		VectorsConfig:  vectorsConfig,
	})

	if err != nil {
		q.logger.ErrorContext(ctx, "创建向量集合失败",
			"collection", q.collection,
			"vector_size", q.config.VectorSize,
			"distance", q.config.Distance,
			"error", err)
		return fmt.Errorf("failed to create collection %s: %w", q.collection, err)
	}

	q.logger.InfoContext(ctx, "向量集合创建成功",
		"collection", q.collection,
		"vector_size", q.config.VectorSize,
		"distance", q.config.Distance)

	return nil
}

// parseDistance 解析距离类型
func (q *QdrantClient) parseDistance() (qdrant.Distance, error) {
	switch q.config.Distance {
	case "cosine":
		return qdrant.Distance_Cosine, nil
	case "euclidean":
		return qdrant.Distance_Euclid, nil
	case "dot":
		return qdrant.Distance_Dot, nil
	case "manhattan":
		return qdrant.Distance_Manhattan, nil
	default:
		return qdrant.Distance_Cosine, fmt.Errorf("unsupported distance type: %s", q.config.Distance)
	}
}

// UpsertPoint 插入或更新单个向量点
func (q *QdrantClient) UpsertPoint(ctx context.Context, id string, vector []float32, payload map[string]interface{}) error {
	if len(vector) != q.config.VectorSize {
		return fmt.Errorf("vector dimension mismatch: expected %d, got %d", q.config.VectorSize, len(vector))
	}

	// 创建点结构
	point := &qdrant.PointStruct{
		Id:      &qdrant.PointId{PointIdOptions: &qdrant.PointId_Uuid{Uuid: id}},
		Vectors: qdrant.NewVectors(vector...),
		Payload: qdrant.NewValueMap(payload),
	}

	// 执行插入
	waitUpsert := true
	_, err := q.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: q.collection,
		Points:         []*qdrant.PointStruct{point},
		Wait:           &waitUpsert,
	})

	if err != nil {
		q.logger.ErrorContext(ctx, "向量点插入失败",
			"id", id,
			"collection", q.collection,
			"error", err)
		return fmt.Errorf("failed to upsert point %s: %w", id, err)
	}

	q.logger.DebugContext(ctx, "向量点插入成功", "id", id, "collection", q.collection)
	return nil
}

// UpsertBatch 批量插入或更新向量点
func (q *QdrantClient) UpsertBatch(ctx context.Context, points []BatchPoint) error {
	if len(points) == 0 {
		return nil
	}

	// 转换为Qdrant点结构
	qdrantPoints := make([]*qdrant.PointStruct, 0, len(points))
	for _, point := range points {
		if len(point.Vector) != q.config.VectorSize {
			return fmt.Errorf("vector dimension mismatch for point %s: expected %d, got %d",
				point.ID, q.config.VectorSize, len(point.Vector))
		}

		qdrantPoint := &qdrant.PointStruct{
			Id:      &qdrant.PointId{PointIdOptions: &qdrant.PointId_Uuid{Uuid: point.ID}},
			Vectors: qdrant.NewVectors(point.Vector...),
			Payload: qdrant.NewValueMap(point.Payload),
		}
		qdrantPoints = append(qdrantPoints, qdrantPoint)
	}

	// 执行批量插入
	waitUpsert := true
	_, err := q.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: q.collection,
		Points:         qdrantPoints,
		Wait:           &waitUpsert,
	})

	if err != nil {
		q.logger.ErrorContext(ctx, "批量向量点插入失败",
			"count", len(points),
			"collection", q.collection,
			"error", err)
		return fmt.Errorf("failed to upsert batch points: %w", err)
	}

	q.logger.DebugContext(ctx, "批量向量点插入成功",
		"count", len(points),
		"collection", q.collection)
	return nil
}

// BatchPoint 批量点结构
type BatchPoint struct {
	ID      string                 `json:"id"`
	Vector  []float32              `json:"vector"`
	Payload map[string]interface{} `json:"payload"`
}

// SearchPoints 搜索相似向量点
func (q *QdrantClient) SearchPoints(ctx context.Context, queryVector []float32, limit uint64, scoreThreshold *float32, filter *qdrant.Filter) ([]*SearchResult, error) {
	if len(queryVector) != q.config.VectorSize {
		return nil, fmt.Errorf("query vector dimension mismatch: expected %d, got %d", q.config.VectorSize, len(queryVector))
	}

	// 构建查询请求
	queryRequest := &qdrant.QueryPoints{
		CollectionName: q.collection,
		Query:          qdrant.NewQuery(queryVector...),
		Limit:          &limit,
		Filter:         filter,
		ScoreThreshold: scoreThreshold,
		WithPayload: &qdrant.WithPayloadSelector{
			SelectorOptions: &qdrant.WithPayloadSelector_Enable{Enable: true},
		},
	}

	// 执行搜索
	q.logger.InfoContext(ctx, "开始向量搜索", "collection", q.collection, "limit", limit)

	queryResult, err := q.client.Query(ctx, queryRequest)
	if err != nil {
		q.logger.ErrorContext(ctx, "向量搜索失败",
			"collection", q.collection,
			"error", err)
		return nil, fmt.Errorf("failed to search points: %w", err)
	}

	// 转换结果
	results := make([]*SearchResult, 0, len(queryResult))
	for _, point := range queryResult {
		result := &SearchResult{
			Score: point.Score,
		}

		// 转换PointId为字符串
		switch id := point.Id.PointIdOptions.(type) {
		case *qdrant.PointId_Num:
			result.ID = fmt.Sprintf("%d", id.Num)
		case *qdrant.PointId_Uuid:
			result.ID = id.Uuid
		}

		// 提取payload
		if point.Payload != nil {
			result.Payload = convertPayload(point.Payload)
		}

		results = append(results, result)
	}

	q.logger.DebugContext(ctx, "向量搜索完成",
		"collection", q.collection,
		"result_count", len(results),
		"limit", limit)

	return results, nil
}

// SearchResult 搜索结果
type SearchResult struct {
	ID      string                 `json:"id"`
	Score   float32                `json:"score"`
	Vector  []float32              `json:"vector,omitempty"`
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// GetPoint 根据ID获取向量点
func (q *QdrantClient) GetPoint(ctx context.Context, id string, withPayload bool, withVector bool) (*SearchResult, error) {
	// 构建查询请求
	getRequest := &qdrant.GetPoints{
		CollectionName: q.collection,
		Ids:            []*qdrant.PointId{qdrant.NewID(id)},
	}

	if withPayload {
		getRequest.WithPayload = qdrant.NewWithPayloadInclude()
	}

	// 执行查询
	getResult, err := q.client.Get(ctx, getRequest)
	if err != nil {
		q.logger.ErrorContext(ctx, "获取向量点失败",
			"id", id,
			"collection", q.collection,
			"error", err)
		return nil, fmt.Errorf("failed to get point %s: %w", id, err)
	}

	if len(getResult) == 0 {
		return nil, nil // 未找到
	}

	point := getResult[0]

	// 转换PointId为字符串
	var idStr string
	switch id := point.Id.PointIdOptions.(type) {
	case *qdrant.PointId_Num:
		idStr = fmt.Sprintf("%d", id.Num)
	case *qdrant.PointId_Uuid:
		idStr = id.Uuid
	}

	result := &SearchResult{
		ID: idStr,
	}

	// 提取payload
	if withPayload && point.Payload != nil {
		result.Payload = convertPayload(point.Payload)
	}

	// 提取vector
	if withVector && point.Vectors != nil {
		if vectors := point.Vectors.GetVector(); vectors != nil {
			result.Vector = vectors.Data
		}
	}

	return result, nil
}

// DeletePoint 删除单个向量点
func (q *QdrantClient) DeletePoint(ctx context.Context, id string) error {
	waitDelete := true
	_, err := q.client.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: q.collection,
		Points: &qdrant.PointsSelector{
			PointsSelectorOneOf: &qdrant.PointsSelector_Points{
				Points: &qdrant.PointsIdsList{
					Ids: []*qdrant.PointId{qdrant.NewID(id)},
				},
			},
		},
		Wait: &waitDelete,
	})

	if err != nil {
		q.logger.ErrorContext(ctx, "删除向量点失败",
			"id", id,
			"collection", q.collection,
			"error", err)
		return fmt.Errorf("failed to delete point %s: %w", id, err)
	}

	q.logger.DebugContext(ctx, "向量点删除成功", "id", id, "collection", q.collection)
	return nil
}

// DeleteBatch 批量删除向量点
func (q *QdrantClient) DeleteBatch(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	// 转换ID列表
	pointIds := make([]*qdrant.PointId, 0, len(ids))
	for _, id := range ids {
		pointIds = append(pointIds, qdrant.NewID(id))
	}

	waitDelete := true
	_, err := q.client.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: q.collection,
		Points: &qdrant.PointsSelector{
			PointsSelectorOneOf: &qdrant.PointsSelector_Points{
				Points: &qdrant.PointsIdsList{
					Ids: pointIds,
				},
			},
		},
		Wait: &waitDelete,
	})

	if err != nil {
		q.logger.ErrorContext(ctx, "批量删除向量点失败",
			"count", len(ids),
			"collection", q.collection,
			"error", err)
		return fmt.Errorf("failed to delete batch points: %w", err)
	}

	q.logger.DebugContext(ctx, "批量向量点删除成功",
		"count", len(ids),
		"collection", q.collection)
	return nil
}

// CountPoints 统计向量点数量
func (q *QdrantClient) CountPoints(ctx context.Context, filter *qdrant.Filter) (uint64, error) {
	exact := true
	countRequest := &qdrant.CountPoints{
		CollectionName: q.collection,
		Filter:         filter,
		Exact:          &exact,
	}

	countResult, err := q.client.Count(ctx, countRequest)
	if err != nil {
		q.logger.ErrorContext(ctx, "统计向量点数量失败",
			"collection", q.collection,
			"error", err)
		return 0, fmt.Errorf("failed to count points: %w", err)
	}

	return countResult, nil
}

// HealthCheck 健康检查
func (q *QdrantClient) HealthCheck(ctx context.Context) error {
	healthResult, err := q.client.HealthCheck(ctx)
	if err != nil {
		q.logger.ErrorContext(ctx, "Qdrant健康检查失败", "error", err)
		return fmt.Errorf("qdrant health check failed: %w", err)
	}

	if healthResult.Title != "qdrant - vector search engine" {
		return fmt.Errorf("unexpected health check response: %s", healthResult.Title)
	}

	q.logger.DebugContext(ctx, "Qdrant健康检查成功", "version", healthResult.Version)
	return nil
}

// GetCollectionInfo 获取集合信息
func (q *QdrantClient) GetCollectionInfo(ctx context.Context) (map[string]interface{}, error) {
	collectionInfo, err := q.client.GetCollectionInfo(ctx, q.collection)
	if err != nil {
		q.logger.ErrorContext(ctx, "获取集合信息失败",
			"collection", q.collection,
			"error", err)
		return nil, fmt.Errorf("failed to get collection info: %w", err)
	}

	info := map[string]interface{}{
		"status":         collectionInfo.Status.String(),
		"vectors_count":  collectionInfo.VectorsCount,
		"segments_count": collectionInfo.SegmentsCount,
		"config":         collectionInfo.Config,
	}

	return info, nil
}

// Close 关闭客户端连接
func (q *QdrantClient) Close() error {
	if q.client != nil {
		return q.client.Close()
	}
	return nil
}

// convertPayload 转换payload格式
func convertPayload(payload map[string]*qdrant.Value) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range payload {
		result[key] = convertValue(value)
	}
	return result
}

// convertValue 转换Value到interface{}
func convertValue(value *qdrant.Value) interface{} {
	if value == nil {
		return nil
	}

	switch v := value.Kind.(type) {
	case *qdrant.Value_NullValue:
		return nil
	case *qdrant.Value_BoolValue:
		return v.BoolValue
	case *qdrant.Value_IntegerValue:
		return v.IntegerValue
	case *qdrant.Value_DoubleValue:
		return v.DoubleValue
	case *qdrant.Value_StringValue:
		return v.StringValue
	case *qdrant.Value_ListValue:
		list := make([]interface{}, len(v.ListValue.Values))
		for i, item := range v.ListValue.Values {
			list[i] = convertValue(item)
		}
		return list
	case *qdrant.Value_StructValue:
		return convertPayload(v.StructValue.Fields)
	default:
		return nil
	}
}
