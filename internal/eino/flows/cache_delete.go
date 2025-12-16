// Package flows 提供 Eino Graph 流程定义
package flows

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	milvusClient "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	qdrantClient "github.com/qdrant/go-client/qdrant"
	"github.com/redis/go-redis/v9"
	"github.com/volcengine/volc-sdk-golang/service/vikingdb"

	"llm-cache/internal/eino/config"
)

// CacheDeleteInput 定义缓存删除请求的输入参数。
// 包含 ID 列表、用户类型和强制删除标志。
type CacheDeleteInput struct {
	CacheIDs []string `json:"cache_ids"`
	UserType string   `json:"user_type"`
	Force    bool     `json:"force,omitempty"`
}

// CacheDeleteOutput 定义缓存删除请求的输出结果。
// 包含成功状态、删除数量、失败 ID 列表和失败原因。
type CacheDeleteOutput struct {
	Success      bool     `json:"success"`
	DeletedCount int      `json:"deleted_count"`
	FailedIDs    []string `json:"failed_ids,omitempty"`
	Reason       string   `json:"reason,omitempty"`
}

// CacheDeleter 定义缓存删除操作的接口。
// 支持多种向量数据库后端。
type CacheDeleter interface {
	// Delete 执行批量删除操作
	Delete(ctx context.Context, input *CacheDeleteInput) (*CacheDeleteOutput, error)
	// DeleteSingle 执行单个缓存项删除操作
	DeleteSingle(ctx context.Context, cacheID string, userType string) error
	// GetByID 根据 ID 获取缓存详情
	GetByID(ctx context.Context, cacheID string) (map[string]any, error)
	// Close 关闭连接
	Close() error
}

// NewCacheDeleter 根据配置创建对应的 CacheDeleter 实例。
// 参数 cfg: Retriever 配置（包含向量数据库连接信息）。
// 返回: CacheDeleter 接口实例或错误。
func NewCacheDeleter(cfg *config.RetrieverConfig) (CacheDeleter, error) {
	switch cfg.Provider {
	case "qdrant":
		return NewQdrantDeleter(cfg)
	case "milvus":
		return NewMilvusDeleter(cfg)
	case "redis":
		return NewRedisDeleter(cfg)
	case "es8":
		return NewES8Deleter(cfg)
	case "vikingdb":
		return NewVikingDBDeleter(cfg)
	default:
		return nil, fmt.Errorf("unsupported delete provider: %s", cfg.Provider)
	}
}

// =============================================================================
// Qdrant Deleter
// =============================================================================

// QdrantDeleter 实现 Qdrant 的缓存删除操作
type QdrantDeleter struct {
	client     *qdrantClient.Client
	collection string
}

// NewQdrantDeleter 创建 Qdrant 删除服务实例
func NewQdrantDeleter(cfg *config.RetrieverConfig) (*QdrantDeleter, error) {
	clientCfg := &qdrantClient.Config{
		Host: cfg.Qdrant.Host,
		Port: cfg.Qdrant.Port,
	}

	if cfg.Qdrant.APIKey != "" {
		clientCfg.APIKey = cfg.Qdrant.APIKey
	}

	if cfg.Qdrant.UseTLS {
		clientCfg.UseTLS = true
	}

	client, err := qdrantClient.NewClient(clientCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create qdrant client: %w", err)
	}

	return &QdrantDeleter{
		client:     client,
		collection: cfg.Collection,
	}, nil
}

// Delete 执行批量删除操作
func (d *QdrantDeleter) Delete(ctx context.Context, input *CacheDeleteInput) (*CacheDeleteOutput, error) {
	if len(input.CacheIDs) == 0 {
		return &CacheDeleteOutput{Success: true, DeletedCount: 0}, nil
	}

	pointIDs := make([]*qdrantClient.PointId, 0, len(input.CacheIDs))
	for _, id := range input.CacheIDs {
		pointIDs = append(pointIDs, &qdrantClient.PointId{
			PointIdOptions: &qdrantClient.PointId_Uuid{Uuid: id},
		})
	}

	_, err := d.client.Delete(ctx, &qdrantClient.DeletePoints{
		CollectionName: d.collection,
		Points: &qdrantClient.PointsSelector{
			PointsSelectorOneOf: &qdrantClient.PointsSelector_Points{
				Points: &qdrantClient.PointsIdsList{Ids: pointIDs},
			},
		},
	})

	if err != nil {
		return &CacheDeleteOutput{
			Success:   false,
			Reason:    err.Error(),
			FailedIDs: input.CacheIDs,
		}, nil
	}

	return &CacheDeleteOutput{
		Success:      true,
		DeletedCount: len(input.CacheIDs),
	}, nil
}

