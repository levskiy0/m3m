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

// AggregationType defines how values are aggregated within an interval
type AggregationType int

const (
	AggregateAvg  AggregationType = iota // Average of values in interval
	AggregateSum                         // Sum of values in interval
	AggregateMax                         // Max value in interval
	AggregateLast                        // Last value in interval
)

// MetricState stores aggregated values for a single metric
// Example: cpuMetric := NewMetricState(20, 5*time.Minute, AggregateAvg)
type MetricState struct {
	mu       sync.RWMutex
	count    int             // number of data points to keep
	interval time.Duration   // aggregation interval per point
	aggType  AggregationType // how to aggregate values

	values      []float64 // finalized aggregated values
	currentSum  float64   // sum for current interval
	currentMax  float64   // max for current interval
	currentLast float64   // last value for current interval
	currentCnt  int       // samples count in current interval
	bucketStart time.Time // when current interval started
}

// NewMetricState creates a new MetricState for a single metric
func NewMetricState(count int, interval time.Duration, aggType AggregationType) *MetricState {
	return &MetricState{
		count:    count,
		interval: interval,
		aggType:  aggType,
		values:   make([]float64, 0, count),
	}
}

// Push adds a new value to the metric
func (m *MetricState) Push(value float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()

	// Check if we need to finalize current interval and start new one
	if m.currentCnt > 0 && now.Sub(m.bucketStart) >= m.interval {
		// Finalize current interval
		var aggValue float64
		switch m.aggType {
		case AggregateAvg:
			aggValue = m.currentSum / float64(m.currentCnt)
		case AggregateSum:
			aggValue = m.currentSum
		case AggregateMax:
			aggValue = m.currentMax
		case AggregateLast:
			aggValue = m.currentLast
		}

		m.values = append(m.values, aggValue)
		if len(m.values) > m.count {
			m.values = m.values[len(m.values)-m.count:]
		}

		// Reset for new interval
		m.currentSum = 0
		m.currentMax = 0
		m.currentLast = 0
		m.currentCnt = 0
		m.bucketStart = now
	}

	// Start new interval if needed
	if m.currentCnt == 0 {
		m.bucketStart = now
	}

	// Add to current interval
	m.currentSum += value
	m.currentLast = value
	if value > m.currentMax {
		m.currentMax = value
	}
	m.currentCnt++
}

// GetData returns aggregated values for charts
func (m *MetricState) GetData() []float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Copy finalized values
	result := make([]float64, len(m.values), len(m.values)+1)
	copy(result, m.values)

	// Add current interval if has data
	if m.currentCnt > 0 {
		var aggValue float64
		switch m.aggType {
		case AggregateAvg:
			aggValue = m.currentSum / float64(m.currentCnt)
		case AggregateSum:
			aggValue = m.currentSum
		case AggregateMax:
			aggValue = m.currentMax
		case AggregateLast:
			aggValue = m.currentLast
		}
		result = append(result, aggValue)
	}

	return result
}

// Clear resets all data
func (m *MetricState) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.values = make([]float64, 0, m.count)
	m.currentSum = 0
	m.currentMax = 0
	m.currentLast = 0
	m.currentCnt = 0
}

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
	// Aggregated metrics for charts (20 points, 5 min each)
	memoryChart   *MetricState
	requestsChart *MetricState
	jobsChart     *MetricState
	cpuChart      *MetricState
}

// NewMetricsHistory creates a new metrics history
func NewMetricsHistory() *MetricsHistory {
	return &MetricsHistory{
		snapshots:     make([]MetricsSnapshot, 0, MetricsHistorySize),
		memoryChart:   NewMetricState(20, 5*time.Minute, AggregateAvg),
		requestsChart: NewMetricState(20, 5*time.Minute, AggregateSum),
		jobsChart:     NewMetricState(20, 5*time.Minute, AggregateSum),
		cpuChart:      NewMetricState(20, 5*time.Minute, AggregateAvg),
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

	// Push to individual chart metrics
	h.memoryChart.Push(float64(snapshot.MemoryAlloc) / 1024 / 1024)
	h.requestsChart.Push(float64(snapshot.Requests))
	h.jobsChart.Push(float64(snapshot.Jobs))
	h.cpuChart.Push(snapshot.CPUPercent)
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
	h.memoryChart.Clear()
	h.requestsChart.Clear()
	h.jobsChart.Clear()
	h.cpuChart.Clear()
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
	Memory   []float64 `json:"memory"`   // MB values (avg per interval)
	Requests []float64 `json:"requests"` // Request counts (sum per interval)
	Jobs     []float64 `json:"jobs"`     // Scheduled job counts (sum per interval)
	CPU      []float64 `json:"cpu"`      // CPU percent values (avg per interval)
}

// GetSparklineData returns aggregated data for sparkline charts
// Returns max 20 data points with 5-minute aggregation intervals
func (h *MetricsHistory) GetSparklineData() SparklineData {
	return SparklineData{
		Memory:   h.memoryChart.GetData(),
		Requests: h.requestsChart.GetData(),
		Jobs:     h.jobsChart.GetData(),
		CPU:      h.cpuChart.GetData(),
	}
}
