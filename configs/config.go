package configs

import (
	"fmt"
	"time"

	einoconfig "llm-cache/internal/eino/config"
)

// Config 主配置结构体，定义了应用程序的所有配置项。
// 包含服务器、数据库、嵌入模型、日志、缓存和质量评估等模块的配置信息。
type Config struct {
	Server    ServerConfig          `yaml:"server"`
	Database  DatabaseConfig        `yaml:"database"`
	Embedding EmbeddingConfig       `yaml:"embedding"`
	Logging   LoggingConfig         `yaml:"logging"`
	Cache     CacheConfig           `yaml:"cache"`
	Quality   QualityConfig         `yaml:"quality"`
	Eino      einoconfig.EinoConfig `yaml:"eino"` // Eino 框架配置
}

// ServerConfig 定义服务器相关的配置参数。
// 包含监听地址、端口、超时设置和连接限制等。
type ServerConfig struct {
	Host                    string        `yaml:"host"`
	Port                    int           `yaml:"port"`
	ReadTimeout             time.Duration `yaml:"read_timeout"`
	WriteTimeout            time.Duration `yaml:"write_timeout"`
	IdleTimeout             time.Duration `yaml:"idle_timeout"`
	GracefulShutdownTimeout time.Duration `yaml:"graceful_shutdown_timeout"`
	MaxConnections          int           `yaml:"max_connections"`
}

// DatabaseConfig 定义向量数据库的配置参数。
// 支持多种类型的向量数据库（目前主要支持 qdrant）。
type DatabaseConfig struct {
	Type   string       `yaml:"type"`
	Qdrant QdrantConfig `yaml:"qdrant"`
}

// QdrantConfig 定义 Qdrant 向量数据库的具体配置。
// 包含连接信息、集合名称、向量维度和距离度量方式等。
type QdrantConfig struct {
	Host           string        `yaml:"host"`
	Port           int           `yaml:"port"`
	APIKey         string        `yaml:"api_key"`
	CollectionName string        `yaml:"collection_name"`
	VectorSize     int           `yaml:"vector_size"`
	Distance       string        `yaml:"distance"`
	Timeout        time.Duration `yaml:"timeout"`
	MaxRetries     int           `yaml:"max_retries"`
	RetryDelay     time.Duration `yaml:"retry_delay"`
}

// EmbeddingConfig 定义嵌入模型（Embedding）的配置参数。
// 支持本地模型和远程 API 调用两种模式。
type EmbeddingConfig struct {
	Type   string          `yaml:"type"`
	Local  LocalEmbedding  `yaml:"local"`
	Remote RemoteEmbedding `yaml:"remote"`
}

// LocalEmbedding 定义本地嵌入模型的配置。
// 包含模型文件路径、最大 token 数和批处理大小等。
type LocalEmbedding struct {
	ModelPath string `yaml:"model_path"`
	MaxTokens int    `yaml:"max_tokens"`
	BatchSize int    `yaml:"batch_size"`
}

// RemoteEmbedding 定义远程嵌入模型 API 的配置。
// 包含 API 端点、密钥、模型名称和重试策略等。
type RemoteEmbedding struct {
	APIEndpoint string            `yaml:"api_endpoint"`
	APIKey      string            `yaml:"api_key"`
	ModelName   string            `yaml:"model_name"`
	Timeout     time.Duration     `yaml:"timeout"`
	MaxRetries  int               `yaml:"max_retries"`
	RetryDelay  time.Duration     `yaml:"retry_delay"`
	Headers     map[string]string `yaml:"headers"`
}

