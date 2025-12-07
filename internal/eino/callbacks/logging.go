// Package callbacks 提供 Eino Callback 处理器实现
package callbacks

import (
	"context"
	"time"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/schema"

	"llm-cache/internal/eino/config"
	"llm-cache/pkg/logger"
)

// LoggingHandler 实现基于日志的 Callback 处理器。
// 它会在组件开始、结束或出错时记录详细的日志信息。
type LoggingHandler struct {
	logger logger.Logger
	cfg    *config.LoggingCallbackConfig
}

// NewLoggingHandler 创建一个新的日志回调处理器。
// 参数 log: 底层日志记录器。
// 参数 cfg: 日志回调配置。
// 返回: callbacks.Handler 接口实现。
func NewLoggingHandler(log logger.Logger, cfg *config.LoggingCallbackConfig) callbacks.Handler {
	return &LoggingHandler{
		logger: log,
		cfg:    cfg,
	}
}

// OnStart 在组件开始执行时被调用。
// 记录组件名称、类型和开始时间，并将开始时间注入上下文以计算耗时。
func (h *LoggingHandler) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	if !h.cfg.Enabled {
		return ctx
	}

	startTime := time.Now()
	ctx = context.WithValue(ctx, startTimeKey, startTime)

	// 注入组件字段，后续日志自动携带
	ctx = logger.InjectFields(ctx, logger.Fields{
		"component": info.Component,
		"name":      info.Name,
		"type":      info.Type,
	})

	h.logger.InfoContext(ctx, "组件开始执行",
		"component", info.Component,
		"name", info.Name,
		"type", info.Type,
	)

	return ctx
}

// OnEnd 在组件执行完成时被调用。
// 计算并记录组件的执行耗时。
func (h *LoggingHandler) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	if !h.cfg.Enabled {
		return ctx
	}

	// 计算执行时间
	startTime, _ := ctx.Value(startTimeKey).(time.Time)
	duration := time.Since(startTime)

	h.logger.InfoContext(ctx, "组件执行完成",
		"component", info.Component,
		"name", info.Name,
		"type", info.Type,
		"duration_ms", duration.Milliseconds(),
	)

	return ctx
}

// OnError 在组件执行出错时被调用。
// 记录错误详情和执行耗时。
func (h *LoggingHandler) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	if !h.cfg.Enabled {
		return ctx
	}

	// 计算执行时间
	startTime, _ := ctx.Value(startTimeKey).(time.Time)
	duration := time.Since(startTime)

	h.logger.ErrorContext(ctx, "组件执行出错",
		"component", info.Component,
		"name", info.Name,
		"type", info.Type,
		"duration_ms", duration.Milliseconds(),
		"error", err.Error(),
	)

	return ctx
}

// OnStartWithStreamInput 在流式输入开始时被调用。
// 记录流式处理的开始时间和组件信息。
func (h *LoggingHandler) OnStartWithStreamInput(ctx context.Context, info *callbacks.RunInfo, input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
	if !h.cfg.Enabled {
		return ctx
	}

	startTime := time.Now()
	ctx = context.WithValue(ctx, startTimeKey, startTime)

	h.logger.InfoContext(ctx, "组件流式输入开始",
		"component", info.Component,
		"name", info.Name,
	)

	return ctx
}

// OnEndWithStreamOutput 在流式输出结束时被调用。
// 记录流式处理的结束时间和总耗时。
func (h *LoggingHandler) OnEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {
	if !h.cfg.Enabled {
		return ctx
	}

	startTime, _ := ctx.Value(startTimeKey).(time.Time)
	duration := time.Since(startTime)

	h.logger.InfoContext(ctx, "组件流式输出完成",
		"component", info.Component,
		"name", info.Name,
		"duration_ms", duration.Milliseconds(),
	)

	return ctx
}

// contextKey 定义了上下文键的类型，用于防止键名冲突。
type contextKey string

const (
	// startTimeKey 用于在上下文中存储组件开始执行的时间。
	startTimeKey contextKey = "callback_start_time"
)
