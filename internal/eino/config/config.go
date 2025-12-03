// Package config 定义 Eino 框架的配置结构
package config

import "time"

// EinoConfig Eino 框架的总配置结构。
// 包含了 Embedder、Retriever、Indexer 等组件的配置，以及查询、存储、质量检查和回调系统的配置。
type EinoConfig struct {
	Embedder  EmbedderConfig  `yaml:"embedder"`
	Retriever RetrieverConfig `yaml:"retriever"`
	Indexer   IndexerConfig   `yaml:"indexer"`
	Query     QueryConfig     `yaml:"query"`
	Store     StoreConfig     `yaml:"store"`
	Quality   QualityConfig   `yaml:"quality"`
	Callbacks CallbacksConfig `yaml:"callbacks"`
}

// EmbedderConfig 定义文本嵌入（Embedding）服务的配置。
// 支持多种提供商（OpenAI, ARK, Ollama 等）及特定的认证和模型参数。
type EmbedderConfig struct {
	Provider   string `yaml:"provider"` // openai, ark, ollama, dashscope, qianfan, tencentcloud
	APIKey     string `yaml:"api_key"`
	BaseURL    string `yaml:"base_url"`
	Model      string `yaml:"model"`
	Timeout    int    `yaml:"timeout"`    // 秒
	Dimensions *int   `yaml:"dimensions"` // 向量维度（可选）

	// OpenAI/Azure 专用
	ByAzure    bool   `yaml:"by_azure"`
	APIVersion string `yaml:"api_version"`

	// ARK 专用
	Region     string `yaml:"region"`
	AccessKey  string `yaml:"access_key"`
	SecretKey  string `yaml:"secret_key"`
	RetryTimes *int   `yaml:"retry_times"`

	// Tencentcloud 专用
	SecretID string `yaml:"secret_id"`
}

// RetrieverConfig 定义检索器（Retriever）的配置。
// 负责从向量数据库中检索相似的文本片段。
type RetrieverConfig struct {
	Provider       string  `yaml:"provider"` // qdrant, milvus, redis, es8, vikingdb
	Collection     string  `yaml:"collection"`
	TopK           int     `yaml:"top_k"`
	ScoreThreshold float64 `yaml:"score_threshold"`

	// Qdrant 专用配置
	Qdrant QdrantRetrieverConfig `yaml:"qdrant"`

	// Milvus 专用配置
	Milvus MilvusRetrieverConfig `yaml:"milvus"`

	// Redis 专用配置
	Redis RedisRetrieverConfig `yaml:"redis"`

	// Elasticsearch 专用配置
	ES8 ES8RetrieverConfig `yaml:"es8"`
}

// QdrantRetrieverConfig 定义 Qdrant 检索器的专用配置。
type QdrantRetrieverConfig struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	APIKey     string `yaml:"api_key"`
	UseTLS     bool   `yaml:"use_tls"`
	VectorName string `yaml:"vector_name"`
}

// MilvusRetrieverConfig 定义 Milvus 检索器的专用配置。
type MilvusRetrieverConfig struct {
	Host         string   `yaml:"host"`
	Port         int      `yaml:"port"`
	Username     string   `yaml:"username"`
	Password     string   `yaml:"password"`
	VectorField  string   `yaml:"vector_field"`
	OutputFields []string `yaml:"output_fields"`
	MetricType   string   `yaml:"metric_type"`
}

// RedisRetrieverConfig 定义 Redis 检索器的专用配置。
type RedisRetrieverConfig struct {
	Addr         string   `yaml:"addr"`
	Password     string   `yaml:"password"`
	DB           int      `yaml:"db"`
	Index        string   `yaml:"index"`
	VectorField  string   `yaml:"vector_field"`
	ReturnFields []string `yaml:"return_fields"`
}

// ES8RetrieverConfig 定义 Elasticsearch 8 检索器的专用配置。
type ES8RetrieverConfig struct {
	Addresses   []string `yaml:"addresses"`
	Username    string   `yaml:"username"`
	Password    string   `yaml:"password"`
	Index       string   `yaml:"index"`
	VectorField string   `yaml:"vector_field"`
	SearchMode  string   `yaml:"search_mode"` // knn, hybrid
}

