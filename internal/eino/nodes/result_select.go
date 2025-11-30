// Package nodes 提供 Eino Graph 中使用的 Lambda 节点实现
package nodes

import (
	"context"
	"math"
	"math/rand"

	"github.com/cloudwego/eino/schema"
)

// SelectionStrategy 选择策略类型
type SelectionStrategy string

const (
	// StrategyFirst 选择第一个结果
	StrategyFirst SelectionStrategy = "first"
	// StrategyHighestScore 选择最高分结果
	StrategyHighestScore SelectionStrategy = "highest_score"
	// StrategyTemperatureSoftmax 使用温度采样
	StrategyTemperatureSoftmax SelectionStrategy = "temperature_softmax"
)

// ResultSelector 结果选择器
type ResultSelector struct {
	strategy    SelectionStrategy
	temperature float64
}

// NewResultSelector 创建结果选择器
func NewResultSelector(strategy string, temperature float64) *ResultSelector {
	s := SelectionStrategy(strategy)
	if s == "" {
		s = StrategyHighestScore
	}
	if temperature <= 0 {
		temperature = 0.7
	}
	return &ResultSelector{
		strategy:    s,
		temperature: temperature,
	}
}

// Select 选择最佳结果 Lambda 函数
func (s *ResultSelector) Select(ctx context.Context, docs []*schema.Document) (*schema.Document, error) {
	if len(docs) == 0 {
		return nil, nil
	}

	switch s.strategy {
	case StrategyFirst:
		return docs[0], nil

	case StrategyHighestScore:
		return s.selectHighestScore(docs), nil

	case StrategyTemperatureSoftmax:
		return s.selectBySoftmax(docs), nil

	default:
		return docs[0], nil
	}
}

// selectHighestScore 选择最高分结果
func (s *ResultSelector) selectHighestScore(docs []*schema.Document) *schema.Document {
	if len(docs) == 0 {
		return nil
	}

	best := docs[0]
	bestScore := getDocScore(best)

	for _, doc := range docs[1:] {
		score := getDocScore(doc)
		if score > bestScore {
			best = doc
			bestScore = score
		}
	}
	return best
}

// selectBySoftmax 使用温度 Softmax 采样选择结果
func (s *ResultSelector) selectBySoftmax(docs []*schema.Document) *schema.Document {
	if len(docs) == 0 {
		return nil
	}
	if len(docs) == 1 {
		return docs[0]
	}

	scores := make([]float64, len(docs))
	maxScore := -math.MaxFloat64

	// 获取所有分数并找到最大值
	for i, doc := range docs {
		scores[i] = getDocScore(doc)
		if scores[i] > maxScore {
			maxScore = scores[i]
		}
	}

	// 计算 softmax 概率（使用数值稳定版本）
	expSum := 0.0
	for i := range scores {
		scores[i] = math.Exp((scores[i] - maxScore) / s.temperature)
		expSum += scores[i]
	}

	// 归一化
	for i := range scores {
		scores[i] /= expSum
	}

	// 采样
	r := rand.Float64()
	cumSum := 0.0
	for i, p := range scores {
		cumSum += p
		if r <= cumSum {
			return docs[i]
		}
	}

	return docs[len(docs)-1]
}

// getDocScore 从文档中获取分数
func getDocScore(doc *schema.Document) float64 {
	if doc == nil {
		return 0
	}

	// 尝试从 MetaData 中获取 score
	if score, ok := doc.MetaData["score"].(float64); ok {
		return score
	}

	// 尝试从 MetaData 中获取 _score
	if score, ok := doc.MetaData["_score"].(float64); ok {
		return score
	}

	return 0
}

// SelectFirstResult 选择第一个结果的便捷函数
func SelectFirstResult(ctx context.Context, docs []*schema.Document) (*schema.Document, error) {
	if len(docs) == 0 {
		return nil, nil
	}
	return docs[0], nil
}

// SelectHighestScoreResult 选择最高分结果的便捷函数
func SelectHighestScoreResult(ctx context.Context, docs []*schema.Document) (*schema.Document, error) {
	selector := NewResultSelector(string(StrategyHighestScore), 0.7)
	return selector.Select(ctx, docs)
}

// CreateSelectFunc 创建选择函数
func CreateSelectFunc(strategy string, temperature float64) func(context.Context, []*schema.Document) (*schema.Document, error) {
	selector := NewResultSelector(strategy, temperature)
	return selector.Select
}