// DeleteSingle 执行单个缓存项删除操作
func (d *QdrantDeleter) DeleteSingle(ctx context.Context, cacheID string, userType string) error {
	output, err := d.Delete(ctx, &CacheDeleteInput{CacheIDs: []string{cacheID}, UserType: userType})
	if err != nil {
		return err
	}
	if !output.Success {
		return fmt.Errorf("delete failed: %s", output.Reason)
	}
	return nil
}

// GetByID 根据 ID 获取缓存详情
func (d *QdrantDeleter) GetByID(ctx context.Context, cacheID string) (map[string]any, error) {
	pointID := &qdrantClient.PointId{
		PointIdOptions: &qdrantClient.PointId_Uuid{Uuid: cacheID},
	}

	points, err := d.client.Get(ctx, &qdrantClient.GetPoints{
		CollectionName: d.collection,
		Ids:            []*qdrantClient.PointId{pointID},
		WithPayload:    &qdrantClient.WithPayloadSelector{SelectorOptions: &qdrantClient.WithPayloadSelector_Enable{Enable: true}},
	})

	if err != nil {
		return nil, fmt.Errorf("get point: %w", err)
	}

	if len(points) == 0 {
		return nil, fmt.Errorf("cache not found: %s", cacheID)
	}

	result := make(map[string]any)
	for k, v := range points[0].Payload {
		result[k] = convertQdrantValue(v)
	}

	return result, nil
}

// Close 关闭连接
func (d *QdrantDeleter) Close() error {
	if d.client != nil {
		return d.client.Close()
	}
	return nil
}

// convertQdrantValue 辅助函数：将 Qdrant Value 类型转换为 Go 原生类型
func convertQdrantValue(v *qdrantClient.Value) any {
	if v == nil {
		return nil
	}
	switch val := v.Kind.(type) {
	case *qdrantClient.Value_StringValue:
		return val.StringValue
	case *qdrantClient.Value_IntegerValue:
		return val.IntegerValue
	case *qdrantClient.Value_DoubleValue:
		return val.DoubleValue
	case *qdrantClient.Value_BoolValue:
		return val.BoolValue
	case *qdrantClient.Value_NullValue:
		return nil
	default:
		return nil
	}
}

// =============================================================================
// Milvus Deleter
// =============================================================================

// MilvusDeleter 实现 Milvus 的缓存删除操作
type MilvusDeleter struct {
	client     milvusClient.Client
	collection string
}

// NewMilvusDeleter 创建 Milvus 删除服务实例
func NewMilvusDeleter(cfg *config.RetrieverConfig) (*MilvusDeleter, error) {
	ctx := context.Background()
	client, err := milvusClient.NewClient(ctx, milvusClient.Config{
		Address:  fmt.Sprintf("%s:%d", cfg.Milvus.Host, cfg.Milvus.Port),
		Username: cfg.Milvus.Username,
		Password: cfg.Milvus.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create milvus client: %w", err)
	}

	return &MilvusDeleter{
		client:     client,
		collection: cfg.Collection,
	}, nil
}

// Delete 执行批量删除操作
func (d *MilvusDeleter) Delete(ctx context.Context, input *CacheDeleteInput) (*CacheDeleteOutput, error) {
	if len(input.CacheIDs) == 0 {
		return &CacheDeleteOutput{Success: true, DeletedCount: 0}, nil
	}

	// 使用 DeleteByPks 删除
	// 假设 ID 字段名为 "id"，类型为 VarChar
	ids := entity.NewColumnVarChar("id", input.CacheIDs)
	err := d.client.DeleteByPks(ctx, d.collection, "", ids)

	if err != nil {
		return &CacheDeleteOutput{
			Success:   false,
			Reason:    err.Error(),
			FailedIDs: input.CacheIDs,
		}, nil
	}

	return &CacheDeleteOutput{
		Success:      true,
		DeletedCount: len(input.CacheIDs),
	}, nil
}

// DeleteSingle 执行单个缓存项删除操作
func (d *MilvusDeleter) DeleteSingle(ctx context.Context, cacheID string, userType string) error {
	output, err := d.Delete(ctx, &CacheDeleteInput{CacheIDs: []string{cacheID}, UserType: userType})
	if err != nil {
		return err
	}
	if !output.Success {
		return fmt.Errorf("delete failed: %s", output.Reason)
	}
	return nil
}

// GetByID 根据 ID 获取缓存详情
func (d *MilvusDeleter) GetByID(ctx context.Context, cacheID string) (map[string]any, error) {
	// 使用 Query 方法获取
	expr := fmt.Sprintf(`id == "%s"`, cacheID)
	results, err := d.client.Query(ctx, d.collection, nil, expr, []string{"*"})
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("cache not found: %s", cacheID)
	}

	// 将结果转换为 map
	result := make(map[string]any)
	for _, col := range results {
		result[col.Name()] = col.FieldData()
	}

	return result, nil
}

