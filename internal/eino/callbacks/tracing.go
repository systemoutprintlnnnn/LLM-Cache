// Package callbacks 提供 Eino Callback 处理器实现
package callbacks

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"

	"llm-cache/internal/eino/config"
	"llm-cache/pkg/logger"
)

// TracingHandler 实现链路追踪的回调处理器。
// 它生成和传播 TraceID/SpanID，并记录每个组件的执行跨度（Span）。
type TracingHandler struct {
	cfg    *config.TracingCallbackConfig
	logger logger.Logger
	spanID uint64
}

// SpanInfo 定义了追踪跨度的详细信息。
// 包含 TraceID, SpanID, 父 SpanID, 组件信息以及执行时间等。
type SpanInfo struct {
	TraceID   string
	SpanID    string
	ParentID  string
	Component string
	Name      string
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
	Status    string
	Error     error
	Tags      map[string]string
}

// NewTracingHandler 创建一个新的链路追踪处理器。
// 参数 cfg: 追踪配置。
// 参数 log: 日志记录器（用于输出追踪信息）。
// 返回: callbacks.Handler 接口实现。
func NewTracingHandler(cfg *config.TracingCallbackConfig, log logger.Logger) callbacks.Handler {
	return &TracingHandler{
		cfg:    cfg,
		logger: log,
	}
}

// OnStart 在组件开始执行时被调用。
// 创建新的 Span，生成或继承 TraceID，并将其注入上下文。
func (h *TracingHandler) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	if !h.cfg.Enabled {
		return ctx
	}

	// 获取或创建 TraceID
	traceID := getTraceID(ctx)
	if traceID == "" {
		traceID = generateTraceID()
		ctx = context.WithValue(ctx, traceIDKey, traceID)
	}

	// 创建新的 SpanID
	spanID := h.generateSpanID()
	parentID := getSpanID(ctx)

	span := &SpanInfo{
		TraceID:   traceID,
		SpanID:    spanID,
		ParentID:  parentID,
		Component: string(info.Component),
		Name:      info.Name,
		StartTime: time.Now(),
		Tags: map[string]string{
			"component_type": info.Type,
		},
	}

	ctx = context.WithValue(ctx, currentSpanKey, span)
	ctx = context.WithValue(ctx, spanIDKey, spanID)

	h.logger.DebugContext(ctx, "开始跨度",
		"trace_id", traceID,
		"span_id", spanID,
		"parent_id", parentID,
		"component", info.Component,
		"name", info.Name,
	)

	return ctx
}

// OnEnd 在组件执行完成时被调用。
// 结束当前 Span，计算耗时并记录状态。
func (h *TracingHandler) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	if !h.cfg.Enabled {
		return ctx
	}

	span := getCurrentSpan(ctx)
	if span == nil {
		return ctx
	}

	span.EndTime = time.Now()
	span.Duration = span.EndTime.Sub(span.StartTime)
	span.Status = "OK"

	h.logger.DebugContext(ctx, "结束跨度",
		"trace_id", span.TraceID,
		"span_id", span.SpanID,
		"duration_ms", span.Duration.Milliseconds(),
		"status", span.Status,
	)

	return ctx
}

// OnError 在组件执行出错时被调用。
// 记录错误信息到当前 Span 并标记为失败。
func (h *TracingHandler) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	if !h.cfg.Enabled {
		return ctx
	}

	span := getCurrentSpan(ctx)
	if span == nil {
		return ctx
	}

	span.EndTime = time.Now()
	span.Duration = span.EndTime.Sub(span.StartTime)
	span.Status = "ERROR"
	span.Error = err

	h.logger.ErrorContext(ctx, "跨度出错",
		"trace_id", span.TraceID,
		"span_id", span.SpanID,
		"duration_ms", span.Duration.Milliseconds(),
		"error", err.Error(),
	)

	return ctx
}

// OnStartWithStreamInput 在流式输入开始时被调用。
// 调用 OnStart 处理开始逻辑。
func (h *TracingHandler) OnStartWithStreamInput(ctx context.Context, info *callbacks.RunInfo, input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
	return h.OnStart(ctx, info, nil)
}

// OnEndWithStreamOutput 在流式输出结束时被调用。
// 调用 OnEnd 处理结束逻辑。
func (h *TracingHandler) OnEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {
	return h.OnEnd(ctx, info, nil)
}

// generateSpanID 生成唯一的 Span ID。
// 使用原子计数器和 UUID 组合生成。
func (h *TracingHandler) generateSpanID() string {
	id := atomic.AddUint64(&h.spanID, 1)
	return uuid.NewSHA1(uuid.NameSpaceDNS, []byte(time.Now().String())).String()[:16] + "-" + string(rune(id))
}

// generateTraceID 生成全局唯一的 Trace ID。
// 使用 UUID v4 生成。
func generateTraceID() string {
	return uuid.New().String()
}

// getTraceID 从上下文中获取 Trace ID。
func getTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// getSpanID 从上下文中获取 Span ID。
func getSpanID(ctx context.Context) string {
	if spanID, ok := ctx.Value(spanIDKey).(string); ok {
		return spanID
	}
	return ""
}

// getCurrentSpan 从上下文中获取当前活动的 SpanInfo 对象。
func getCurrentSpan(ctx context.Context) *SpanInfo {
	if span, ok := ctx.Value(currentSpanKey).(*SpanInfo); ok {
		return span
	}
	return nil
}

const (
	// traceIDKey 用于在上下文中存储 Trace ID。
	traceIDKey contextKey = "trace_id"
	// spanIDKey 用于在上下文中存储 Span ID。
	spanIDKey contextKey = "span_id"
	// currentSpanKey 用于在上下文中存储当前 SpanInfo 对象。
	currentSpanKey contextKey = "current_span"
)

// WithTraceID 将指定的 TraceID 注入到上下文中。
// 返回: 包含 TraceID 的新上下文。
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// ExtractTraceID 从上下文中提取当前 TraceID。
// 如果不存在则返回空字符串。
func ExtractTraceID(ctx context.Context) string {
	return getTraceID(ctx)
}
