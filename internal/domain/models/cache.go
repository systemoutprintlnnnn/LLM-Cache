package models

import (
	"time"
)

// CacheItem 定义了语义缓存中的核心数据单元。
// 它包含原始问题、生成的答案、用户类型、向量表示以及相关的元数据和统计信息。
type CacheItem struct {
	// ID 缓存项的唯一标识符
	ID string `json:"id" validate:"required"`

	// Question 用户原始问题文本
	Question string `json:"question" validate:"required,min=1,max=1000"`

	// Answer LLM生成的答案文本
	Answer string `json:"answer" validate:"required,min=1,max=10000"`

	// UserType 用于场景隔离的用户类型标识
	UserType string `json:"user_type" validate:"required"`

	// Vector 问题的向量表示
	Vector []float32 `json:"vector,omitempty"`

	// Metadata 缓存项的元数据
	Metadata CacheMetadata `json:"metadata"`

	// Statistics 缓存项的统计信息
	Statistics CacheStatistics `json:"statistics"`

	// CreateTime 创建时间
	CreateTime time.Time `json:"create_time"`

	// UpdateTime 更新时间
	UpdateTime time.Time `json:"update_time"`
}

// CacheMetadata 定义了缓存项的元数据信息。
// 包含数据来源、质量评估分数和版本号等附加信息。
type CacheMetadata struct {
	// Source 数据来源标识
	Source string `json:"source,omitempty"`

	// QualityScore 质量评估分数 (-1.0 或 0.0-1.0，其中 -1.0 表示无自定义质量函数可用)
	QualityScore float64 `json:"quality_score,omitempty" validate:"omitempty,min=-1.0,max=1.0"`

	// Version 数据版本号
	Version int `json:"version,omitempty"`
}

// CacheStatistics 定义了缓存项的使用统计信息。
// 包含命中次数、点赞次数和最后命中时间，用于缓存淘汰和热点分析。
type CacheStatistics struct {
	// HitCount 命中次数
	HitCount int64 `json:"hit_count"`

	// LikeCount 点赞次数
	LikeCount int64 `json:"like_count"`

	// LastHitTime 最后命中时间
	LastHitTime *time.Time `json:"last_hit_time,omitempty"`
}

// CacheQuery 定义了缓存查询请求的参数结构。
// 包含查询问题、用户类型以及可选的搜索配置（如相似度阈值、返回数量）。
type CacheQuery struct {
	// Question 查询问题
	Question string `json:"question" validate:"required,min=1"`

	// UserType 用户类型，用于场景隔离
	UserType string `json:"user_type" validate:"required"`

	// SimilarityThreshold 相似度阈值，覆盖默认配置
	SimilarityThreshold float64 `json:"similarity_threshold,omitempty" validate:"omitempty,min=0,max=1"`

	// TopK 返回结果数量，覆盖默认配置
	TopK int `json:"top_k,omitempty" validate:"omitempty,min=1,max=100"`

	// Filters 额外的过滤条件
	Filters map[string]interface{} `json:"filters,omitempty"`
}

// CacheResult 定义了缓存查询的返回结果。
// 包含是否命中、匹配的答案、相似度分数以及相关的元数据和统计信息。
type CacheResult struct {
	// Found 是否找到匹配的缓存
	Found bool `json:"found"`

	// CacheID 缓存项ID
	CacheID string `json:"cache_id,omitempty"`

	// Answer 缓存的答案
	Answer string `json:"answer,omitempty"`

	// Similarity 相似度分数
	Similarity float64 `json:"similarity,omitempty"`

	// ResponseTime 响应时间（毫秒）
	ResponseTime float64 `json:"response_time,omitempty" validate:"omitempty,min=0"`

	// Metadata 缓存元数据（可选）
	Metadata *CacheMetadata `json:"metadata,omitempty"`

	// Statistics 统计信息（可选）
	Statistics *CacheStatistics `json:"statistics,omitempty"`

	// Reason 失败原因（当 found=false 时可选）
	Reason string `json:"reason,omitempty"`
}

// CacheWriteRequest 定义了写入缓存的请求参数。
// 包含问题、答案、用户类型以及元数据，支持强制写入选项。
type CacheWriteRequest struct {
	// Question 问题文本
	Question string `json:"question" validate:"required,min=1,max=1000"`

	// Answer 答案文本
	Answer string `json:"answer" validate:"required,min=1,max=10000"`

	// UserType 用户类型
	UserType string `json:"user_type" validate:"required"`

	// Metadata 元数据
	Metadata *CacheMetadata `json:"metadata,omitempty"`

	// ForceWrite 是否强制写入（跳过质量评估）
	ForceWrite bool `json:"force_write,omitempty"`
}

// CacheWriteResult 定义了写入缓存的操作结果。
// 包含操作是否成功、生成的缓存 ID、结果消息以及质量评估分数。
type CacheWriteResult struct {
	// Success 是否写入成功
	Success bool `json:"success"`

	// CacheID 缓存项ID
	CacheID string `json:"cache_id,omitempty"`

	// Message 结果消息（失败原因统一放在此字段）
	Message string `json:"message,omitempty"`

	// QualityScore 质量评估分数 (-1.0 或 0.0-1.0)
	QualityScore float64 `json:"quality_score,omitempty" validate:"omitempty,min=-1.0,max=1.0"`
}

// CacheDeleteRequest 定义了删除缓存的请求参数。
// 支持批量删除，需要提供 ID 列表和用户类型。
type CacheDeleteRequest struct {
	// CacheIDs 要删除的缓存项ID列表
	CacheIDs []string `json:"cache_ids" validate:"required,min=1"`

	// UserType 用户类型，用于权限验证
	UserType string `json:"user_type" validate:"required"`

	// Force 是否强制删除
	Force bool `json:"force,omitempty"`
}

// CacheDeleteResult 定义了删除缓存的操作结果。
// 包含是否成功、实际删除的数量以及失败的 ID 列表。
type CacheDeleteResult struct {
	// Success 是否删除成功
	Success bool `json:"success"`

	// DeletedCount 实际删除的数量
	DeletedCount int `json:"deleted_count" validate:"min=0"`

	// FailedIDs 删除失败的ID列表
	FailedIDs []string `json:"failed_ids,omitempty"`

	// Message 结果消息
	Message string `json:"message,omitempty"`
}
