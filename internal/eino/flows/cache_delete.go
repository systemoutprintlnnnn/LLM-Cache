// Package flows 提供 Eino Graph 流程定义
package flows

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	milvusClient "github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	qdrantClient "github.com/qdrant/go-client/qdrant"
	redis "github.com/redis/go-redis/v9"

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

// CacheDeleteService 提供缓存删除服务的实现。
// 由于 Eino 框架目前未直接提供删除接口，该服务直接封装了向量数据库客户端进行删除操作。
type CacheDeleteService struct {
	provider string

	collection      string
	qdrantClient    *qdrantClient.Client
	milvusClient    milvusClient.Client
	milvusPartition string
	redisClient     *redis.Client
	redisIndex      string
	redisPrefix     string
	esClient        *elasticsearch.Client
	esIndex         string
}

// NewCacheDeleteService 创建一个新的缓存删除服务实例。
// 参数 cfg: Retriever 配置（包含向量数据库连接信息）。
// 返回: CacheDeleteService 指针或错误。
func NewCacheDeleteService(ctx context.Context, cfg *config.RetrieverConfig) (*CacheDeleteService, error) {
	service := &CacheDeleteService{
		provider:    cfg.Provider,
		collection:  cfg.Collection,
		redisIndex:  cfg.Redis.Index,
		redisPrefix: cfg.Redis.Prefix,
		esIndex:     cfg.ES8.Index,
	}

	switch cfg.Provider {
	case "qdrant":
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
		service.qdrantClient = client
	case "milvus":
		client, err := milvusClient.NewClient(ctx, milvusClient.Config{
			Address:  fmt.Sprintf("%s:%d", cfg.Milvus.Host, cfg.Milvus.Port),
			Username: cfg.Milvus.Username,
			Password: cfg.Milvus.Password,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create milvus client: %w", err)
		}
		service.milvusClient = client
		if len(cfg.Milvus.Partition) > 0 {
			service.milvusPartition = cfg.Milvus.Partition[0]
		}
	case "redis":
		service.redisClient = redis.NewClient(&redis.Options{
			Addr:     cfg.Redis.Addr,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
			Protocol: 2,
		})
	case "es8":
		esClient, err := elasticsearch.NewClient(elasticsearch.Config{
			Addresses: cfg.ES8.Addresses,
			Username:  cfg.ES8.Username,
			Password:  cfg.ES8.Password,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
		}
		service.esClient = esClient
	default:
		return nil, fmt.Errorf("unsupported retriever provider for delete: %s", cfg.Provider)
	}

	return service, nil
}

// Delete 执行批量删除操作。
// 根据 ID 列表从向量数据库中删除对应的缓存项。
// 参数 ctx: 上下文对象。
// 参数 input: 删除请求输入。
// 返回: 删除结果输出或错误。
func (s *CacheDeleteService) Delete(ctx context.Context, input *CacheDeleteInput) (*CacheDeleteOutput, error) {
	if len(input.CacheIDs) == 0 {
		return &CacheDeleteOutput{
			Success:      true,
			DeletedCount: 0,
		}, nil
	}

	switch s.provider {
	case "qdrant":
		return s.deleteFromQdrant(ctx, input.CacheIDs)
	case "milvus":
		return s.deleteFromMilvus(ctx, input.CacheIDs)
	case "redis":
		return s.deleteFromRedis(ctx, input.CacheIDs)
	case "es8":
		return s.deleteFromElasticsearch(ctx, input.CacheIDs)
	default:
		return &CacheDeleteOutput{
			Success:   false,
			FailedIDs: input.CacheIDs,
			Reason:    fmt.Sprintf("delete not supported for provider %s", s.provider),
		}, nil
	}
}

// DeleteSingle 执行单个缓存项删除操作。
// 是 Delete 方法的便捷封装。
// 参数 ctx: 上下文对象。
// 参数 cacheID: 缓存 ID。
// 参数 userType: 用户类型。
// 返回: 错误（如果操作失败）。
func (s *CacheDeleteService) DeleteSingle(ctx context.Context, cacheID string, userType string) error {
	output, err := s.Delete(ctx, &CacheDeleteInput{
		CacheIDs: []string{cacheID},
		UserType: userType,
	})

	if err != nil {
		return err
	}

	if !output.Success {
		return fmt.Errorf("delete failed: %s", output.Reason)
	}

	return nil
}

// GetByID 根据 ID 获取缓存详情（通常用于验证存在性）。
// 参数 ctx: 上下文对象。
// 参数 cacheID: 缓存 ID。
// 返回: 缓存项的 Payload 数据 Map 或错误。
func (s *CacheDeleteService) GetByID(ctx context.Context, cacheID string) (map[string]any, error) {
	if s.provider != "qdrant" {
		return nil, fmt.Errorf("get by id is only supported for qdrant, current: %s", s.provider)
	}

	pointID := &qdrantClient.PointId{
		PointIdOptions: &qdrantClient.PointId_Uuid{
			Uuid: cacheID,
		},
	}

	points, err := s.qdrantClient.Get(ctx, &qdrantClient.GetPoints{
		CollectionName: s.collection,
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

// convertQdrantValue 辅助函数：将 Qdrant Value 类型转换为 Go 原生类型。
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

// Close 关闭与向量数据库的连接。
func (s *CacheDeleteService) Close() error {
	if s.qdrantClient != nil {
		return s.qdrantClient.Close()
	}
	return nil
}

func (s *CacheDeleteService) deleteFromQdrant(ctx context.Context, ids []string) (*CacheDeleteOutput, error) {
	output := &CacheDeleteOutput{
		FailedIDs: make([]string, 0),
	}

	pointIDs := make([]*qdrantClient.PointId, 0, len(ids))
	for _, id := range ids {
		pointIDs = append(pointIDs, &qdrantClient.PointId{
			PointIdOptions: &qdrantClient.PointId_Uuid{
				Uuid: id,
			},
		})
	}

	_, err := s.qdrantClient.Delete(ctx, &qdrantClient.DeletePoints{
		CollectionName: s.collection,
		Points: &qdrantClient.PointsSelector{
			PointsSelectorOneOf: &qdrantClient.PointsSelector_Points{
				Points: &qdrantClient.PointsIdsList{
					Ids: pointIDs,
				},
			},
		},
	})

	if err != nil {
		output.Success = false
		output.Reason = err.Error()
		output.FailedIDs = ids
		return output, nil
	}

	output.Success = true
	output.DeletedCount = len(ids)
	return output, nil
}

func (s *CacheDeleteService) deleteFromMilvus(ctx context.Context, ids []string) (*CacheDeleteOutput, error) {
	output := &CacheDeleteOutput{
		FailedIDs: make([]string, 0),
	}

	if s.milvusClient == nil {
		return nil, errors.New("milvus client is not configured")
	}

	idColumn := entity.NewColumnVarChar("id", ids)
	if err := s.milvusClient.DeleteByPks(ctx, s.collection, s.milvusPartition, idColumn); err != nil {
		output.Success = false
		output.Reason = err.Error()
		output.FailedIDs = ids
		return output, nil
	}

	output.Success = true
	output.DeletedCount = len(ids)
	return output, nil
}

func (s *CacheDeleteService) deleteFromRedis(ctx context.Context, ids []string) (*CacheDeleteOutput, error) {
	output := &CacheDeleteOutput{
		FailedIDs: make([]string, 0),
	}

	if s.redisClient == nil {
		return nil, errors.New("redis client is not configured")
	}
	if s.redisIndex == "" {
		return nil, errors.New("redis index is not configured")
	}

	for _, id := range ids {
		keysToTry := []string{id}
		if s.redisPrefix != "" && !strings.HasPrefix(id, s.redisPrefix) {
			keysToTry = append([]string{s.redisPrefix + id}, keysToTry...)
		}

		deleted := false
		for _, key := range keysToTry {
			if _, err := s.redisClient.Do(ctx, "FT.DEL", s.redisIndex, key, "DD").Result(); err == nil {
				deleted = true
				break
			}
		}

		if deleted {
			output.DeletedCount++
		} else {
			output.FailedIDs = append(output.FailedIDs, id)
		}
	}

	output.Success = len(output.FailedIDs) == 0
	if !output.Success {
		output.Reason = "partial redis delete failure"
	}

	return output, nil
}

func (s *CacheDeleteService) deleteFromElasticsearch(ctx context.Context, ids []string) (*CacheDeleteOutput, error) {
	output := &CacheDeleteOutput{
		FailedIDs: make([]string, 0),
	}

	if s.esClient == nil {
		return nil, errors.New("elasticsearch client is not configured")
	}
	if s.esIndex == "" {
		return nil, errors.New("elasticsearch index is not configured")
	}

	for _, id := range ids {
		resp, err := s.esClient.Delete(s.esIndex, id)
		if err != nil {
			output.FailedIDs = append(output.FailedIDs, id)
			continue
		}
		if resp != nil && resp.Body != nil {
			_ = resp.Body.Close()
		}

		if resp != nil && resp.IsError() {
			output.FailedIDs = append(output.FailedIDs, id)
			continue
		}
		output.DeletedCount++
	}

	output.Success = len(output.FailedIDs) == 0
	if !output.Success {
		output.Reason = "partial elasticsearch delete failure"
	}

	return output, nil
}
