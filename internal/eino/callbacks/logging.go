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

// LoggingHandler 日志回调处理器
type LoggingHandler struct {
	logger logger.Logger
	cfg    *config.LoggingCallbackConfig
}

// NewLoggingHandler 创建日志回调处理器
func NewLoggingHandler(log logger.Logger, cfg *config.LoggingCallbackConfig) callbacks.Handler {
	return &LoggingHandler{
		logger: log,
		cfg:    cfg,
	}
}

// OnStart 组件开始执行时调用
func (h *LoggingHandler) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	if !h.cfg.Enabled {
		return ctx
	}

	startTime := time.Now()
	ctx = context.WithValue(ctx, startTimeKey, startTime)

	h.logger.InfoContext(ctx, "组件开始执行",
		"component", info.Component,
		"name", info.Name,
		"type", info.Type,
	)

	return ctx
}

// OnEnd 组件执行完成时调用
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

// OnError 组件执行出错时调用
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

// OnStartWithStreamInput 流式输入开始时调用
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

// OnEndWithStreamOutput 流式输出结束时调用
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

// contextKey 上下文键类型
type contextKey string

const (
	startTimeKey contextKey = "callback_start_time"
)
