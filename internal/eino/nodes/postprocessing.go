// Package nodes 提供 Eino Graph 中使用的 Lambda 节点实现
package nodes

import (
	"context"

	"github.com/cloudwego/eino/schema"
)

// CacheQueryOutput 定义缓存查询的最终输出结构。
// 包含命中状态、问答对、相似度分数、缓存 ID 以及元数据。
type CacheQueryOutput struct {
	Hit      bool           `json:"hit"`
	Question string         `json:"question,omitempty"`
	Answer   string         `json:"answer,omitempty"`
	Score    float64        `json:"score,omitempty"`
	CacheID  string         `json:"cache_id,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// PostprocessResult 后处理结果 Lambda 函数。
// 将 Eino 检索到的 schema.Document 对象转换为业务层使用的 CacheQueryOutput 结构。
// 参数 ctx: 上下文对象。
// 参数 doc: 检索到的文档对象。
// 返回: 转换后的查询输出对象或错误。
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

// PostprocessResults 批量处理多个检索结果。
// 将文档列表转换为查询输出列表。
// 参数 ctx: 上下文对象。
// 参数 docs: 文档对象切片。
// 返回: 查询输出对象切片或错误。
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

// FormatCacheResult 对缓存查询结果进行最终格式化。
// 该函数可用于添加额外的业务逻辑，如截断过长的文本或添加特定的标记。
// 参数 ctx: 上下文对象。
// 参数 output: 初步处理后的查询输出。
// 返回: 格式化后的查询输出或错误。
func FormatCacheResult(ctx context.Context, output *CacheQueryOutput) (*CacheQueryOutput, error) {
	if output == nil || !output.Hit {
		return output, nil
	}

	// 可以在这里添加额外的格式化逻辑
	// 例如：截断过长的答案、添加引用标记等

	return output, nil
}

// ExtractAnswer 辅助函数：从文档对象中提取答案文本。
// 优先尝试从元数据中获取 "answer" 字段，如果不存在则回退到 Content 字段。
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

// ExtractQuestion 辅助函数：从文档对象中提取问题文本。
// 尝试从元数据中获取 "question" 字段。
func ExtractQuestion(doc *schema.Document) string {
	if doc == nil {
		return ""
	}

	if question, ok := doc.MetaData["question"].(string); ok {
		return question
	}

	return ""
}

// ExtractScore 辅助函数：从文档对象中提取相关性分数。
// 尝试从元数据中获取 "score" 字段。
func ExtractScore(doc *schema.Document) float64 {
	if doc == nil {
		return 0
	}

	if score, ok := doc.MetaData["score"].(float64); ok {
		return score
	}

	return 0
}