// Close 关闭连接
func (d *MilvusDeleter) Close() error {
	if d.client != nil {
		return d.client.Close()
	}
	return nil
}

// =============================================================================
// Redis Deleter
// =============================================================================

// RedisDeleter 实现 Redis 的缓存删除操作
type RedisDeleter struct {
	client *redis.Client
	prefix string
	index  string
}

// NewRedisDeleter 创建 Redis 删除服务实例
func NewRedisDeleter(cfg *config.RetrieverConfig) (*RedisDeleter, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	return &RedisDeleter{
		client: rdb,
		prefix: cfg.Redis.Prefix,
		index:  cfg.Redis.Index,
	}, nil
}

// Delete 执行批量删除操作
func (d *RedisDeleter) Delete(ctx context.Context, input *CacheDeleteInput) (*CacheDeleteOutput, error) {
	if len(input.CacheIDs) == 0 {
		return &CacheDeleteOutput{Success: true, DeletedCount: 0}, nil
	}

	// 构建带前缀的键
	keys := make([]string, len(input.CacheIDs))
	for i, id := range input.CacheIDs {
		keys[i] = d.prefix + id
	}

	// 使用 DEL 命令批量删除
	result := d.client.Del(ctx, keys...)
	if result.Err() != nil {
		return &CacheDeleteOutput{
			Success:   false,
			Reason:    result.Err().Error(),
			FailedIDs: input.CacheIDs,
		}, nil
	}

	return &CacheDeleteOutput{
		Success:      true,
		DeletedCount: int(result.Val()),
	}, nil
}

// DeleteSingle 执行单个缓存项删除操作
func (d *RedisDeleter) DeleteSingle(ctx context.Context, cacheID string, userType string) error {
	output, err := d.Delete(ctx, &CacheDeleteInput{CacheIDs: []string{cacheID}, UserType: userType})
	if err != nil {
		return err
	}
	if !output.Success {
		return fmt.Errorf("delete failed: %s", output.Reason)
	}
	return nil
}

// GetByID 根据 ID 获取缓存详情
func (d *RedisDeleter) GetByID(ctx context.Context, cacheID string) (map[string]any, error) {
	key := d.prefix + cacheID
	result := d.client.HGetAll(ctx, key)
	if result.Err() != nil {
		return nil, fmt.Errorf("get hash failed: %w", result.Err())
	}

	data := result.Val()
	if len(data) == 0 {
		return nil, fmt.Errorf("cache not found: %s", cacheID)
	}

	// 转换为 map[string]any
	m := make(map[string]any)
	for k, v := range data {
		m[k] = v
	}

	return m, nil
}

// Close 关闭连接
func (d *RedisDeleter) Close() error {
	if d.client != nil {
		return d.client.Close()
	}
	return nil
}

// =============================================================================
// Elasticsearch 8 Deleter
// =============================================================================

// ES8Deleter 实现 Elasticsearch 8 的缓存删除操作
type ES8Deleter struct {
	client *elasticsearch.Client
	index  string
}

// NewES8Deleter 创建 Elasticsearch 8 删除服务实例
func NewES8Deleter(cfg *config.RetrieverConfig) (*ES8Deleter, error) {
	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: cfg.ES8.Addresses,
		Username:  cfg.ES8.Username,
		Password:  cfg.ES8.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	return &ES8Deleter{
		client: esClient,
		index:  cfg.ES8.Index,
	}, nil
}

// Delete 执行批量删除操作
func (d *ES8Deleter) Delete(ctx context.Context, input *CacheDeleteInput) (*CacheDeleteOutput, error) {
	if len(input.CacheIDs) == 0 {
		return &CacheDeleteOutput{Success: true, DeletedCount: 0}, nil
	}

	// 构建 Bulk Delete 请求
	var buf bytes.Buffer
	for _, id := range input.CacheIDs {
		meta := map[string]map[string]string{
			"delete": {"_index": d.index, "_id": id},
		}
		if err := json.NewEncoder(&buf).Encode(meta); err != nil {
			return nil, fmt.Errorf("failed to encode delete meta: %w", err)
		}
	}

	// 执行 Bulk 请求
	res, err := d.client.Bulk(bytes.NewReader(buf.Bytes()), d.client.Bulk.WithContext(ctx))
	if err != nil {
		return &CacheDeleteOutput{
			Success:   false,
			Reason:    err.Error(),
			FailedIDs: input.CacheIDs,
		}, nil
	}
	defer res.Body.Close()

	if res.IsError() {
		return &CacheDeleteOutput{
			Success:   false,
			Reason:    res.String(),
			FailedIDs: input.CacheIDs,
		}, nil
	}

	return &CacheDeleteOutput{
		Success:      true,
		DeletedCount: len(input.CacheIDs),
	}, nil
}

