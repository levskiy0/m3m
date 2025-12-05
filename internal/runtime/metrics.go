package runtime

import (
	"runtime"
	"sync"
	"time"
)

const (
	// MetricsInterval is the interval between metrics snapshots
	MetricsInterval = 1 * time.Minute
	// MetricsHistorySize is the number of snapshots to keep (24h / 1min = 1440)
	MetricsHistorySize = 1440
)

// MetricsSnapshot represents a point-in-time metrics snapshot
type MetricsSnapshot struct {
	Timestamp   time.Time `json:"timestamp"`
	MemoryAlloc uint64    `json:"memory_alloc"`
	MemorySys   uint64    `json:"memory_sys"`
	NumGC       uint32    `json:"num_gc"`
	Requests    int64     `json:"requests"`
	Jobs        int64     `json:"jobs"`
	// CPU usage as percentage (0-100)
	CPUPercent float64 `json:"cpu_percent"`
}

// MetricsHistory stores historical metrics data
type MetricsHistory struct {
	mu        sync.RWMutex
	snapshots []MetricsSnapshot
	// For CPU calculation
	lastCPUTime   time.Time
	lastCPUSample float64
}

// NewMetricsHistory creates a new metrics history
func NewMetricsHistory() *MetricsHistory {
	return &MetricsHistory{
		snapshots: make([]MetricsSnapshot, 0, MetricsHistorySize),
	}
}

// AddSnapshot adds a new snapshot
func (h *MetricsHistory) AddSnapshot(snapshot MetricsSnapshot) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.snapshots = append(h.snapshots, snapshot)
	// Keep only last MetricsHistorySize snapshots
	if len(h.snapshots) > MetricsHistorySize {
		h.snapshots = h.snapshots[len(h.snapshots)-MetricsHistorySize:]
	}
}

// GetSnapshots returns all snapshots
func (h *MetricsHistory) GetSnapshots() []MetricsSnapshot {
	h.mu.RLock()
	defer h.mu.RUnlock()

	result := make([]MetricsSnapshot, len(h.snapshots))
	copy(result, h.snapshots)
	return result
}

// GetLatest returns the last N snapshots
func (h *MetricsHistory) GetLatest(n int) []MetricsSnapshot {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if n > len(h.snapshots) {
		n = len(h.snapshots)
	}
	if n == 0 {
		return nil
	}

	result := make([]MetricsSnapshot, n)
	copy(result, h.snapshots[len(h.snapshots)-n:])
	return result
}

// Clear clears all snapshots
func (h *MetricsHistory) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.snapshots = make([]MetricsSnapshot, 0, MetricsHistorySize)
}

// CollectSnapshot collects current metrics and returns a snapshot
func (h *MetricsHistory) CollectSnapshot(requestsDelta, jobsDelta int64, cpuPercent float64) MetricsSnapshot {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	snapshot := MetricsSnapshot{
		Timestamp:   time.Now(),
		MemoryAlloc: memStats.Alloc,
		MemorySys:   memStats.Sys,
		NumGC:       memStats.NumGC,
		Requests:    requestsDelta,
		Jobs:        jobsDelta,
		CPUPercent:  cpuPercent,
	}

	return snapshot
}

// SparklineData returns simplified data for sparkline charts
type SparklineData struct {
	Memory   []float64 `json:"memory"`   // MB values
	Requests []int64   `json:"requests"` // Request counts
	Jobs     []int64   `json:"jobs"`     // Scheduled job execution counts
	CPU      []float64 `json:"cpu"`      // CPU percent values
}

// GetSparklineData returns data formatted for sparkline charts
func (h *MetricsHistory) GetSparklineData() SparklineData {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data := SparklineData{
		Memory:   make([]float64, len(h.snapshots)),
		Requests: make([]int64, len(h.snapshots)),
		Jobs:     make([]int64, len(h.snapshots)),
		CPU:      make([]float64, len(h.snapshots)),
	}

	for i, s := range h.snapshots {
		data.Memory[i] = float64(s.MemoryAlloc) / 1024 / 1024 // Convert to MB
		data.Requests[i] = s.Requests
		data.Jobs[i] = s.Jobs
		data.CPU[i] = s.CPUPercent
	}

	return data
}
