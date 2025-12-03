package models

import (
	"fmt"
	"math"
	"time"
)

// Vector 定义了高维向量的数据结构。
// 包含向量 ID、数值数组、维度信息以及相关的元数据（如创建时间、模型名称）。
type Vector struct {
	// ID 向量唯一标识符
	ID string `json:"id"`

	// Values 向量值数组
	Values []float32 `json:"values" validate:"required,min=1"`

	// Dimension 向量维度
	Dimension int `json:"dimension"`

	// CreateTime 创建时间
	CreateTime time.Time `json:"create_time"`

	// UpdateTime 更新时间
	UpdateTime time.Time `json:"update_time"`

	// Normalized 是否已归一化
	Normalized bool `json:"normalized"`

	// ModelName 生成向量的模型名称
	ModelName string `json:"model_name"`
}

// VectorSearchRequest 定义了向量搜索请求的参数。
// 支持通过文本或直接向量进行搜索，包含 TopK、相似度阈值和过滤条件。
type VectorSearchRequest struct {
	// QueryText 查询文本（用于日志记录）
	QueryText string `json:"query_text,omitempty"`

	// QueryID 查询向量ID（可选，如果已有向量可直接使用）
	QueryID string `json:"query_id,omitempty"`

	// QueryVector 查询向量（主要搜索方式）
	QueryVector []float32 `json:"query_vector,omitempty"`

	// TopK 返回最相似的K个结果
	TopK int `json:"top_k" validate:"required,min=1,max=100"`

	// SimilarityThreshold 相似度阈值
	SimilarityThreshold float64 `json:"similarity_threshold" validate:"required,min=0,max=1"`

	// Filters 过滤条件
	Filters map[string]interface{} `json:"filters,omitempty"`

	// UserType 用户类型，用于场景隔离
	UserType string `json:"user_type" validate:"required"`
}

