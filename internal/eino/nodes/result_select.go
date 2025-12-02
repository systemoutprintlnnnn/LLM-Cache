// Package nodes 提供 Eino Graph 中使用的 Lambda 节点实现
package nodes

import (
	"context"
	"math"
	"math/rand"

	"github.com/cloudwego/eino/schema"
)

// SelectionStrategy 定义结果选择策略类型。
type SelectionStrategy string

const (
	// StrategyFirst 策略：选择结果列表中的第一个文档。
	StrategyFirst SelectionStrategy = "first"
	// StrategyHighestScore 策略：选择分数最高的文档。
	StrategyHighestScore SelectionStrategy = "highest_score"
	// StrategyTemperatureSoftmax 策略：基于分数的 Softmax 概率进行随机采样。
	StrategyTemperatureSoftmax SelectionStrategy = "temperature_softmax"
)

// ResultSelector 实现搜索结果选择器。
// 支持多种选择策略（首选、最高分、Softmax 采样），用于从多个候选中选出最佳结果。
type ResultSelector struct {
	strategy    SelectionStrategy
	temperature float64
}

// NewResultSelector 创建一个新的结果选择器。
// 参数 strategy: 选择策略 ("first", "highest_score", "temperature_softmax")。
// 参数 temperature: 用于 Softmax 采样的温度系数（默认 0.7）。
// 返回: ResultSelector 指针。
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

// Select 根据配置的策略从候选文档列表中选择一个最佳文档。
// 参数 ctx: 上下文对象。
// 参数 docs: 候选文档列表。
// 返回: 选中的文档或 nil（如果列表为空）。
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

// selectHighestScore 内部方法：选择最高分结果。
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

// selectBySoftmax 内部方法：使用温度 Softmax 采样选择结果。
// 温度越高，随机性越大；温度越低，越倾向于选择高分结果。
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

// getDocScore 从文档元数据中提取分数。
// 优先查找 "score" 字段，其次是 "_score"。
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

// SelectFirstResult 便捷函数：选择第一个结果。
// 适用于简单的选择场景。
func SelectFirstResult(ctx context.Context, docs []*schema.Document) (*schema.Document, error) {
	if len(docs) == 0 {
		return nil, nil
	}
	return docs[0], nil
}

// SelectHighestScoreResult 便捷函数：选择最高分结果。
// 使用默认的最高分策略。
func SelectHighestScoreResult(ctx context.Context, docs []*schema.Document) (*schema.Document, error) {
	selector := NewResultSelector(string(StrategyHighestScore), 0.7)
	return selector.Select(ctx, docs)
}

// CreateSelectFunc 创建一个符合 Eino 接口签名的选择函数。
// 闭包捕获了选择策略配置。
func CreateSelectFunc(strategy string, temperature float64) func(context.Context, []*schema.Document) (*schema.Document, error) {
	selector := NewResultSelector(strategy, temperature)
	return selector.Select
}
