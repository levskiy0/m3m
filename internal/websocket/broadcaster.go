package websocket

import (
	"context"
	"log/slog"
	"sort"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/levskiy0/m3m/internal/domain"
	"github.com/levskiy0/m3m/internal/runtime"
	"github.com/levskiy0/m3m/internal/service"
)

// Verify Broadcaster implements RuntimeStopHandler at compile time
var _ runtime.RuntimeStopHandler = (*Broadcaster)(nil)

// GoalStatsResponse represents aggregated goal statistics (same as handler)
type GoalStatsResponse struct {
	GoalID     string          `json:"goalID"`
	Value      int64           `json:"value"`
	TotalValue int64           `json:"totalValue,omitempty"`
	DailyStats []DailyStatItem `json:"dailyStats,omitempty"`
}

// DailyStatItem represents a single day's statistics
type DailyStatItem struct {
	Date  string `json:"date"`
	Value int64  `json:"value"`
}

// RuntimeManager interface for getting runtime stats
type RuntimeManager interface {
	GetStats(projectID primitive.ObjectID) (*runtime.RuntimeStats, error)
	IsRunning(projectID primitive.ObjectID) bool
	GetRunningProjects() []primitive.ObjectID
}

// ProjectService interface for updating project status
type ProjectService interface {
	UpdateStatus(ctx context.Context, projectID primitive.ObjectID, status domain.ProjectStatus) error
	SetRunningSource(ctx context.Context, projectID primitive.ObjectID, source string) error
}

// Broadcaster handles periodic event broadcasting
type Broadcaster struct {
	hub            *Hub
	logger         *slog.Logger
	runtimeManager RuntimeManager
	goalService    *service.GoalService
	projectService ProjectService

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

// SetProjectService sets the project service for status updates
func (b *Broadcaster) SetProjectService(ps ProjectService) {
	b.projectService = ps
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

// monitorLoop broadcasts monitor data to subscribers every 10 seconds
func (b *Broadcaster) monitorLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
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

// broadcastGoalsData sends aggregated goals stats to all project subscribers
func (b *Broadcaster) broadcastGoalsData() {
	if b.goalService == nil {
		return
	}

	if b.runtimeManager == nil {
		return
	}

	runningProjects := b.runtimeManager.GetRunningProjects()

	for _, projectID := range runningProjects {
		projectIDStr := projectID.Hex()

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

		// Get stats for last 7 days
		now := time.Now()
		startDate := now.AddDate(0, 0, -7)
		today := now.Format("2006-01-02")

		stats, err := b.goalService.GetStats(ctx, &domain.GoalStatsQuery{
			GoalIDs:   goalIDs,
			ProjectID: projectIDStr,
			StartDate: startDate.Format("2006-01-02"),
			EndDate:   today,
		})
		if err != nil {
			cancel()
			continue
		}

		// Get total values
		totalValues, err := b.goalService.GetTotalValues(ctx, goalIDs)
		cancel()
		if err != nil {
			continue
		}

		// Aggregate stats by goal ID (same logic as handler)
		goalStatsMap := make(map[string]*GoalStatsResponse)

		for _, stat := range stats {
			goalID := stat.GoalID.Hex()
			if _, ok := goalStatsMap[goalID]; !ok {
				goalStatsMap[goalID] = &GoalStatsResponse{
					GoalID:     goalID,
					DailyStats: []DailyStatItem{},
				}
			}

			gs := goalStatsMap[goalID]
			gs.DailyStats = append(gs.DailyStats, DailyStatItem{
				Date:  stat.Date,
				Value: stat.Value,
			})

			// Set current value (today's value)
			if stat.Date == today || stat.Date == "total" {
				gs.Value = stat.Value
			}
		}

		// Build result with total values
		result := make([]GoalStatsResponse, 0, len(goalStatsMap))
		for goalID, gs := range goalStatsMap {
			gs.TotalValue = totalValues[goalID]
			sort.Slice(gs.DailyStats, func(i, j int) bool {
				return gs.DailyStats[i].Date < gs.DailyStats[j].Date
			})
			result = append(result, *gs)
		}

		// Add empty responses for goals with no stats
		for _, goalID := range goalIDs {
			found := false
			for _, gs := range result {
				if gs.GoalID == goalID {
					found = true
					break
				}
			}
			if !found {
				result = append(result, GoalStatsResponse{
					GoalID:     goalID,
					Value:      0,
					TotalValue: totalValues[goalID],
					DailyStats: []DailyStatItem{},
				})
			}
		}

		b.hub.BroadcastToProject(projectIDStr, EventGoals, result)
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

// BroadcastActionStates broadcasts action state changes to project subscribers
func (b *Broadcaster) BroadcastActionStates(projectID string, states []domain.ActionRuntimeState) {
	if !b.hub.HasSubscribers(projectID) {
		return
	}

	b.hub.BroadcastToProject(projectID, EventActions, states)
}

// OnRuntimeStopped implements runtime.RuntimeStopHandler
// Called when a runtime stops (either normally or due to crash)
func (b *Broadcaster) OnRuntimeStopped(projectID primitive.ObjectID, reason runtime.CrashReason, message string) {
	projectIDStr := projectID.Hex()

	b.logger.Info("Runtime stopped, updating status and broadcasting",
		"project", projectIDStr,
		"reason", reason,
		"message", message,
	)

	// Update project status in database
	if b.projectService != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := b.projectService.UpdateStatus(ctx, projectID, domain.ProjectStatusStopped); err != nil {
			b.logger.Error("Failed to update project status on runtime stop",
				"project", projectIDStr,
				"error", err,
			)
		}
		if err := b.projectService.SetRunningSource(ctx, projectID, ""); err != nil {
			b.logger.Error("Failed to clear running source on runtime stop",
				"project", projectIDStr,
				"error", err,
			)
		}
	}

	// Broadcast running=false to all subscribers
	b.BroadcastRunning(projectIDStr, false)
}
