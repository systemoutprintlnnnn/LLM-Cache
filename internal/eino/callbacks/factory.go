// Package callbacks 提供 Eino Callback 处理器实现
package callbacks

import (
	"github.com/cloudwego/eino/callbacks"

	"llm-cache/internal/eino/config"
	"llm-cache/pkg/logger"
)

// Factory 负责创建和管理 Eino 框架的各类 Callback 处理器。
// 它根据配置动态实例化日志、监控、追踪等处理器。
type Factory struct {
	cfg    *config.CallbacksConfig
	logger logger.Logger
}

// NewFactory 创建一个新的 Callback 工厂实例。
// 参数 cfg: 回调系统配置。
// 参数 log: 日志记录器。
// 返回: 初始化后的 Factory 指针。
func NewFactory(cfg *config.CallbacksConfig, log logger.Logger) *Factory {
	return &Factory{
		cfg:    cfg,
		logger: log,
	}
}

// CreateHandlers 根据配置创建并返回所有启用的 Callback 处理器列表。
// 这些处理器将被注入到 Eino Graph 中，用于监控组件的运行状态。
func (f *Factory) CreateHandlers() []callbacks.Handler {
	handlers := make([]callbacks.Handler, 0)

	// 日志回调
	if f.cfg.Logging.Enabled {
		handlers = append(handlers, NewLoggingHandler(f.logger, &f.cfg.Logging))
	}

	// 指标回调
	if f.cfg.Metrics.Enabled {
		handlers = append(handlers, NewMetricsHandler(&f.cfg.Metrics))
	}

	// 链路追踪回调
	if f.cfg.Tracing.Enabled {
		handlers = append(handlers, NewTracingHandler(&f.cfg.Tracing, f.logger))
	}

	// Langfuse 回调（需要额外导入 eino-ext/callbacks/langfuse）
	if f.cfg.Langfuse.Enabled {
		handler := f.createLangfuseHandler()
		if handler != nil {
			handlers = append(handlers, handler)
		}
	}

	// APMPlus 回调（需要额外导入 eino-ext/callbacks/apmplus）
	if f.cfg.APMPlus.Enabled {
		handler := f.createAPMPlusHandler()
		if handler != nil {
			handlers = append(handlers, handler)
		}
	}

	// Cozeloop 回调（需要额外导入 eino-ext/callbacks/cozeloop）
	if f.cfg.Cozeloop.Enabled {
		handler := f.createCozeloopHandler()
		if handler != nil {
			handlers = append(handlers, handler)
		}
	}

	return handlers
}

// createLangfuseHandler 创建 Langfuse 回调处理器
func (f *Factory) createLangfuseHandler() callbacks.Handler {
	/*
		import langfuse "github.com/cloudwego/eino-ext/callbacks/langfuse"

		handler, err := langfuse.NewLangfuseHandler(&langfuse.Config{
			PublicKey:     f.cfg.Langfuse.PublicKey,
			SecretKey:     f.cfg.Langfuse.SecretKey,
			Host:          f.cfg.Langfuse.Host,
			FlushInterval: time.Duration(f.cfg.Langfuse.FlushInterval) * time.Second,
			BatchSize:     f.cfg.Langfuse.BatchSize,
		})
		if err != nil {
			f.logger.Error("创建 Langfuse Handler 失败", "error", err)
			return nil
		}
		return handler
	*/
	f.logger.Warn("Langfuse 回调未启用：需要添加 github.com/cloudwego/eino-ext/callbacks/langfuse 依赖")
	return nil
}

// createAPMPlusHandler 创建 APMPlus 回调处理器
func (f *Factory) createAPMPlusHandler() callbacks.Handler {
	/*
		import apmplus "github.com/cloudwego/eino-ext/callbacks/apmplus"

		handler, err := apmplus.NewAPMPlusHandler(&apmplus.Config{
			AppKey:      f.cfg.APMPlus.AppKey,
			Region:      f.cfg.APMPlus.Region,
			ServiceName: f.cfg.APMPlus.ServiceName,
			Environment: f.cfg.APMPlus.Environment,
		})
		if err != nil {
			f.logger.Error("创建 APMPlus Handler 失败", "error", err)
			return nil
		}
		return handler
	*/
	f.logger.Warn("APMPlus 回调未启用：需要添加 github.com/cloudwego/eino-ext/callbacks/apmplus 依赖")
	return nil
}

// createCozeloopHandler 创建 Cozeloop 回调处理器
func (f *Factory) createCozeloopHandler() callbacks.Handler {
	/*
		import cozeloop "github.com/cloudwego/eino-ext/callbacks/cozeloop"

		handler, err := cozeloop.NewCozeloopHandler(&cozeloop.Config{
			APIKey:   f.cfg.Cozeloop.APIKey,
			Endpoint: f.cfg.Cozeloop.Endpoint,
		})
		if err != nil {
			f.logger.Error("创建 Cozeloop Handler 失败", "error", err)
			return nil
		}
		return handler
	*/
	f.logger.Warn("Cozeloop 回调未启用：需要添加 github.com/cloudwego/eino-ext/callbacks/cozeloop 依赖")
	return nil
}

// GetLoggingHandler 获取日志回调处理器实例。
// 如果未启用日志回调，则返回 nil。
func (f *Factory) GetLoggingHandler() callbacks.Handler {
	if !f.cfg.Logging.Enabled {
		return nil
	}
	return NewLoggingHandler(f.logger, &f.cfg.Logging)
}

// GetMetricsHandler 获取指标回调处理器实例。
// 如果未启用指标回调，则返回 nil。
func (f *Factory) GetMetricsHandler() *MetricsHandler {
	if !f.cfg.Metrics.Enabled {
		return nil
	}
	return NewMetricsHandler(&f.cfg.Metrics).(*MetricsHandler)
}

// GetTracingHandler 获取链路追踪回调处理器实例。
// 如果未启用追踪回调，则返回 nil。
func (f *Factory) GetTracingHandler() callbacks.Handler {
	if !f.cfg.Tracing.Enabled {
		return nil
	}
	return NewTracingHandler(&f.cfg.Tracing, f.logger)
}
