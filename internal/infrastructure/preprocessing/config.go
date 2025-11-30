package preprocessing

import (
	"time"
)

// Config 请求预处理服务配置
type Config struct {
	// Timeout 处理超时时间
	Timeout time.Duration `json:"timeout" yaml:"timeout"`

	// EnableLogging 是否启用详细日志
	EnableLogging bool `json:"enable_logging" yaml:"enable_logging"`
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Timeout <= 0 {
		c.Timeout = 30 * time.Second
	}
	return nil
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Timeout:       30 * time.Second,
		EnableLogging: true,
	}
}
