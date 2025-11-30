package services

import (
	"context"
	"llm-cache/internal/domain/models"
	"llm-cache/pkg/status"
)

// PreprocessorFunc 预处理函数类型
// 用户自定义的预处理函数必须符合此签名
//
// 参数:
//
//	text: 待处理的文本
//	metadata: 元数据，可以包含额外的处理信息
//
// 返回:
//
//	string: 处理后的文本
type PreprocessorFunc func(text string, metadata map[string]interface{}) string

// RequestPreprocessingService 请求预处理服务接口
// 负责对用户原始查询请求进行预处理，提高查询质量和匹配准确性
// 采用函数式设计，支持用户注册自定义预处理函数
type RequestPreprocessingService interface {
	// PreprocessQuery 预处理查询请求
	// 使用注册的自定义预处理函数链来处理查询
	//
	// 参数:
	//   ctx: 上下文，用于控制请求生命周期和传递追踪信息
	//   request: 原始查询请求
	//
	// 返回:
	//   *models.PreprocessedRequest: 预处理后的请求，包含处理结果和元数据
	//   status.StatusCode: 处理状态码
	//   error: 错误信息
	PreprocessQuery(ctx context.Context, request *models.CacheQuery) (*models.PreprocessedRequest, status.StatusCode, error)

	// RegisterPreprocessor 注册预处理函数
	// 允许用户注册自定义的预处理函数，注册的函数将按注册顺序链式执行
	//
	// 参数:
	//   name: 预处理函数名称，用于标识和管理
	//   processor: 预处理函数
	//
	// 返回:
	//   error: 错误信息，如果名称已存在则返回错误
	RegisterPreprocessor(name string, processor PreprocessorFunc) error

	// UnregisterPreprocessor 取消注册预处理函数
	// 移除指定名称的预处理函数
	//
	// 参数:
	//   name: 预处理函数名称
	//
	// 返回:
	//   error: 错误信息，如果名称不存在则返回错误
	UnregisterPreprocessor(name string) error

	// ListPreprocessors 列出所有已注册的预处理函数名称
	//
	// 返回:
	//   []string: 预处理函数名称列表，按注册顺序返回
	ListPreprocessors() []string
}

// RequestPreprocessingConfig 请求预处理配置
type RequestPreprocessingConfig struct {
	// Timeout 处理超时时间（秒）
	Timeout int `json:"timeout" yaml:"timeout"`

	// EnableLogging 是否启用详细日志
	EnableLogging bool `json:"enable_logging" yaml:"enable_logging"`
}
