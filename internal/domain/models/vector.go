package models

import (
	"fmt"
	"math"
	"time"
)

// Vector 向量数据结构
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

// VectorSearchRequest 向量搜索请求
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

// VectorSearchResult 向量搜索结果
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

// VectorSearchResponse 向量搜索响应
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

// VectorQueryInfo 查询信息
type VectorQueryInfo struct {
	// Dimension 查询向量维度
	Dimension int `json:"dimension" validate:"required,gt=0"`

	// FilterApplied 是否应用了过滤条件
	FilterApplied bool `json:"filter_applied"`
}

// VectorBatchStoreRequest 【批量】向量存储请求
type VectorBatchStoreRequest struct {
	// Vectors 要存储的向量列表
	Vectors []VectorStoreItem `json:"vectors" validate:"required,min=1"`

	// CollectionName 集合名称
	CollectionName string `json:"collection_name" validate:"required"`

	// UpsertMode 是否为更新插入模式
	UpsertMode bool `json:"upsert_mode,omitempty"`
}

// VectorStoreRequest 【单个】向量存储请求
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

// VectorStoreItem 向量存储项
type VectorStoreItem struct {
	// ID 向量ID
	ID string `json:"id" validate:"required"`

	// Vector 向量值
	Vector []float32 `json:"vector" validate:"required,min=1"`

	// Payload 关联的负载数据
	Payload map[string]interface{} `json:"payload,omitempty"`
}

// VectorBatchStoreResponse 向量存储响应
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

// VectorStoreResponse 【单个】向量存储响应
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

// NewVector 创建新的向量
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

// Validate 验证向量数据的有效性
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

// Normalize 向量归一化（L2范数）
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

// L2Norm 计算L2范数
func (v *Vector) L2Norm() float32 {
	var sum float32
	for _, val := range v.Values {
		sum += val * val
	}
	return float32(math.Sqrt(float64(sum)))
}
