package configs

import (
	"fmt"
	"time"

	einoconfig "llm-cache/internal/eino/config"
)

// Config 主配置结构体
type Config struct {
	Server    ServerConfig          `yaml:"server"`
	Database  DatabaseConfig        `yaml:"database"`
	Embedding EmbeddingConfig       `yaml:"embedding"`
	Logging   LoggingConfig         `yaml:"logging"`
	Cache     CacheConfig           `yaml:"cache"`
	Quality   QualityConfig         `yaml:"quality"`
	Eino      einoconfig.EinoConfig `yaml:"eino"` // Eino 框架配置
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host                    string        `yaml:"host"`
	Port                    int           `yaml:"port"`
	ReadTimeout             time.Duration `yaml:"read_timeout"`
	WriteTimeout            time.Duration `yaml:"write_timeout"`
	IdleTimeout             time.Duration `yaml:"idle_timeout"`
	GracefulShutdownTimeout time.Duration `yaml:"graceful_shutdown_timeout"`
	MaxConnections          int           `yaml:"max_connections"`
}

// DatabaseConfig 向量数据库配置
type DatabaseConfig struct {
	Type   string       `yaml:"type"`
	Qdrant QdrantConfig `yaml:"qdrant"`
}

// QdrantConfig Qdrant向量数据库配置
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

// EmbeddingConfig 嵌入模型配置
type EmbeddingConfig struct {
	Type   string          `yaml:"type"`
	Local  LocalEmbedding  `yaml:"local"`
	Remote RemoteEmbedding `yaml:"remote"`
}

// LocalEmbedding 本地嵌入模型配置
type LocalEmbedding struct {
	ModelPath string `yaml:"model_path"`
	MaxTokens int    `yaml:"max_tokens"`
	BatchSize int    `yaml:"batch_size"`
}

// RemoteEmbedding 远程嵌入模型配置
type RemoteEmbedding struct {
	APIEndpoint string            `yaml:"api_endpoint"`
	APIKey      string            `yaml:"api_key"`
	ModelName   string            `yaml:"model_name"`
	Timeout     time.Duration     `yaml:"timeout"`
	MaxRetries  int               `yaml:"max_retries"`
	RetryDelay  time.Duration     `yaml:"retry_delay"`
	Headers     map[string]string `yaml:"headers"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `yaml:"level"`
	Output     string `yaml:"output"`
	FilePath   string `yaml:"file_path"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	SimilarityThreshold float64       `yaml:"similarity_threshold"`
	TopK                int           `yaml:"top_k"`
	TTL                 time.Duration `yaml:"ttl"`
	MaxCacheSize        int64         `yaml:"max_cache_size"`
	EnableAsyncUpdate   bool          `yaml:"enable_async_update"`
	UpdateBatchSize     int           `yaml:"update_batch_size"`
	UpdateInterval      time.Duration `yaml:"update_interval"`
}

// QualityConfig 质量评估配置
type QualityConfig struct {
	Enabled    bool              `yaml:"enabled"`
	Threshold  float64           `yaml:"threshold"`
	Strategies []QualityStrategy `yaml:"strategies"`
	Blacklist  QualityBlacklist  `yaml:"blacklist"`
}

// QualityStrategy 质量评估策略
type QualityStrategy struct {
	Name    string                 `yaml:"name"`
	Weight  float64                `yaml:"weight"`
	Enabled bool                   `yaml:"enabled"`
	Config  map[string]interface{} `yaml:"config"`
}

// QualityBlacklist 质量评估黑名单
type QualityBlacklist struct {
	ApologyKeywords   []string `yaml:"apology_keywords"`
	ErrorKeywords     []string `yaml:"error_keywords"`
	MinAnswerLength   int      `yaml:"min_answer_length"`
	MaxAnswerLength   int      `yaml:"max_answer_length"`
	MinQuestionLength int      `yaml:"min_question_length"`
	MaxQuestionLength int      `yaml:"max_question_length"`
}

// Validate 验证配置的有效性
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

// Validate 验证服务器配置
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

// Validate 验证数据库配置
func (d *DatabaseConfig) Validate() error {
	if d.Type == "" {
		return fmt.Errorf("database type is required")
	}

	if d.Type == "qdrant" {
		return d.Qdrant.Validate()
	}

	return fmt.Errorf("unsupported database type: %s", d.Type)
}

// Validate 验证Qdrant配置
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

// Validate 验证嵌入模型配置
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

// Validate 验证本地嵌入模型配置
func (l *LocalEmbedding) Validate() error {
	if l.ModelPath == "" {
		return fmt.Errorf("local embedding model path is required")
	}

	if l.BatchSize <= 0 {
		l.BatchSize = 32 // 默认批处理大小
	}

	return nil
}

// Validate 验证远程嵌入模型配置
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

// Validate 验证日志配置
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

	return nil
}

// Validate 验证缓存配置
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

// Validate 验证质量评估配置
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

// GetAddr 获取服务器监听地址
func (s *ServerConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

// GetQdrantAddr 获取Qdrant地址
func (q *QdrantConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", q.Host, q.Port)
}
