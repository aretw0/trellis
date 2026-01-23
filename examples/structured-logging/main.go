package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

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

	// 5. Run Flow with Cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle SIGINT/SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Info("received interrupt signal, shutting down...")
		cancel()
	}()

	state, err := eng.Start(ctx, "logging-demo", nil)
	if err != nil {
		if err == context.Canceled {
			logger.Info("start canceled")
			return
		}
		logger.Error("failed to start", "error", err)
		os.Exit(1)
	}

	r := runner.NewRunner()

	if _, err := r.Run(context.Background(), eng, state); err != nil {
		if err == context.Canceled || err == context.DeadlineExceeded {
			logger.Info("execution canceled")
			return
		}
		logger.Error("execution failed", "error", err)
		os.Exit(1)
	}

	// Wait for final interrupt if not already received, or exit if flow finished naturally
	// For this demo, we want to keep the metrics server alive until manual cancellation
	logger.Info("Flow finished. Metrics available at http://localhost:2112/metrics. Press Ctrl+C to exit.")
	<-ctx.Done()
	logger.Info("goodbye")
}
