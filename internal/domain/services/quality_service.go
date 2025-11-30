package services

import (
	"context"
	"llm-cache/internal/domain/models"
	"llm-cache/pkg/status"
)

// QualityService 质量评估服务接口
// 负责评估问答对的质量，决定是否允许写入缓存，确保缓存数据的高质量
type QualityService interface {
	// AssessQuality 综合质量评估
	// 对问答对进行全面的质量评估，返回评估结果和建议
	//
	// 参数:
	//   ctx: 上下文，用于控制请求生命周期和传递追踪信息
	//   request: 质量评估请求
	//
	// 返回:
	//   *models.QualityAssessmentResult: 质量评估结果
	//   status.StatusCode: 评估状态码
	//   error: 错误信息
	AssessQuality(ctx context.Context, request *models.QualityAssessmentRequest) (*models.QualityAssessmentResult, status.StatusCode, error)

	// AssessQuestionQuality 评估问题质量
	// 专门评估问题文本的质量
	//
	// 参数:
	//   ctx: 上下文
	//   question: 问题文本
	//   userType: 用户类型
	//
	// 返回:
	//   float64: 问题质量分数 (0.0-1.0)
	//   []string: 质量问题描述
	//   status.StatusCode: 评估状态码
	//   error: 错误信息
	AssessQuestionQuality(ctx context.Context, question string, userType string) (float64, []string, status.StatusCode, error)

	// AssessAnswerQuality 评估答案质量
	// 专门评估答案文本的质量
	//
	// 参数:
	//   ctx: 上下文
	//   answer: 答案文本
	//   question: 对应的问题文本（用于相关性评估）
	//   userType: 用户类型
	//
	// 返回:
	//   float64: 答案质量分数 (0.0-1.0)
	//   []string: 质量问题描述
	//   status.StatusCode: 评估状态码
	//   error: 错误信息
	AssessAnswerQuality(ctx context.Context, answer string, question string, userType string) (float64, []string, status.StatusCode, error)

	// CheckBlacklist 检查黑名单
	// 检查问答对是否包含黑名单关键词或模式
	//
	// 参数:
	//   ctx: 上下文
	//   question: 问题文本
	//   answer: 答案文本
	//   userType: 用户类型
	//
	// 返回:
	//   bool: 是否命中黑名单
	//   []string: 命中的黑名单项
	//   status.StatusCode: 检查状态码
	//   error: 错误信息
	CheckBlacklist(ctx context.Context, question string, answer string, userType string) (bool, []string, status.StatusCode, error)

	// CalculateOverallScore 计算综合分数
	// 基于多个评估维度计算综合质量分数
	//
	// 参数:
	//   ctx: 上下文
	//   scores: 各维度分数
	//   weights: 各维度权重
	//
	// 返回:
	//   float64: 综合分数 (0.0-1.0)
	//   status.StatusCode: 计算状态码
	//   error: 错误信息
	CalculateOverallScore(ctx context.Context, scores map[string]float64, weights map[string]float64) (float64, status.StatusCode, error)

	// IsQualityAcceptable 判断质量是否可接受
	// 基于配置的阈值判断质量是否达到要求
	//
	// 参数:
	//   ctx: 上下文
	//   score: 质量分数
	//   userType: 用户类型
	//
	// 返回:
	//   bool: 质量是否可接受
	//   string: 拒绝原因（如果不可接受）
	//   status.StatusCode: 判断状态码
	//   error: 错误信息
	IsQualityAcceptable(ctx context.Context, score float64, userType string) (bool, string, status.StatusCode, error)

	// ApplyCustomQualityFunction 应用自定义质量评估函数
	// 用户可以提供自定义的质量评估函数来执行特定的质量检查
	//
	// 参数:
	//   ctx: 上下文
	//   question: 问题文本
	//   answer: 答案文本
	//   customFunc: 自定义质量评估函数
	//
	// 返回:
	//   float64: 质量分数 (0.0-1.0)
	//   status.StatusCode: 评估状态码
	//   error: 错误信息
	ApplyCustomQualityFunction(ctx context.Context, question string, answer string, customFunc CustomQualityFunction) (float64, status.StatusCode, error)

	// RegisterCustomFunction 注册自定义质量评估函数
	// 将自定义函数注册到服务中，以便在综合评估中使用
	//
	// 参数:
	//   name: 函数名称
	//   function: 自定义质量评估函数
	//   weight: 函数权重
	RegisterCustomFunction(name string, function CustomQualityFunction, weight float64)

	// GetRegisteredFunctions 获取所有已注册的函数名称
	// 用于调试和监控，返回当前注册的所有函数名称列表
	//
	// 返回:
	//   []string: 函数名称列表
	GetRegisteredFunctions() []string

	// GetFunctionWeight 获取指定函数的权重
	//
	// 参数:
	//   functionName: 函数名称
	//
	// 返回:
	//   float64: 函数权重，如果函数不存在则返回0.0
	GetFunctionWeight(functionName string) float64
}

