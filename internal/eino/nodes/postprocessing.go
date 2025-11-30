// Package nodes 提供 Eino Graph 中使用的 Lambda 节点实现
package nodes

import (
	"context"

	"github.com/cloudwego/eino/schema"
)

// CacheQueryOutput 缓存查询输出
type CacheQueryOutput struct {
	Hit      bool           `json:"hit"`
	Question string         `json:"question,omitempty"`
	Answer   string         `json:"answer,omitempty"`
	Score    float64        `json:"score,omitempty"`
	CacheID  string         `json:"cache_id,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// PostprocessResult 后处理结果 Lambda 函数
// 将 schema.Document 转换为 CacheQueryOutput
func PostprocessResult(ctx context.Context, doc *schema.Document) (*CacheQueryOutput, error) {
	if doc == nil {
		return &CacheQueryOutput{
			Hit: false,
		}, nil
	}

	output := &CacheQueryOutput{
		Hit:      true,
		CacheID:  doc.ID,
		Metadata: doc.MetaData,
	}

	// 从 MetaData 提取问答
	if question, ok := doc.MetaData["question"].(string); ok {
		output.Question = question
	}
	if answer, ok := doc.MetaData["answer"].(string); ok {
		output.Answer = answer
	}
	if score, ok := doc.MetaData["score"].(float64); ok {
		output.Score = score
	}

	return output, nil
}

// PostprocessResults 后处理多个结果
// 将 []*schema.Document 转换为 []*CacheQueryOutput
func PostprocessResults(ctx context.Context, docs []*schema.Document) ([]*CacheQueryOutput, error) {
	if len(docs) == 0 {
		return []*CacheQueryOutput{}, nil
	}

	outputs := make([]*CacheQueryOutput, 0, len(docs))
	for _, doc := range docs {
		output, err := PostprocessResult(ctx, doc)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, output)
	}

	return outputs, nil
}

// FormatCacheResult 格式化缓存结果
// 将查询结果格式化为用户友好的格式
func FormatCacheResult(ctx context.Context, output *CacheQueryOutput) (*CacheQueryOutput, error) {
	if output == nil || !output.Hit {
		return output, nil
	}

	// 可以在这里添加额外的格式化逻辑
	// 例如：截断过长的答案、添加引用标记等

	return output, nil
}

// ExtractAnswer 从文档中提取答案
func ExtractAnswer(doc *schema.Document) string {
	if doc == nil {
		return ""
	}

	if answer, ok := doc.MetaData["answer"].(string); ok {
		return answer
	}

	// 回退到 Content
	return doc.Content
}

// ExtractQuestion 从文档中提取问题
func ExtractQuestion(doc *schema.Document) string {
	if doc == nil {
		return ""
	}

	if question, ok := doc.MetaData["question"].(string); ok {
		return question
	}

	return ""
}

// ExtractScore 从文档中提取分数
func ExtractScore(doc *schema.Document) float64 {
	if doc == nil {
		return 0
	}

	if score, ok := doc.MetaData["score"].(float64); ok {
		return score
	}

	return 0
}
