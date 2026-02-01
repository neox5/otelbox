package monitor

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

// Monitor tracks system resource usage and saturation indicators.
type Monitor struct {
	interval time.Duration
	logger   *slog.Logger
	wg       sync.WaitGroup
	proc     *process.Process
}

// New creates a new monitor with specified collection interval.
func New(interval time.Duration, logger *slog.Logger) *Monitor {
	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		logger.Error("failed to get process handle", "error", err)
		return nil
	}

	return &Monitor{
		interval: interval,
		logger:   logger,
		proc:     proc,
	}
}

// Run starts the monitoring loop in a background goroutine.
// Blocks until context is cancelled.
func (m *Monitor) Run(ctx context.Context) {
	m.wg.Go(func() {
		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()

		// Immediate first collection
		m.collect()

		for {
			select {
			case <-ctx.Done():
				m.logger.Info("monitor shutdown complete")
				return
			case <-ticker.C:
				m.collect()
			}
		}
	})
}

// Wait blocks until the monitor goroutine exits.
func (m *Monitor) Wait() {
	m.wg.Wait()
}

// collect reads current metrics and logs resource usage.
func (m *Monitor) collect() {
	// ---- CPU ----
	processCPU, err := m.proc.CPUPercent()
	if err != nil {
		m.logger.Warn("failed to get CPU percent", "error", err)
		processCPU = 0
	}

	cores := runtime.GOMAXPROCS(-1)
	maxCPU := float64(cores * 100)

	utilization := 0.0
	if maxCPU > 0 {
		utilization = processCPU / maxCPU
	}

	// ---- Runtime / Memory ----
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	goroutines := runtime.NumGoroutine()

	// ---- Saturation ----
	saturation := "normal"
	if utilization > 0.95 {
		saturation = "saturated"
	} else if utilization > 0.80 {
		saturation = "high"
	}

	// ---- Helpers ----
	mb := func(b uint64) float64 {
		return float64(b) / (1024 * 1024)
	}
	kb := func(b uint64) float64 {
		return float64(b) / 1024
	}

	// ---- Compact INFO log (Option B) ----
	m.logger.LogAttrs(
		context.Background(),
		slog.LevelInfo,
		"resource",
		slog.String("cpu", fmt.Sprintf("%.4f%%", processCPU)),
		slog.String("util", fmt.Sprintf("%.4f%%", utilization*100)),
		slog.Int("cores", cores),
		slog.Int("gor", goroutines),
		slog.String(
			"mem",
			fmt.Sprintf(
				"alloc:%.2fMB sys:%.2fMB stack:%.0fKB",
				mb(ms.HeapAlloc),
				mb(ms.HeapSys),
				kb(ms.StackInuse),
			),
		),
		slog.Uint64("gc", uint64(ms.NumGC)),
		slog.String("gc_cpu", fmt.Sprintf("%.3f", ms.GCCPUFraction)),
		slog.String("sat", saturation),
	)

	// ---- Saturation warning (unchanged semantics, compact keys) ----
	if saturation == "saturated" {
		m.logger.Warn(
			"cpu saturation detected",
			"cpu", processCPU,
			"util_pct", utilization*100,
			"action", "reduce load or increase GOMAXPROCS",
		)
	}
}
