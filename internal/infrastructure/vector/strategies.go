package vector

import (
	"context"
	"fmt"
	"llm-cache/internal/domain/models"
	"llm-cache/pkg/logger"
	"math"
	"math/rand"
	"time"
)

// TemperatureSoftmaxConfig 温度softmax策略配置
type TemperatureSoftmaxConfig struct {
	// Temperature 温度参数，控制选择的随机性
	Temperature float64 `json:"temperature" yaml:"temperature"`

	// TopK 考虑的前K个结果
	TopK int `json:"top_k" yaml:"top_k"`

	// MinProbability 最小概率阈值
	MinProbability float64 `json:"min_probability" yaml:"min_probability"`
}

// DefaultTemperatureSoftmaxConfig 默认温度softmax配置
func DefaultTemperatureSoftmaxConfig() *TemperatureSoftmaxConfig {
	return &TemperatureSoftmaxConfig{
		Temperature:    1.0,
		TopK:           5,
		MinProbability: 0.1,
	}
}

// ResultSelectionStrategy 结果选择策略接口
type ResultSelectionStrategy interface {
	// Name 策略名称
	Name() string

	// Select 选择结果
	Select(ctx context.Context, results []*models.VectorSearchResult, query *models.CacheQuery, config map[string]interface{}) (*models.VectorSearchResult, error)

	// Validate 验证策略配置
	Validate(config map[string]interface{}) (bool, error)
}

// FirstSelectionStrategy 选择第一个结果的策略
type FirstSelectionStrategy struct {
	logger logger.Logger
}

// NewFirstSelectionStrategy 创建第一个结果选择策略
func NewFirstSelectionStrategy(log logger.Logger) *FirstSelectionStrategy {
	return &FirstSelectionStrategy{
		logger: log,
	}
}

func (s *FirstSelectionStrategy) Name() string {
	return "first"
}

func (s *FirstSelectionStrategy) Select(ctx context.Context, results []*models.VectorSearchResult, query *models.CacheQuery, config map[string]interface{}) (*models.VectorSearchResult, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no results to select from")
	}

	selected := results[0]
	s.logger.DebugContext(ctx, "Selected first result",
		"result_id", selected.ID,
		"score", selected.Score,
		"strategy", s.Name())

	return selected, nil
}

func (s *FirstSelectionStrategy) Validate(config map[string]interface{}) (bool, error) {
	// 第一个结果策略不需要特殊配置
	return true, nil
}

// HighestScoreSelectionStrategy 选择分数最高结果的策略
type HighestScoreSelectionStrategy struct {
	logger logger.Logger
}

// NewHighestScoreSelectionStrategy 创建最高分数选择策略
func NewHighestScoreSelectionStrategy(log logger.Logger) *HighestScoreSelectionStrategy {
	return &HighestScoreSelectionStrategy{
		logger: log,
	}
}

func (s *HighestScoreSelectionStrategy) Name() string {
	return "highest_score"
}

func (s *HighestScoreSelectionStrategy) Select(ctx context.Context, results []*models.VectorSearchResult, query *models.CacheQuery, config map[string]interface{}) (*models.VectorSearchResult, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no results to select from")
	}

	var bestResult *models.VectorSearchResult
	var bestScore float64 = -1

	for _, result := range results {
		if result.Score > bestScore {
			bestScore = result.Score
			bestResult = result
		}
	}

	if bestResult == nil {
		return nil, fmt.Errorf("no valid result found")
	}

	s.logger.DebugContext(ctx, "Selected highest score result",
		"result_id", bestResult.ID,
		"score", bestResult.Score,
		"strategy", s.Name())

	return bestResult, nil
}

func (s *HighestScoreSelectionStrategy) Validate(config map[string]interface{}) (bool, error) {
	// 最高分数策略不需要特殊配置
	return true, nil
}

// TemperatureSoftmaxSelectionStrategy 基于温度的softmax选择策略
type TemperatureSoftmaxSelectionStrategy struct {
	logger logger.Logger
	config *TemperatureSoftmaxConfig
}

// NewTemperatureSoftmaxSelectionStrategy 创建温度softmax选择策略
func NewTemperatureSoftmaxSelectionStrategy(log logger.Logger, config *TemperatureSoftmaxConfig) *TemperatureSoftmaxSelectionStrategy {
	return &TemperatureSoftmaxSelectionStrategy{
		logger: log,
		config: config,
	}
}

func (s *TemperatureSoftmaxSelectionStrategy) Name() string {
	return "temperature_softmax"
}

