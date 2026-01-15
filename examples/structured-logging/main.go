package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/aretw0/trellis"
	"github.com/aretw0/trellis/pkg/domain"
	"github.com/aretw0/trellis/pkg/runner"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// 1. Setup Structured Logger (Slog)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// 2. Setup Prometheus Metrics
	nodeVisits := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "trellis_node_visits_total",
			Help: "Total number of node visits",
		},
		[]string{"node_id"},
	)
	toolDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "trellis_tool_duration_seconds",
			Help: "Duration of tool executions",
		},
		[]string{"tool_name"},
	)
	prometheus.MustRegister(nodeVisits, toolDuration)

	// Start Metrics Server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		logger.Info("Starting metrics server on :2112")
		http.ListenAndServe(":2112", nil)
	}()

	// 3. Create Lifecycle Hooks
	hooks := domain.LifecycleHooks{
		OnNodeEnter: func(ctx context.Context, e *domain.NodeEvent) {
			// Log Event
			logger.Info("node_enter",
				"node_id", e.NodeID,
				"type", e.NodeType,
			)
			// Record Metric
			nodeVisits.WithLabelValues(e.NodeID).Inc()
		},
		OnNodeLeave: func(ctx context.Context, e *domain.NodeEvent) {
			logger.Info("node_leave", "node_id", e.NodeID)
		},
		OnToolCall: func(ctx context.Context, e *domain.ToolEvent) {
			logger.Info("tool_call", "tool_name", e.ToolName)
		},
		OnToolReturn: func(ctx context.Context, e *domain.ToolEvent) {
			logger.Info("tool_return",
				"tool_name", e.ToolName,
				"is_error", e.IsError,
			)
			// Record Metric (Mock duration for demo)
			toolDuration.WithLabelValues(e.ToolName).Observe(0.1)
		},
	}

	// 4. Initialize Engine
	eng, err := trellis.New("examples/structured-logging", trellis.WithLifecycleHooks(hooks))
	if err != nil {
		logger.Error("failed to create engine", "error", err)
		os.Exit(1)
	}

	// 5. Run Flow
	ctx := context.Background()
	state, err := eng.Start(ctx)
	if err != nil {
		logger.Error("failed to start", "error", err)
		os.Exit(1)
	}

	r := &runner.Runner{
		Input:  os.Stdin,
		Output: os.Stdout,
	}

	if err := r.Run(eng, state); err != nil {
		logger.Error("execution failed", "error", err)
		os.Exit(1)
	}

	// Keep alive for scraping
	logger.Info("Flow finished. Metrics available at http://localhost:2112/metrics. Press Ctrl+C to exit.")
	select {}
}