// IndexerConfig 定义索引器（Indexer）的配置。
// 负责将文本向量化并存储到向量数据库中。
type IndexerConfig struct {
	Provider   string `yaml:"provider"`
	Collection string `yaml:"collection"`
	VectorSize int    `yaml:"vector_size"`

	// Qdrant 专用配置
	Qdrant QdrantIndexerConfig `yaml:"qdrant"`

	// Milvus 专用配置
	Milvus MilvusIndexerConfig `yaml:"milvus"`

	// Redis 专用配置
	Redis RedisIndexerConfig `yaml:"redis"`

	// Elasticsearch 专用配置
	ES8 ES8IndexerConfig `yaml:"es8"`
}

// QdrantIndexerConfig 定义 Qdrant 索引器的专用配置。
type QdrantIndexerConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	APIKey   string `yaml:"api_key"`
	UseTLS   bool   `yaml:"use_tls"`
	Distance string `yaml:"distance"` // Cosine, Euclid, Dot
}

// MilvusIndexerConfig 定义 Milvus 索引器的专用配置。
type MilvusIndexerConfig struct {
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	VectorField string `yaml:"vector_field"`
}

// RedisIndexerConfig 定义 Redis 索引器的专用配置。
type RedisIndexerConfig struct {
	Addr        string `yaml:"addr"`
	Password    string `yaml:"password"`
	DB          int    `yaml:"db"`
	Index       string `yaml:"index"`
	Prefix      string `yaml:"prefix"`
	VectorField string `yaml:"vector_field"`
}

// ES8IndexerConfig 定义 Elasticsearch 8 索引器的专用配置。
type ES8IndexerConfig struct {
	Addresses   []string `yaml:"addresses"`
	Username    string   `yaml:"username"`
	Password    string   `yaml:"password"`
	Index       string   `yaml:"index"`
	VectorField string   `yaml:"vector_field"`
}

// QueryConfig 定义查询流程（Query Graph）的配置。
// 包含预处理、后处理开关，结果选择策略以及超时设置。
type QueryConfig struct {
	// 节点开关
	PreprocessEnabled  bool `yaml:"preprocess_enabled"`
	PostprocessEnabled bool `yaml:"postprocess_enabled"`

	// 选择策略
	SelectionStrategy string  `yaml:"selection_strategy"` // first, highest_score, temperature_softmax
	Temperature       float64 `yaml:"temperature"`

	// 超时配置（秒）
	EmbeddingTimeout int `yaml:"embedding_timeout"`
	RetrieveTimeout  int `yaml:"retrieve_timeout"`
}

// StoreConfig 定义存储流程（Store Graph）的配置。
// 包含质量检查开关、文本长度限制和相似度阈值。
type StoreConfig struct {
	// 质量检查配置
	QualityCheckEnabled bool    `yaml:"quality_check_enabled"`
	MinQuestionLength   int     `yaml:"min_question_length"`
	MinAnswerLength     int     `yaml:"min_answer_length"`
	ScoreThreshold      float64 `yaml:"score_threshold"`
}

// QualityConfig 定义质量检查组件的详细配置。
// 包含文本长度、语义相关性、综合分数阈值以及并行处理参数。
type QualityConfig struct {
	// 是否启用质量检查
	Enabled bool `yaml:"enabled"`

	// 长度检查
	MinQuestionLength int `yaml:"min_question_length"`
	MinAnswerLength   int `yaml:"min_answer_length"`
	MaxQuestionLength int `yaml:"max_question_length"`
	MaxAnswerLength   int `yaml:"max_answer_length"`

	// 语义检查
	SemanticRelevanceThreshold float64 `yaml:"semantic_relevance_threshold"`

	// 综合分数阈值
	ScoreThreshold float64 `yaml:"score_threshold"`

	// 并行执行配置
	ParallelWorkers int           `yaml:"parallel_workers"`
	CheckTimeout    time.Duration `yaml:"check_timeout"`

	// 黑名单关键词
	BlacklistKeywords []string `yaml:"blacklist_keywords"`
}

// CallbacksConfig 定义 Eino 框架的回调系统配置。
// 支持日志、监控指标、链路追踪以及 Langfuse 等第三方平台集成。
type CallbacksConfig struct {
	Logging  LoggingCallbackConfig  `yaml:"logging"`
	Metrics  MetricsCallbackConfig  `yaml:"metrics"`
	Tracing  TracingCallbackConfig  `yaml:"tracing"`
	Langfuse LangfuseCallbackConfig `yaml:"langfuse"`
	APMPlus  APMPlusCallbackConfig  `yaml:"apmplus"`
	Cozeloop CozeloopCallbackConfig `yaml:"cozeloop"`
}