func (s *TemperatureSoftmaxSelectionStrategy) Select(ctx context.Context, results []*models.VectorSearchResult, query *models.CacheQuery, config map[string]interface{}) (*models.VectorSearchResult, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no results to select from")
	}

	// 获取温度参数
	temperature := s.config.Temperature
	if tempValue, ok := config["temperature"]; ok {
		if tempFloat, ok := tempValue.(float64); ok {
			temperature = tempFloat
		}
	}

	// 限制考虑的结果数量
	topK := s.config.TopK
	if topKValue, ok := config["top_k"]; ok {
		if topKInt, ok := topKValue.(int); ok {
			topK = topKInt
		}
	}

	// 取前TopK个结果
	candidates := results
	if len(candidates) > topK {
		candidates = candidates[:topK]
	}

	// 计算softmax概率
	probabilities := s.calculateSoftmaxProbabilities(candidates, temperature)

	// 基于概率进行随机选择
	selectedIndex := s.weightedRandomSelect(probabilities)
	if selectedIndex >= len(candidates) {
		return nil, fmt.Errorf("invalid selection index: %d", selectedIndex)
	}

	selected := candidates[selectedIndex]
	s.logger.DebugContext(ctx, "Selected result using temperature softmax",
		"result_id", selected.ID,
		"score", selected.Score,
		"probability", probabilities[selectedIndex],
		"temperature", temperature,
		"strategy", s.Name())

	return selected, nil
}

func (s *TemperatureSoftmaxSelectionStrategy) Validate(config map[string]interface{}) (bool, error) {
	if temp, ok := config["temperature"]; ok {
		if tempFloat, ok := temp.(float64); ok {
			if tempFloat <= 0 {
				return false, fmt.Errorf("temperature must be positive")
			}
		} else {
			return false, fmt.Errorf("temperature must be a float64")
		}
	}

	if topK, ok := config["top_k"]; ok {
		if topKInt, ok := topK.(int); ok {
			if topKInt <= 0 {
				return false, fmt.Errorf("top_k must be positive")
			}
		} else {
			return false, fmt.Errorf("top_k must be an integer")
		}
	}

	return true, nil
}

// calculateSoftmaxProbabilities 计算softmax概率分布
func (s *TemperatureSoftmaxSelectionStrategy) calculateSoftmaxProbabilities(results []*models.VectorSearchResult, temperature float64) []float64 {
	if len(results) == 0 {
		return []float64{}
	}

	// 计算指数值
	expValues := make([]float64, len(results))
	var sumExp float64

	for i, result := range results {
		expValues[i] = math.Exp(result.Score / temperature)
		sumExp += expValues[i]
	}

	// 归一化为概率
	probabilities := make([]float64, len(results))
	for i := range expValues {
		probabilities[i] = expValues[i] / sumExp
	}

	return probabilities
}

// weightedRandomSelect 基于权重进行随机选择
func (s *TemperatureSoftmaxSelectionStrategy) weightedRandomSelect(probabilities []float64) int {
	if len(probabilities) == 0 {
		return 0
	}

	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	// 生成随机数
	r := rand.Float64()

	// 累积概率选择
	cumulative := 0.0
	for i, prob := range probabilities {
		cumulative += prob
		if r <= cumulative {
			return i
		}
	}

	// 如果由于浮点精度问题没有选中，返回最后一个
	return len(probabilities) - 1
}

// StrategyFactory 策略工厂
type StrategyFactory struct {
	logger                   logger.Logger
	temperatureSoftmaxConfig *TemperatureSoftmaxConfig
}

// NewStrategyFactory 创建策略工厂
func NewStrategyFactory(log logger.Logger, temperatureSoftmaxConfig *TemperatureSoftmaxConfig) *StrategyFactory {
	if temperatureSoftmaxConfig == nil {
		temperatureSoftmaxConfig = DefaultTemperatureSoftmaxConfig()
	}
	return &StrategyFactory{
		logger:                   log,
		temperatureSoftmaxConfig: temperatureSoftmaxConfig,
	}
}

// CreateStrategy 创建策略实例
func (f *StrategyFactory) CreateStrategy(strategyName string) (ResultSelectionStrategy, error) {
	switch strategyName {
	case "first":
		return NewFirstSelectionStrategy(f.logger), nil
	case "highest_score":
		return NewHighestScoreSelectionStrategy(f.logger), nil
	case "temperature_softmax":
		return NewTemperatureSoftmaxSelectionStrategy(f.logger, f.temperatureSoftmaxConfig), nil
	default:
		f.logger.ErrorContext(context.Background(), "Unknown selection strategy",
			"strategy", strategyName)
		return nil, fmt.Errorf("unknown selection strategy: %s", strategyName)
	}
}