// LoggingConfig 定义日志系统的配置参数。
// 包含日志级别、输出目标（stdout/file）、格式（text/json）和文件轮转设置等。
type LoggingConfig struct {
	Level      string `yaml:"level"`
	Output     string `yaml:"output"`
	FilePath   string `yaml:"file_path"`
	Format     string `yaml:"format"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}

// CacheConfig 定义语义缓存的核心配置参数。
// 包含相似度阈值、返回结果数量、过期时间和缓存大小限制等。
type CacheConfig struct {
	SimilarityThreshold float64       `yaml:"similarity_threshold"`
	TopK                int           `yaml:"top_k"`
	TTL                 time.Duration `yaml:"ttl"`
	MaxCacheSize        int64         `yaml:"max_cache_size"`
	EnableAsyncUpdate   bool          `yaml:"enable_async_update"`
	UpdateBatchSize     int           `yaml:"update_batch_size"`
	UpdateInterval      time.Duration `yaml:"update_interval"`
}

// QualityConfig 定义回答质量评估的配置参数。
// 包含开关、阈值、评估策略和黑名单规则等。
type QualityConfig struct {
	Enabled    bool              `yaml:"enabled"`
	Threshold  float64           `yaml:"threshold"`
	Strategies []QualityStrategy `yaml:"strategies"`
	Blacklist  QualityBlacklist  `yaml:"blacklist"`
}

// QualityStrategy 定义具体的质量评估策略。
// 包含策略名称、权重和特定的配置参数。
type QualityStrategy struct {
	Name    string                 `yaml:"name"`
	Weight  float64                `yaml:"weight"`
	Enabled bool                   `yaml:"enabled"`
	Config  map[string]interface{} `yaml:"config"`
}

// QualityBlacklist 定义质量评估的黑名单规则。
// 包含关键词过滤和文本长度限制等条件。
type QualityBlacklist struct {
	ApologyKeywords   []string `yaml:"apology_keywords"`
	ErrorKeywords     []string `yaml:"error_keywords"`
	MinAnswerLength   int      `yaml:"min_answer_length"`
	MaxAnswerLength   int      `yaml:"max_answer_length"`
	MinQuestionLength int      `yaml:"min_question_length"`
	MaxQuestionLength int      `yaml:"max_question_length"`
}

// Validate 检查 Config 配置结构体的有效性。
// 依次调用各个子配置项的 Validate 方法，如果发现无效配置，返回相应的错误。
func (c *Config) Validate() error {
	if err := c.Server.Validate(); err != nil {
		return fmt.Errorf("server config validation failed: %w", err)
	}

	if err := c.Database.Validate(); err != nil {
		return fmt.Errorf("database config validation failed: %w", err)
	}

	if err := c.Embedding.Validate(); err != nil {
		return fmt.Errorf("embedding config validation failed: %w", err)
	}

	if err := c.Logging.Validate(); err != nil {
		return fmt.Errorf("logging config validation failed: %w", err)
	}

	if err := c.Cache.Validate(); err != nil {
		return fmt.Errorf("cache config validation failed: %w", err)
	}

	if err := c.Quality.Validate(); err != nil {
		return fmt.Errorf("quality config validation failed: %w", err)
	}

	return nil
}

// Validate 检查 ServerConfig 配置的有效性。
// 确保端口号在有效范围内，且超时设置和最大连接数为正数。
func (s *ServerConfig) Validate() error {
	if s.Port <= 0 || s.Port > 65535 {
		return fmt.Errorf("invalid port: %d", s.Port)
	}

	if s.ReadTimeout <= 0 {
		return fmt.Errorf("read_timeout must be positive")
	}

	if s.WriteTimeout <= 0 {
		return fmt.Errorf("write_timeout must be positive")
	}

	if s.MaxConnections <= 0 {
		return fmt.Errorf("max_connections must be positive")
	}

	return nil
}

// Validate 检查 DatabaseConfig 配置的有效性。
// 确保存储类型已指定且受支持（如 qdrant）。
func (d *DatabaseConfig) Validate() error {
	if d.Type == "" {
		return fmt.Errorf("database type is required")
	}

	if d.Type == "qdrant" {
		return d.Qdrant.Validate()
	}

	return fmt.Errorf("unsupported database type: %s", d.Type)
}

// Validate 检查 QdrantConfig 配置的有效性。
// 确保 Host、Port、CollectionName 和 VectorSize 等关键参数已正确设置。
func (q *QdrantConfig) Validate() error {
	if q.Host == "" {
		return fmt.Errorf("qdrant host is required")
	}

	if q.Port <= 0 || q.Port > 65535 {
		return fmt.Errorf("invalid qdrant port: %d", q.Port)
	}

	if q.CollectionName == "" {
		return fmt.Errorf("qdrant collection name is required")
	}

	if q.VectorSize <= 0 {
		return fmt.Errorf("vector size must be positive")
	}

	if q.Distance == "" {
		q.Distance = "cosine"
	}

	return nil
}

// Validate 检查 EmbeddingConfig 配置的有效性。
// 确保存储类型已指定，并调用相应的本地或远程配置验证方法。
func (e *EmbeddingConfig) Validate() error {
	if e.Type == "" {
		return fmt.Errorf("embedding type is required")
	}

	switch e.Type {
	case "local":
		return e.Local.Validate()
	case "remote":
		return e.Remote.Validate()
	default:
		return fmt.Errorf("unsupported embedding type: %s", e.Type)
	}
}

// Validate 检查 LocalEmbedding 配置的有效性。
// 确保模型路径已指定，并设置默认的批处理大小（如果未设置）。
func (l *LocalEmbedding) Validate() error {
	if l.ModelPath == "" {
		return fmt.Errorf("local embedding model path is required")
	}

	if l.BatchSize <= 0 {
		l.BatchSize = 32 // 默认批处理大小
	}

	return nil
}

// Validate 检查 RemoteEmbedding 配置的有效性。
// 确保 API 端点和模型名称已指定，并设置默认的超时时间（如果未设置）。
func (r *RemoteEmbedding) Validate() error {
	if r.APIEndpoint == "" {
		return fmt.Errorf("remote embedding API endpoint is required")
	}

	if r.ModelName == "" {
		return fmt.Errorf("remote embedding model name is required")
	}

	if r.Timeout <= 0 {
		r.Timeout = 30 * time.Second // 默认超时时间
	}

	return nil
}

// Validate 检查 LoggingConfig 配置的有效性。
// 确保日志级别、输出目标和格式有效，如果输出到文件，确保文件路径已指定。
func (l *LoggingConfig) Validate() error {
	validLevels := map[string]bool{
		"debug": true, "info": true, "warn": true, "error": true,
	}

	if !validLevels[l.Level] {
		return fmt.Errorf("invalid log level: %s", l.Level)
	}

	validOutputs := map[string]bool{
		"stdout": true, "stderr": true, "file": true,
	}

	if !validOutputs[l.Output] {
		return fmt.Errorf("invalid log output: %s", l.Output)
	}

	if l.Output == "file" && l.FilePath == "" {
		return fmt.Errorf("file path is required when output is file")
	}

	// 验证日志格式，空值默认为 text
	validFormats := map[string]bool{
		"text": true, "json": true, "": true,
	}

	if !validFormats[l.Format] {
		return fmt.Errorf("invalid log format: %s", l.Format)
	}

	return nil
}

// Validate 检查 CacheConfig 配置的有效性。
// 确保相似度阈值在 0-1 之间，TopK 和 TTL 为正数，且异步更新配置正确。
func (c *CacheConfig) Validate() error {
	if c.SimilarityThreshold < 0 || c.SimilarityThreshold > 1 {
		return fmt.Errorf("similarity threshold must be between 0 and 1")
	}

	if c.TopK <= 0 {
		return fmt.Errorf("top_k must be positive")
	}

	if c.TTL <= 0 {
		return fmt.Errorf("ttl must be positive")
	}

	if c.EnableAsyncUpdate && c.UpdateBatchSize <= 0 {
		return fmt.Errorf("update_batch_size must be positive when async update is enabled")
	}

	return nil
}

// Validate 检查 QualityConfig 配置的有效性。
// 如果启用，确保阈值在 0-1 之间，且启用策略的总权重为正数。
func (q *QualityConfig) Validate() error {
	if !q.Enabled {
		return nil
	}

	if q.Threshold < 0 || q.Threshold > 1 {
		return fmt.Errorf("quality threshold must be between 0 and 1")
	}

	// 只有当配置了策略时才检查权重
	if len(q.Strategies) > 0 {
		totalWeight := 0.0
		for _, strategy := range q.Strategies {
			if strategy.Enabled {
				totalWeight += strategy.Weight
			}
		}

		if totalWeight <= 0 {
			return fmt.Errorf("total weight of enabled strategies must be positive")
		}
	}

	return nil
}

// GetAddr 获取服务器的完整监听地址。
// 返回格式为 "Host:Port" 的字符串。
func (s *ServerConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// GetQdrantAddr 获取 Qdrant 服务的完整地址。
// 返回格式为 "Host:Port" 的字符串。
func (q *QdrantConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", q.Host, q.Port)
}