// LoggingCallbackConfig 定义日志回调的配置。
type LoggingCallbackConfig struct {
	Enabled bool   `yaml:"enabled"`
	Level   string `yaml:"level"`
}

// MetricsCallbackConfig 定义指标监控回调的配置。
type MetricsCallbackConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint"`
}

// TracingCallbackConfig 定义链路追踪回调的配置。
type TracingCallbackConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint"`
}

// LangfuseCallbackConfig 定义 Langfuse 平台的集成配置。
type LangfuseCallbackConfig struct {
	Enabled       bool   `yaml:"enabled"`
	PublicKey     string `yaml:"public_key"`
	SecretKey     string `yaml:"secret_key"`
	Host          string `yaml:"host"`
	FlushInterval int    `yaml:"flush_interval"` // 秒
	BatchSize     int    `yaml:"batch_size"`
}

// APMPlusCallbackConfig 定义 APMPlus 平台的集成配置。
type APMPlusCallbackConfig struct {
	Enabled     bool   `yaml:"enabled"`
	AppKey      string `yaml:"app_key"`
	Region      string `yaml:"region"`
	ServiceName string `yaml:"service_name"`
	Environment string `yaml:"environment"`
}

// CozeloopCallbackConfig 定义 Cozeloop 平台的集成配置。
type CozeloopCallbackConfig struct {
	Enabled  bool   `yaml:"enabled"`
	APIKey   string `yaml:"api_key"`
	Endpoint string `yaml:"endpoint"`
}

// DefaultEinoConfig 创建并返回一个包含默认值的 EinoConfig 对象。
// 默认配置使用 OpenAI (Embedder) 和 Qdrant (Retriever/Indexer)，并开启了常用的功能开关。
func DefaultEinoConfig() *EinoConfig {
	return &EinoConfig{
		Embedder: EmbedderConfig{
			Provider: "openai",
			Model:    "text-embedding-3-small",
			Timeout:  30,
		},
		Retriever: RetrieverConfig{
			Provider:       "qdrant",
			Collection:     "llm_cache",
			TopK:           5,
			ScoreThreshold: 0.7,
			Qdrant: QdrantRetrieverConfig{
				Host: "localhost",
				Port: 6334,
			},
		},
		Indexer: IndexerConfig{
			Provider:   "qdrant",
			Collection: "llm_cache",
			VectorSize: 1536,
			Qdrant: QdrantIndexerConfig{
				Host:     "localhost",
				Port:     6334,
				Distance: "Cosine",
			},
		},
		Query: QueryConfig{
			PreprocessEnabled:  true,
			PostprocessEnabled: true,
			SelectionStrategy:  "highest_score",
			Temperature:        0.7,
			EmbeddingTimeout:   30,
			RetrieveTimeout:    30,
		},
		Store: StoreConfig{
			QualityCheckEnabled: true,
			MinQuestionLength:   5,
			MinAnswerLength:     10,
			ScoreThreshold:      0.5,
		},
		Quality: QualityConfig{
			Enabled:                    true,
			MinQuestionLength:          5,
			MinAnswerLength:            10,
			MaxQuestionLength:          10000,
			MaxAnswerLength:            100000,
			SemanticRelevanceThreshold: 0.3,
			ScoreThreshold:             0.5,
			ParallelWorkers:            3,
			CheckTimeout:               5 * time.Second,
			BlacklistKeywords:          []string{},
		},
		Callbacks: CallbacksConfig{
			Logging: LoggingCallbackConfig{
				Enabled: true,
				Level:   "info",
			},
			Metrics: MetricsCallbackConfig{
				Enabled:  false,
				Endpoint: "/metrics",
			},
			Tracing: TracingCallbackConfig{
				Enabled: false,
			},
			Langfuse: LangfuseCallbackConfig{
				Enabled: false,
			},
			APMPlus: APMPlusCallbackConfig{
				Enabled: false,
			},
			Cozeloop: CozeloopCallbackConfig{
				Enabled: false,
			},
		},
	}
}
