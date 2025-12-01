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

// TracingHandler 链路追踪回调处理器
type TracingHandler struct {
	cfg    *config.TracingCallbackConfig
	logger logger.Logger
	spanID uint64
}

// SpanInfo 跨度信息
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

// NewTracingHandler 创建链路追踪回调处理器
func NewTracingHandler(cfg *config.TracingCallbackConfig, log logger.Logger) callbacks.Handler {
	return &TracingHandler{
		cfg:    cfg,
		logger: log,
	}
}

// OnStart 组件开始执行时调用
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

// OnEnd 组件执行完成时调用
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

// OnError 组件执行出错时调用
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

// OnStartWithStreamInput 流式输入开始时调用
func (h *TracingHandler) OnStartWithStreamInput(ctx context.Context, info *callbacks.RunInfo, input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
	return h.OnStart(ctx, info, nil)
}

// OnEndWithStreamOutput 流式输出结束时调用
func (h *TracingHandler) OnEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {
	return h.OnEnd(ctx, info, nil)
}

// generateSpanID 生成 Span ID
func (h *TracingHandler) generateSpanID() string {
	id := atomic.AddUint64(&h.spanID, 1)
	return uuid.NewSHA1(uuid.NameSpaceDNS, []byte(time.Now().String())).String()[:16] + "-" + string(rune(id))
}

// generateTraceID 生成 Trace ID
func generateTraceID() string {
	return uuid.New().String()
}

// getTraceID 从上下文获取 Trace ID
func getTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// getSpanID 从上下文获取 Span ID
func getSpanID(ctx context.Context) string {
	if spanID, ok := ctx.Value(spanIDKey).(string); ok {
		return spanID
	}
	return ""
}

// getCurrentSpan 从上下文获取当前跨度
func getCurrentSpan(ctx context.Context) *SpanInfo {
	if span, ok := ctx.Value(currentSpanKey).(*SpanInfo); ok {
		return span
	}
	return nil
}

const (
	traceIDKey     contextKey = "trace_id"
	spanIDKey      contextKey = "span_id"
	currentSpanKey contextKey = "current_span"
)

// WithTraceID 设置 Trace ID 到上下文
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// ExtractTraceID 从上下文提取 Trace ID
func ExtractTraceID(ctx context.Context) string {
	return getTraceID(ctx)
}
