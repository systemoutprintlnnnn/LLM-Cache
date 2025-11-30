// Package nodes 提供 Eino Graph 中使用的 Lambda 节点实现
package nodes

import (
	"context"
	"strings"

	"llm-cache/internal/eino/config"
)

// QualityCheckInput 质量检查输入
type QualityCheckInput struct {
	Question   string
	Answer     string
	UserType   string
	Metadata   map[string]any
	ForceWrite bool
}

// QualityCheckResult 质量检查结果
type QualityCheckResult struct {
	Passed   bool
	Reason   string
	Score    float64
	Question string
	Answer   string
	UserType string
	Metadata map[string]any
}

// QualityChecker 质量检查器
type QualityChecker struct {
	cfg *config.QualityConfig
}

// NewQualityChecker 创建质量检查器
func NewQualityChecker(cfg *config.QualityConfig) *QualityChecker {
	return &QualityChecker{cfg: cfg}
}

// Check 执行质量检查 Lambda 函数
func (c *QualityChecker) Check(ctx context.Context, input *QualityCheckInput) (*QualityCheckResult, error) {
	result := &QualityCheckResult{
		Question: input.Question,
		Answer:   input.Answer,
		UserType: input.UserType,
		Metadata: input.Metadata,
	}

	// 跳过质量检查（如果配置禁用或强制写入）
	if !c.cfg.Enabled || input.ForceWrite {
		result.Passed = true
		result.Score = 1.0
		return result, nil
	}

	// 1. 检查问题长度
	questionLen := len(strings.TrimSpace(input.Question))
	if questionLen < c.cfg.MinQuestionLength {
		result.Passed = false
		result.Reason = "question too short"
		result.Score = 0.0
		return result, nil
	}
	if c.cfg.MaxQuestionLength > 0 && questionLen > c.cfg.MaxQuestionLength {
		result.Passed = false
		result.Reason = "question too long"
		result.Score = 0.0
		return result, nil
	}

	// 2. 检查答案长度
	answerLen := len(strings.TrimSpace(input.Answer))
	if answerLen < c.cfg.MinAnswerLength {
		result.Passed = false
		result.Reason = "answer too short"
		result.Score = 0.0
		return result, nil
	}
	if c.cfg.MaxAnswerLength > 0 && answerLen > c.cfg.MaxAnswerLength {
		result.Passed = false
		result.Reason = "answer too long"
		result.Score = 0.0
		return result, nil
	}

	// 3. 检查黑名单
	if containsBlacklistWords(input.Question, c.cfg.BlacklistKeywords) ||
		containsBlacklistWords(input.Answer, c.cfg.BlacklistKeywords) {
		result.Passed = false
		result.Reason = "contains blacklisted content"
		result.Score = 0.0
		return result, nil
	}

	// 4. 计算质量分数
	score := calculateQualityScore(input.Question, input.Answer)
	if score < c.cfg.ScoreThreshold {
		result.Passed = false
		result.Reason = "quality score below threshold"
		result.Score = score
		return result, nil
	}

	result.Passed = true
	result.Score = score
	return result, nil
}

// containsBlacklistWords 检查是否包含黑名单关键词
func containsBlacklistWords(text string, blacklist []string) bool {
	if len(blacklist) == 0 {
		return false
	}

	lower := strings.ToLower(text)
	for _, word := range blacklist {
		if strings.Contains(lower, strings.ToLower(word)) {
			return true
		}
	}
	return false
}

// calculateQualityScore 计算质量分数
func calculateQualityScore(question, answer string) float64 {
	score := 1.0

	// 基于长度的评分
	questionLen := len(strings.TrimSpace(question))
	answerLen := len(strings.TrimSpace(answer))

	// 问题长度评分（10-200 字符最佳）
	if questionLen < 10 {
		score -= 0.2
	} else if questionLen > 500 {
		score -= 0.1
	}

	// 答案长度评分（50-2000 字符最佳）
	if answerLen < 50 {
		score -= 0.2
	} else if answerLen > 5000 {
		score -= 0.1
	}

	// 问题完整性检查（是否包含问号）
	if !strings.Contains(question, "?") && !strings.Contains(question, "？") {
		score -= 0.1
	}

	// 确保分数在 0-1 范围内
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	return score
}

// CheckDetail 单项检查详情
type CheckDetail struct {
	Name    string
	Passed  bool
	Score   float64
	Message string
}

// LengthCheck 长度检查
func LengthCheck(ctx context.Context, input *QualityCheckInput, cfg *config.QualityConfig) (*CheckDetail, error) {
	detail := &CheckDetail{Name: "length_check", Passed: true, Score: 1.0}

	// 检查问题长度
	questionLen := len(strings.TrimSpace(input.Question))
	if questionLen < cfg.MinQuestionLength {
		detail.Passed = false
		detail.Score = 0.0
		detail.Message = "question too short"
		return detail, nil
	}

	// 检查答案长度
	answerLen := len(strings.TrimSpace(input.Answer))
	if answerLen < cfg.MinAnswerLength {
		detail.Passed = false
		detail.Score = 0.0
		detail.Message = "answer too short"
		return detail, nil
	}

	// 计算长度分数
	qScore := min(float64(questionLen)/100.0, 1.0)
	aScore := min(float64(answerLen)/500.0, 1.0)
	detail.Score = (qScore + aScore) / 2.0

	return detail, nil
}

// BlacklistCheck 黑名单检查
func BlacklistCheck(ctx context.Context, input *QualityCheckInput, blacklist []string) (*CheckDetail, error) {
	detail := &CheckDetail{Name: "blacklist_check", Passed: true, Score: 1.0}

	if len(blacklist) == 0 {
		return detail, nil
	}

	text := strings.ToLower(input.Question + " " + input.Answer)
	for _, word := range blacklist {
		if strings.Contains(text, strings.ToLower(word)) {
			detail.Passed = false
			detail.Score = 0.0
			detail.Message = "contains blacklisted word: " + word
			return detail, nil
		}
	}

	return detail, nil
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