// CustomQualityFunction 自定义质量评估函数类型
// 用户可以实现此函数来提供自定义的质量评估逻辑
//
// 参数:
//
//	question: 问题文本
//	answer: 答案文本
//	config: 配置参数
//
// 返回:
//
//	float64: 质量分数 (仅返回分数)
type CustomQualityFunction func(question string, answer string, config map[string]interface{}) float64

// QualityAssessmentStrategy 质量评估策略接口
// 允许插拔式的质量评估策略实现
type QualityAssessmentStrategy interface {
	// Name 策略名称
	Name() string

	// Assess 执行评估
	//
	// 参数:
	//   ctx: 上下文
	//   question: 问题文本
	//   answer: 答案文本
	//   userType: 用户类型
	//   config: 策略配置
	//
	// 返回:
	//   float64: 评估分数 (0.0-1.0)
	//   map[string]interface{}: 评估详情
	//   error: 错误信息
	Assess(ctx context.Context, question string, answer string, userType string, config map[string]interface{}) (float64, map[string]interface{}, error)

	// GetWeight 获取策略权重
	//
	// 参数:
	//   userType: 用户类型
	//
	// 返回:
	//   float64: 权重值
	GetWeight(userType string) float64

	// Validate 验证策略配置
	//
	// 参数:
	//   config: 策略配置
	//
	// 返回:
	//   bool: 配置是否有效
	//   error: 验证错误
	Validate(config map[string]interface{}) (bool, error)
}

// BlacklistChecker 黑名单检查器接口
type BlacklistChecker interface {
	// CheckKeywords 检查关键词黑名单
	//
	// 参数:
	//   text: 待检查文本
	//   userType: 用户类型
	//
	// 返回:
	//   bool: 是否命中黑名单
	//   []string: 命中的关键词
	//   error: 错误信息
	CheckKeywords(text string, userType string) (bool, []string, error)

	// CheckPatterns 检查模式黑名单
	//
	// 参数:
	//   text: 待检查文本
	//   userType: 用户类型
	//
	// 返回:
	//   bool: 是否命中黑名单
	//   []string: 命中的模式
	//   error: 错误信息
	CheckPatterns(text string, userType string) (bool, []string, error)
}

// QualityConfig 质量评估配置
type QualityConfig struct {
	// Strategies 启用的质量评估策略
	Strategies []string `json:"strategies" yaml:"strategies"`

	// StrategyWeights 策略权重配置
	StrategyWeights map[string]float64 `json:"strategy_weights" yaml:"strategy_weights"`

	// UserTypeThresholds 不同用户类型的质量阈值
	UserTypeThresholds map[string]float64 `json:"user_type_thresholds" yaml:"user_type_thresholds"`

	// DefaultThreshold 默认质量阈值
	DefaultThreshold float64 `json:"default_threshold" yaml:"default_threshold"`

	// MinQuestionLength 最小问题长度
	MinQuestionLength int `json:"min_question_length" yaml:"min_question_length"`

	// MaxQuestionLength 最大问题长度
	MaxQuestionLength int `json:"max_question_length" yaml:"max_question_length"`

	// MinAnswerLength 最小答案长度
	MinAnswerLength int `json:"min_answer_length" yaml:"min_answer_length"`

	// MaxAnswerLength 最大答案长度
	MaxAnswerLength int `json:"max_answer_length" yaml:"max_answer_length"`

	// BlacklistKeywords 黑名单关键词
	BlacklistKeywords []string `json:"blacklist_keywords" yaml:"blacklist_keywords"`

	// BlacklistPatterns 黑名单正则模式
	BlacklistPatterns []string `json:"blacklist_patterns" yaml:"blacklist_patterns"`

	// EnableBlacklistCheck 是否启用黑名单检查
	EnableBlacklistCheck bool `json:"enable_blacklist_check" yaml:"enable_blacklist_check"`

	// Timeout 评估超时时间（秒）
	Timeout int `json:"timeout" yaml:"timeout"`

	// CustomFunctionConfig 自定义函数配置
	CustomFunctionConfig map[string]interface{} `json:"custom_function_config" yaml:"custom_function_config"`
}
