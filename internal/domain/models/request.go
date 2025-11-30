package models

// PreprocessedRequest 预处理后的请求
type PreprocessedRequest struct {
	// Original 原始请求
	Original *CacheQuery `json:"original,omitempty"`

	// ProcessedQuestion 预处理后的问题文本
	ProcessedQuestion string `json:"processed_question" validate:"required"`

	// ProcessingTime 预处理耗时（毫秒）
	ProcessingTime float64 `json:"processing_time" validate:"min=0"`

	// Success 预处理是否成功
	Success bool `json:"success"`

	// Error 预处理错误信息
	Error string `json:"error,omitempty"`
}

// VectorProcessingRequest 向量处理请求
type VectorProcessingRequest struct {
	// Text 待向量化的文本
	Text string `json:"text" validate:"required"`

	// ModelName 指定使用的向量化模型
	ModelName string `json:"model_name,omitempty"`

	// Normalize 是否归一化向量
	Normalize bool `json:"normalize,omitempty"`
}

// VectorProcessingResult 向量处理结果
type VectorProcessingResult struct {
	// Vector 生成的向量
	Vector *Vector `json:"vector,omitempty"`

	// ProcessingTime 处理耗时（毫秒）
	ProcessingTime float64 `json:"processing_time" validate:"min=0"`

	// TokenCount 实际处理的token数量
	TokenCount int `json:"token_count" validate:"min=0"`

	// ModelUsed 实际使用的模型
	ModelUsed string `json:"model_used" validate:"required"`

	// Success 是否成功
	Success bool `json:"success"`

	// Error 错误信息
	Error string `json:"error,omitempty"`
}

// QualityAssessmentRequest 质量评估请求
type QualityAssessmentRequest struct {
	// Question 问题文本
	Question string `json:"question" validate:"required"`

	// Answer 答案文本
	Answer string `json:"answer" validate:"required"`
}

// QualityAssessmentResult 质量评估结果
type QualityAssessmentResult struct {
	// Passed 是否通过评估
	Passed bool `json:"passed"`

	// Score 综合质量分数 (-1.0 或 0.0-1.0，其中 -1.0 表示无自定义质量函数可用)
	Score float64 `json:"score" validate:"min=-1.0,max=1.0"`

	// Threshold 使用的阈值
	Threshold float64 `json:"threshold" validate:"min=-2.0,max=1.0"`

	// AssessmentTime 评估耗时（毫秒）
	AssessmentTime float64 `json:"assessment_time" validate:"min=0"`

	// Reason 失败原因（如果未通过）
	Reason string `json:"reason,omitempty"`
}
