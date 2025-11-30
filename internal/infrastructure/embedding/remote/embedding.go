package remote

import (
	"context"
	"fmt"
	"time"

	"github.com/openai/openai-go"

	"llm-cache/configs"
	"llm-cache/internal/domain/models"
	"llm-cache/pkg/logger"
)

// RemoteEmbeddingService 远程嵌入模型服务实现
// 基于OpenAI Format API实现文本向量化功能
type RemoteEmbeddingService struct {
	client openai.Client
	config *configs.RemoteEmbedding
	logger logger.Logger
}

// GenerateEmbedding 生成单个文本的向量
func (s *RemoteEmbeddingService) GenerateEmbedding(ctx context.Context, request *models.VectorProcessingRequest) (*models.VectorProcessingResult, error) {
	startTime := time.Now()

	s.logger.DebugContext(ctx, "开始生成单个文本向量",
		"text_length", len(request.Text),
		"model", s.getModelName(request.ModelName))

	// 构建OpenAI嵌入请求
	embeddingParams := openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfString: openai.String(request.Text),
		},
		Model: s.getModelName(request.ModelName),
	}

	// 调用OpenAI API
	response, err := s.client.Embeddings.New(ctx, embeddingParams)
	if err != nil {
		processingTime := float64(time.Since(startTime).Nanoseconds()) / 1e6
		errorMsg := fmt.Sprintf("failed to generate embedding: %v", err)

		s.logger.ErrorContext(ctx, "向量生成失败",
			"error", err,
			"processing_time_ms", processingTime)

		return &models.VectorProcessingResult{
			ProcessingTime: processingTime,
			ModelUsed:      s.getModelName(request.ModelName),
			Success:        false,
			Error:          errorMsg,
		}, fmt.Errorf("failed to generate embedding: %w", err)
	}

	// 验证响应
	if len(response.Data) == 0 {
		processingTime := float64(time.Since(startTime).Nanoseconds()) / 1e6
		errorMsg := "no embedding data in response"

		s.logger.ErrorContext(ctx, "响应数据为空",
			"processing_time_ms", processingTime)

		return &models.VectorProcessingResult{
			ProcessingTime: processingTime,
			ModelUsed:      response.Model,
			Success:        false,
			Error:          errorMsg,
		}, fmt.Errorf(errorMsg)
	}

	// 提取向量数据
	embedding := response.Data[0]
	vectorValues := make([]float32, len(embedding.Embedding))
	for i, val := range embedding.Embedding {
		vectorValues[i] = float32(val)
	}

	// 创建向量对象
	now := time.Now()
	vector := &models.Vector{
		ID:         fmt.Sprintf("embedding_%d", now.UnixNano()),
		Values:     vectorValues,
		Dimension:  len(vectorValues),
		CreateTime: now,
		UpdateTime: now,
		Normalized: false,
		ModelName:  response.Model,
	}

	// 如果需要归一化
	if request.Normalize {
		vector.Normalize()
	}

	processingTime := float64(time.Since(startTime).Nanoseconds()) / 1e6
	tokenCount := int(response.Usage.PromptTokens)

	s.logger.InfoContext(ctx, "向量生成成功",
		"dimension", len(vectorValues),
		"token_count", tokenCount,
		"processing_time_ms", processingTime,
		"model_used", response.Model)

	return &models.VectorProcessingResult{
		Vector:         vector,
		ProcessingTime: processingTime,
		TokenCount:     tokenCount,
		ModelUsed:      response.Model,
		Success:        true,
	}, nil
}

// GenerateBatchEmbeddings 批量生成文本向量
func (s *RemoteEmbeddingService) GenerateBatchEmbeddings(ctx context.Context, requests []*models.VectorProcessingRequest) ([]*models.VectorProcessingResult, error) {

	s.logger.InfoContext(ctx, "开始批量生成向量",
		"batch_size", len(requests))

	// 提取所有文本
	texts := make([]string, len(requests))
	for i, req := range requests {
		if req == nil || req.Text == "" {
			return nil, fmt.Errorf("request at index %d is nil or has empty text", i)
		}
		texts[i] = req.Text
	}

	startTime := time.Now()

	// 构建批量嵌入请求
	embeddingParams := openai.EmbeddingNewParams{
		Input: openai.EmbeddingNewParamsInputUnion{
			OfArrayOfStrings: texts,
		},
		Model: openai.EmbeddingModel(s.getModelName(requests[0].ModelName)),
	}

	// 设置编码格式为float
	embeddingParams.EncodingFormat = openai.EmbeddingNewParamsEncodingFormatFloat

	// 调用OpenAI API
	response, err := s.client.Embeddings.New(ctx, embeddingParams)
	if err != nil {
		processingTime := float64(time.Since(startTime).Nanoseconds()) / 1e6
		errorMsg := fmt.Sprintf("failed to generate batch embeddings: %v", err)

		s.logger.ErrorContext(ctx, "批量向量生成失败",
			"error", err,
			"batch_size", len(requests),
			"processing_time_ms", processingTime)

		// 返回所有失败的结果
		results := make([]*models.VectorProcessingResult, len(requests))
		for i := range requests {
			results[i] = &models.VectorProcessingResult{
				ProcessingTime: processingTime / float64(len(requests)),
				ModelUsed:      s.getModelName(requests[i].ModelName),
				Success:        false,
				Error:          errorMsg,
			}
		}
		return results, fmt.Errorf(errorMsg)
	}

	// 验证响应数量
	if len(response.Data) != len(requests) {
		return nil, fmt.Errorf("response data count (%d) doesn't match request count (%d)",
			len(response.Data), len(requests))
	}

	processingTime := float64(time.Since(startTime).Nanoseconds()) / 1e6
	totalTokens := int(response.Usage.PromptTokens)
	avgTokensPerRequest := totalTokens / len(requests)

	// 构建结果
	results := make([]*models.VectorProcessingResult, len(requests))
	for i, req := range requests {
		embedding := response.Data[i]

		// 转换向量数据
		vectorValues := make([]float32, len(embedding.Embedding))
		for j, val := range embedding.Embedding {
			vectorValues[j] = float32(val)
		}

		// 创建向量对象
		now := time.Now()
		vector := &models.Vector{
			ID:         fmt.Sprintf("embedding_%d_%d", now.UnixNano(), i),
			Values:     vectorValues,
			Dimension:  len(vectorValues),
			CreateTime: now,
			UpdateTime: now,
			Normalized: false,
			ModelName:  response.Model,
		}

		// 如果需要归一化
		if req.Normalize {
			vector.Normalize()
		}

		results[i] = &models.VectorProcessingResult{
			Vector:         vector,
			ProcessingTime: processingTime / float64(len(requests)),
			TokenCount:     avgTokensPerRequest,
			ModelUsed:      response.Model,
			Success:        true,
		}
	}

	s.logger.InfoContext(ctx, "批量向量生成成功",
		"batch_size", len(requests),
		"total_tokens", totalTokens,
		"processing_time_ms", processingTime,
		"model_used", response.Model)

	return results, nil
}

// getModelName 获取使用的模型名称，支持请求级别的模型覆盖
func (s *RemoteEmbeddingService) getModelName(requestModel string) string {
	if requestModel != "" {
		return requestModel
	}
	return s.config.ModelName
}

// Close 关闭服务，清理资源
func (s *RemoteEmbeddingService) Close() error {
	s.logger.Info("远程嵌入模型服务关闭")
	// OpenAI客户端不需要显式关闭
	return nil
}
