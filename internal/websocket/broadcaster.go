package websocket

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"m3m/internal/domain"
	"m3m/internal/runtime"
	"m3m/internal/service"
)

// RuntimeManager interface for getting runtime stats
type RuntimeManager interface {
	GetStats(projectID primitive.ObjectID) (*runtime.RuntimeStats, error)
	IsRunning(projectID primitive.ObjectID) bool
	GetRunningProjects() []primitive.ObjectID
}

// Broadcaster handles periodic event broadcasting
type Broadcaster struct {
	hub            *Hub
	logger         *slog.Logger
	runtimeManager RuntimeManager
	goalService    *service.GoalService

	// Track last log update times per project
	lastLogUpdate map[string]time.Time
	logMu         sync.RWMutex

	cancel context.CancelFunc
}

// NewBroadcaster creates a new event broadcaster
func NewBroadcaster(
	hub *Hub,
	logger *slog.Logger,
	goalService *service.GoalService,
) *Broadcaster {
	return &Broadcaster{
		hub:           hub,
		logger:        logger,
		goalService:   goalService,
		lastLogUpdate: make(map[string]time.Time),
	}
}

// SetRuntimeManager sets the runtime manager (called after runtime.Manager is created)
func (b *Broadcaster) SetRuntimeManager(rm RuntimeManager) {
	b.runtimeManager = rm
}

// Start starts the broadcaster goroutines
func (b *Broadcaster) Start(_ context.Context) {
	// Use background context - broadcaster should run for entire app lifetime
	bgCtx, cancel := context.WithCancel(context.Background())
	b.cancel = cancel

	// Start monitor broadcast (every 1 second)
	go b.monitorLoop(bgCtx)

	// Start goals broadcast (every 30 seconds)
	go b.goalsLoop(bgCtx)

	b.logger.Info("WebSocket broadcaster started")
}

// Stop stops the broadcaster
func (b *Broadcaster) Stop() {
	if b.cancel != nil {
		b.cancel()
	}
	b.logger.Info("WebSocket broadcaster stopped")
}

// monitorLoop broadcasts monitor data to subscribers every second
func (b *Broadcaster) monitorLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			b.broadcastMonitorData()
		}
	}
}

// goalsLoop broadcasts goals data to subscribers
func (b *Broadcaster) goalsLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			b.broadcastGoalsData()
		}
	}
}

// broadcastMonitorData sends monitor data to all project subscribers
func (b *Broadcaster) broadcastMonitorData() {
	if b.runtimeManager == nil {
		return
	}

	runningProjects := b.runtimeManager.GetRunningProjects()

	for _, projectID := range runningProjects {
		projectIDStr := projectID.Hex()

		// Only broadcast if there are subscribers
		if !b.hub.HasSubscribers(projectIDStr) {
			continue
		}

		stats, err := b.runtimeManager.GetStats(projectID)
		if err != nil {
			continue
		}

		b.hub.BroadcastToProject(projectIDStr, EventMonitor, stats)
	}
}

// broadcastGoalsData sends goals stats to all project subscribers
func (b *Broadcaster) broadcastGoalsData() {
	if b.goalService == nil {
		return
	}

	// Get all running projects
	if b.runtimeManager == nil {
		return
	}

	runningProjects := b.runtimeManager.GetRunningProjects()

	for _, projectID := range runningProjects {
		projectIDStr := projectID.Hex()

		// Only broadcast if there are subscribers
		if !b.hub.HasSubscribers(projectIDStr) {
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		// Get goals for project
		goals, err := b.goalService.GetProjectGoals(ctx, projectID)
		if err != nil {
			cancel()
			continue
		}

		if len(goals) == 0 {
			cancel()
			continue
		}

		// Get goal IDs
		goalIDs := make([]string, len(goals))
		for i, g := range goals {
			goalIDs[i] = g.ID.Hex()
		}

		// Get stats for last 14 days
		now := time.Now()
		startDate := now.AddDate(0, 0, -14)

		stats, err := b.goalService.GetStats(ctx, &domain.GoalStatsQuery{
			GoalIDs:   goalIDs,
			ProjectID: projectIDStr,
			StartDate: startDate.Format("2006-01-02"),
			EndDate:   now.Format("2006-01-02"),
		})
		cancel()

		if err != nil {
			continue
		}

		b.hub.BroadcastToProject(projectIDStr, EventGoals, stats)
	}
}

// BroadcastRunning broadcasts running status change
func (b *Broadcaster) BroadcastRunning(projectID string, running bool) {
	b.hub.BroadcastToProject(projectID, EventRunning, map[string]bool{
		"running": running,
	})
}

// BroadcastLogUpdate notifies that new logs are available
func (b *Broadcaster) BroadcastLogUpdate(projectID string) {
	b.logMu.RLock()
	// lastUpdate, exists := b.lastLogUpdate[projectID]
	b.logMu.RUnlock()

	b.logMu.Lock()
	b.lastLogUpdate[projectID] = time.Now()
	b.logMu.Unlock()

	b.hub.BroadcastToProject(projectID, EventLog, map[string]bool{
		"hasNewLogs": true,
	})
}

// BroadcastMonitorNow immediately broadcasts monitor data for a project
func (b *Broadcaster) BroadcastMonitorNow(projectID primitive.ObjectID) {
	if b.runtimeManager == nil {
		return
	}

	projectIDStr := projectID.Hex()

	if !b.hub.HasSubscribers(projectIDStr) {
		return
	}

	stats, err := b.runtimeManager.GetStats(projectID)
	if err != nil {
		return
	}

	b.hub.BroadcastToProject(projectIDStr, EventMonitor, stats)
}