// DeleteSingle 执行单个缓存项删除操作
func (d *ES8Deleter) DeleteSingle(ctx context.Context, cacheID string, userType string) error {
	output, err := d.Delete(ctx, &CacheDeleteInput{CacheIDs: []string{cacheID}, UserType: userType})
	if err != nil {
		return err
	}
	if !output.Success {
		return fmt.Errorf("delete failed: %s", output.Reason)
	}
	return nil
}

// GetByID 根据 ID 获取缓存详情
func (d *ES8Deleter) GetByID(ctx context.Context, cacheID string) (map[string]any, error) {
	res, err := d.client.Get(d.index, cacheID, d.client.Get.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("get document failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("cache not found: %s", cacheID)
	}

	var result map[string]any
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 提取 _source 字段
	if source, ok := result["_source"].(map[string]any); ok {
		return source, nil
	}

	return result, nil
}

// Close 关闭连接
func (d *ES8Deleter) Close() error {
	// Elasticsearch client 不需要显式关闭
	return nil
}

// =============================================================================
// VikingDB Deleter
// =============================================================================

// VikingDBDeleter 实现 VikingDB 的缓存删除操作
type VikingDBDeleter struct {
	service    *vikingdb.VikingDBService
	collection *vikingdb.Collection
}

// NewVikingDBDeleter 创建 VikingDB 删除服务实例
func NewVikingDBDeleter(cfg *config.RetrieverConfig) (*VikingDBDeleter, error) {
	service := vikingdb.NewVikingDBService(
		cfg.VikingDB.Host,
		cfg.VikingDB.Region,
		cfg.VikingDB.AK,
		cfg.VikingDB.SK,
		cfg.VikingDB.Scheme,
	)

	if cfg.VikingDB.ConnectionTimeout > 0 {
		service.SetConnectionTimeout(cfg.VikingDB.ConnectionTimeout)
	}

	collection, err := service.GetCollection(cfg.Collection)
	if err != nil {
		return nil, fmt.Errorf("failed to get vikingdb collection: %w", err)
	}

	return &VikingDBDeleter{
		service:    service,
		collection: collection,
	}, nil
}

// Delete 执行批量删除操作
func (d *VikingDBDeleter) Delete(ctx context.Context, input *CacheDeleteInput) (*CacheDeleteOutput, error) {
	if len(input.CacheIDs) == 0 {
		return &CacheDeleteOutput{Success: true, DeletedCount: 0}, nil
	}

	// 使用 VikingDB 的删除接口
	err := d.collection.DeleteData(input.CacheIDs)
	if err != nil {
		return &CacheDeleteOutput{
			Success:   false,
			Reason:    err.Error(),
			FailedIDs: input.CacheIDs,
		}, nil
	}

	return &CacheDeleteOutput{
		Success:      true,
		DeletedCount: len(input.CacheIDs),
	}, nil
}

// DeleteSingle 执行单个缓存项删除操作
func (d *VikingDBDeleter) DeleteSingle(ctx context.Context, cacheID string, userType string) error {
	output, err := d.Delete(ctx, &CacheDeleteInput{CacheIDs: []string{cacheID}, UserType: userType})
	if err != nil {
		return err
	}
	if !output.Success {
		return fmt.Errorf("delete failed: %s", output.Reason)
	}
	return nil
}

// GetByID 根据 ID 获取缓存详情
func (d *VikingDBDeleter) GetByID(ctx context.Context, cacheID string) (map[string]any, error) {
	// VikingDB 使用 FetchData 获取数据
	data, err := d.collection.FetchData([]interface{}{cacheID})
	if err != nil {
		return nil, fmt.Errorf("fetch data failed: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("cache not found: %s", cacheID)
	}

	// 转换第一条数据为 map
	result := make(map[string]any)
	if data[0].Fields != nil {
		for k, v := range data[0].Fields {
			result[k] = v
		}
	}

	return result, nil
}

// Close 关闭连接
func (d *VikingDBDeleter) Close() error {
	// VikingDB service 不需要显式关闭
	return nil
}

// =============================================================================
// 保持向后兼容性的别名
// =============================================================================

// CacheDeleteService 是 QdrantDeleter 的别名，保持向后兼容性
// Deprecated: 请使用 NewCacheDeleter 工厂函数
type CacheDeleteService = QdrantDeleter

// NewCacheDeleteService 创建一个新的缓存删除服务实例（仅支持 Qdrant）
// Deprecated: 请使用 NewCacheDeleter 工厂函数
func NewCacheDeleteService(cfg *config.RetrieverConfig) (*CacheDeleteService, error) {
	return NewQdrantDeleter(cfg)
}
