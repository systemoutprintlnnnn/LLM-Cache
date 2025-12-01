// Package callbacks 提供 Eino Callback 处理器实现
package callbacks

import (
	"context"
	"sync"
	"time"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/schema"

	"llm-cache/internal/eino/config"
)

// MetricsHandler 指标回调处理器
type MetricsHandler struct {
	cfg     *config.MetricsCallbackConfig
	metrics *MetricsCollector
}

// MetricsCollector 指标收集器
type MetricsCollector struct {
	mu sync.RWMutex

	// 调用计数
	TotalCalls      int64
	SuccessfulCalls int64
	FailedCalls     int64

	// 延迟统计
	TotalLatencyMs   int64
	ComponentLatency map[string]*LatencyStats

	// 组件调用计数
	ComponentCalls map[string]int64
}

// LatencyStats 延迟统计
type LatencyStats struct {
	Count   int64
	TotalMs int64
	MinMs   int64
	MaxMs   int64
}

// NewMetricsHandler 创建指标回调处理器
func NewMetricsHandler(cfg *config.MetricsCallbackConfig) callbacks.Handler {
	return &MetricsHandler{
		cfg: cfg,
		metrics: &MetricsCollector{
			ComponentLatency: make(map[string]*LatencyStats),
			ComponentCalls:   make(map[string]int64),
		},
	}
}

// OnStart 组件开始执行时调用
func (h *MetricsHandler) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	if !h.cfg.Enabled {
		return ctx
	}

	h.metrics.mu.Lock()
	defer h.metrics.mu.Unlock()

	h.metrics.TotalCalls++
	h.metrics.ComponentCalls[info.Name]++

	startTime := time.Now()
	ctx = context.WithValue(ctx, metricsStartTimeKey, startTime)

	return ctx
}

// OnEnd 组件执行完成时调用
func (h *MetricsHandler) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	if !h.cfg.Enabled {
		return ctx
	}

	startTime, ok := ctx.Value(metricsStartTimeKey).(time.Time)
	if !ok {
		return ctx
	}

	duration := time.Since(startTime)
	durationMs := duration.Milliseconds()

	h.metrics.mu.Lock()
	defer h.metrics.mu.Unlock()

	h.metrics.SuccessfulCalls++
	h.metrics.TotalLatencyMs += durationMs

	// 更新组件延迟统计
	stats, exists := h.metrics.ComponentLatency[info.Name]
	if !exists {
		stats = &LatencyStats{
			MinMs: durationMs,
			MaxMs: durationMs,
		}
		h.metrics.ComponentLatency[info.Name] = stats
	}

	stats.Count++
	stats.TotalMs += durationMs
	if durationMs < stats.MinMs {
		stats.MinMs = durationMs
	}
	if durationMs > stats.MaxMs {
		stats.MaxMs = durationMs
	}

	return ctx
}

// OnError 组件执行出错时调用
func (h *MetricsHandler) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	if !h.cfg.Enabled {
		return ctx
	}

	h.metrics.mu.Lock()
	defer h.metrics.mu.Unlock()

	h.metrics.FailedCalls++

	return ctx
}

// OnStartWithStreamInput 流式输入开始时调用
func (h *MetricsHandler) OnStartWithStreamInput(ctx context.Context, info *callbacks.RunInfo, input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
	if !h.cfg.Enabled {
		return ctx
	}

	h.metrics.mu.Lock()
	h.metrics.TotalCalls++
	h.metrics.ComponentCalls[info.Name]++
	h.metrics.mu.Unlock()

	startTime := time.Now()
	ctx = context.WithValue(ctx, metricsStartTimeKey, startTime)

	return ctx
}

// OnEndWithStreamOutput 流式输出结束时调用
func (h *MetricsHandler) OnEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {
	return h.OnEnd(ctx, info, nil)
}

// GetMetrics 获取当前指标
func (h *MetricsHandler) GetMetrics() map[string]interface{} {
	h.metrics.mu.RLock()
	defer h.metrics.mu.RUnlock()

	avgLatency := int64(0)
	if h.metrics.SuccessfulCalls > 0 {
		avgLatency = h.metrics.TotalLatencyMs / h.metrics.SuccessfulCalls
	}

	componentStats := make(map[string]interface{})
	for name, stats := range h.metrics.ComponentLatency {
		avgMs := int64(0)
		if stats.Count > 0 {
			avgMs = stats.TotalMs / stats.Count
		}
		componentStats[name] = map[string]interface{}{
			"count":  stats.Count,
			"avg_ms": avgMs,
			"min_ms": stats.MinMs,
			"max_ms": stats.MaxMs,
		}
	}

	return map[string]interface{}{
		"total_calls":      h.metrics.TotalCalls,
		"successful_calls": h.metrics.SuccessfulCalls,
		"failed_calls":     h.metrics.FailedCalls,
		"avg_latency_ms":   avgLatency,
		"component_stats":  componentStats,
		"component_calls":  h.metrics.ComponentCalls,
	}
}

// Reset 重置指标
func (h *MetricsHandler) Reset() {
	h.metrics.mu.Lock()
	defer h.metrics.mu.Unlock()

	h.metrics.TotalCalls = 0
	h.metrics.SuccessfulCalls = 0
	h.metrics.FailedCalls = 0
	h.metrics.TotalLatencyMs = 0
	h.metrics.ComponentLatency = make(map[string]*LatencyStats)
	h.metrics.ComponentCalls = make(map[string]int64)
}

const (
	metricsStartTimeKey contextKey = "metrics_start_time"
)
