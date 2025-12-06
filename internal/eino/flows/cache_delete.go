// Package flows 提供 Eino Graph 流程定义
package flows

import (
	"context"
	"fmt"

	qdrantClient "github.com/qdrant/go-client/qdrant"

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
// 由于 Eino 框架目前未直接提供删除接口，该服务直接封装了向量数据库客户端（目前支持 Qdrant）进行删除操作。
type CacheDeleteService struct {
	client     *qdrantClient.Client
	collection string
}

// NewCacheDeleteService 创建一个新的缓存删除服务实例。
// 参数 cfg: Retriever 配置（包含向量数据库连接信息）。
// 返回: CacheDeleteService 指针或错误。
func NewCacheDeleteService(cfg *config.RetrieverConfig) (*CacheDeleteService, error) {
	if cfg.Provider != "qdrant" {
		return nil, fmt.Errorf("cache delete is only implemented for qdrant provider, got: %s", cfg.Provider)
	}

	// 创建 Qdrant 客户端
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

	return &CacheDeleteService{
		client:     client,
		collection: cfg.Collection,
	}, nil
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

	output := &CacheDeleteOutput{
		FailedIDs: make([]string, 0),
	}

	deletedCount := 0

	// 构建删除条件
	// 使用 point IDs 进行删除
	pointIDs := make([]*qdrantClient.PointId, 0, len(input.CacheIDs))
	for _, id := range input.CacheIDs {
		pointIDs = append(pointIDs, &qdrantClient.PointId{
			PointIdOptions: &qdrantClient.PointId_Uuid{
				Uuid: id,
			},
		})
	}

	// 执行删除
	_, err := s.client.Delete(ctx, &qdrantClient.DeletePoints{
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
		output.FailedIDs = input.CacheIDs
		return output, nil
	}

	deletedCount = len(input.CacheIDs)
	output.Success = true
	output.DeletedCount = deletedCount

	return output, nil
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
	// 使用 Qdrant Get 方法获取单个点
	pointID := &qdrantClient.PointId{
		PointIdOptions: &qdrantClient.PointId_Uuid{
			Uuid: cacheID,
		},
	}

	points, err := s.client.Get(ctx, &qdrantClient.GetPoints{
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

	// 转换 payload 为 map
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
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}