// VectorSearchResult 定义了向量搜索的单个结果项。
// 包含匹配的向量 ID、相似度分数、向量数据和关联的 Payload。
type VectorSearchResult struct {
	// ID 向量ID
	ID string `json:"id" validate:"required"`

	// Score 相似度分数 (0.0-1.1, 允许稍微超过1.0以容纳浮点数计算误差)
	Score float64 `json:"score" validate:"required,min=0,max=1.1"`

	// Vector 向量数据（可选）
	Vector *Vector `json:"vector,omitempty"`

	// Payload 关联的负载数据
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// VectorSearchResponse 定义了向量搜索的完整响应。
// 包含搜索结果列表、总数、耗时和查询相关信息。
type VectorSearchResponse struct {
	// Results 搜索结果列表
	Results []VectorSearchResult `json:"results"`

	// TotalCount 总结果数量
	TotalCount int `json:"total_count" validate:"min=0"`

	// SearchTime 搜索耗时（毫秒）
	SearchTime float64 `json:"search_time" validate:"min=0"`

	// QueryInfo 查询信息
	QueryInfo VectorQueryInfo `json:"query_info"`
}

// VectorQueryInfo 定义了搜索查询的元数据。
// 包含查询向量的维度和是否应用了过滤器。
type VectorQueryInfo struct {
	// Dimension 查询向量维度
	Dimension int `json:"dimension" validate:"required,gt=0"`

	// FilterApplied 是否应用了过滤条件
	FilterApplied bool `json:"filter_applied"`
}

// VectorBatchStoreRequest 定义了批量向量存储请求的参数。
// 包含要存储的向量列表、目标集合名称和是否启用 Upsert 模式。
type VectorBatchStoreRequest struct {
	// Vectors 要存储的向量列表
	Vectors []VectorStoreItem `json:"vectors" validate:"required,min=1"`

	// CollectionName 集合名称
	CollectionName string `json:"collection_name" validate:"required"`

	// UpsertMode 是否为更新插入模式
	UpsertMode bool `json:"upsert_mode,omitempty"`
}

// VectorStoreRequest 定义了单个向量存储请求的参数。
// 包含向量 ID、数值、目标集合名称和关联 Payload。
type VectorStoreRequest struct {
	// ID 向量ID
	ID string `json:"id" validate:"required"`

	// Vector 向量值
	Vector []float32 `json:"vector" validate:"required,min=1"`

	// CollectionName 集合名称
	CollectionName string `json:"collection_name" validate:"required"`

	// Payload 关联的负载数据
	Payload map[string]interface{} `json:"payload,omitempty"`

	// UpsertMode 是否为更新插入模式
	UpsertMode bool `json:"upsert_mode,omitempty"`
}

// VectorStoreItem 定义了批量存储中的单个向量项。
// 包含 ID、向量数值和 Payload。
type VectorStoreItem struct {
	// ID 向量ID
	ID string `json:"id" validate:"required"`

	// Vector 向量值
	Vector []float32 `json:"vector" validate:"required,min=1"`

	// Payload 关联的负载数据
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// VectorBatchStoreResponse 定义了批量向量存储的操作结果。
// 包含成功和失败的数量、失败 ID 列表和操作耗时。
type VectorBatchStoreResponse struct {
	// Success 是否存储成功
	Success bool `json:"success"`

	// StoredCount 成功存储的数量
	StoredCount int `json:"stored_count" validate:"min=0"`

	// FailedCount 存储失败的数量
	FailedCount int `json:"failed_count" validate:"min=0"`

	// FailedIDs 存储失败的ID列表
	FailedIDs []string `json:"failed_ids,omitempty"`

	// Message 结果消息
	Message string `json:"message,omitempty"`

	// StoreTime 存储耗时（毫秒）
	StoreTime float64 `json:"store_time" validate:"min=0"`
}

// VectorStoreResponse 定义了单个向量存储的操作结果。
// 包含存储成功的向量 ID 和操作耗时。
type VectorStoreResponse struct {
	// Success 是否存储成功
	Success bool `json:"success"`

	// VectorID 存储的向量ID
	VectorID string `json:"vector_id" validate:"required"`

	// Message 结果消息
	Message string `json:"message,omitempty"`

	// StoreTime 存储耗时（毫秒）
	StoreTime float64 `json:"store_time" validate:"min=0"`
}

// NewVector 创建并初始化一个新的 Vector 实例。
// 自动计算维度并设置创建时间。
// 参数 id: 向量的唯一标识符。
// 参数 values: 向量的数值数组。
// 返回: 初始化后的 Vector 指针。
func NewVector(id string, values []float32) *Vector {
	now := time.Now()
	return &Vector{
		ID:         id,
		Values:     values,
		Dimension:  len(values),
		CreateTime: now,
		UpdateTime: now,
		Normalized: false,
		ModelName:  "",
	}
}

// Validate 检查向量数据的有效性。
// 验证 ID 是否为空，数值数组是否非空，维度是否匹配，以及是否包含无效值 (NaN/Inf)。
// 返回: 如果验证失败返回 error，否则返回 nil。
func (v *Vector) Validate() error {
	if v.ID == "" {
		return fmt.Errorf("vector ID cannot be empty")
	}

	if len(v.Values) == 0 {
		return fmt.Errorf("vector values cannot be empty")
	}

	if v.Dimension != len(v.Values) {
		return fmt.Errorf("dimension mismatch: expected %d, got %d", v.Dimension, len(v.Values))
	}

	// 检查是否包含无效值
	for i, val := range v.Values {
		if math.IsNaN(float64(val)) || math.IsInf(float64(val), 0) {
			return fmt.Errorf("invalid value at index %d: %f", i, val)
		}
	}

	return nil
}

// Normalize 对向量进行归一化处理（使用 L2 范数）。
// 归一化后，向量的模长为 1，便于计算余弦相似度。
// 如果向量模长为 0，则不进行操作。
func (v *Vector) Normalize() {
	norm := v.L2Norm()
	if norm == 0 {
		return // 零向量不进行归一化
	}

	for i := range v.Values {
		v.Values[i] /= norm
	}
	v.Normalized = true
	v.UpdateTime = time.Now()
}

// L2Norm 计算向量的 L2 范数（欧几里得范数）。
// 返回: 向量的模长。
func (v *Vector) L2Norm() float32 {
	var sum float32
	for _, val := range v.Values {
		sum += val * val
	}
	return float32(math.Sqrt(float64(sum)))
}
