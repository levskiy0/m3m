package handler

import (
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"m3m/internal/domain"
	"m3m/internal/middleware"
	"m3m/internal/service"
)

// GoalStatsResponse represents aggregated goal statistics for frontend
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

type GoalHandler struct {
	goalService    *service.GoalService
	projectService *service.ProjectService
}

func NewGoalHandler(goalService *service.GoalService, projectService *service.ProjectService) *GoalHandler {
	return &GoalHandler{
		goalService:    goalService,
		projectService: projectService,
	}
}

func (h *GoalHandler) Register(r *gin.RouterGroup, authMiddleware *middleware.AuthMiddleware) {
	// Global goals
	goals := r.Group("/goals")
	goals.Use(authMiddleware.Authenticate())
	{
		goals.GET("", h.ListGlobal)
		goals.POST("", authMiddleware.RequirePermission("manage_users"), h.CreateGlobal)
		goals.GET("/stats", h.GetStats)
		goals.GET("/:id", h.Get)
		goals.GET("/:id/stats", h.GetGoalStats)
		goals.PUT("/:id", authMiddleware.RequirePermission("manage_users"), h.Update)
		goals.DELETE("/:id", authMiddleware.RequirePermission("manage_users"), h.Delete)
	}

	// Project goals
	projectGoals := r.Group("/projects/:id/goals")
	projectGoals.Use(authMiddleware.Authenticate())
	{
		projectGoals.GET("", h.ListProject)
		projectGoals.POST("", h.CreateProject)
		projectGoals.PUT("/:goalId", h.UpdateProject)
		projectGoals.DELETE("/:goalId", h.DeleteProject)
		projectGoals.POST("/:goalId/reset", h.ResetProjectGoal)
	}
}

func (h *GoalHandler) ListGlobal(c *gin.Context) {
	goals, err := h.goalService.GetGlobalGoals(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, goals)
}

func (h *GoalHandler) CreateGlobal(c *gin.Context) {
	var req domain.CreateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	goal, err := h.goalService.CreateGlobal(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, goal)
}

func (h *GoalHandler) Get(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	goal, err := h.goalService.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "goal not found"})
		return
	}

	c.JSON(http.StatusOK, goal)
}

func (h *GoalHandler) Update(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req domain.UpdateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	goal, err := h.goalService.Update(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, goal)
}

func (h *GoalHandler) Delete(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.goalService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "goal deleted successfully"})
}

func (h *GoalHandler) GetStats(c *gin.Context) {
	// Parse goalIds from query (frontend sends comma-separated)
	goalIdsParam := c.Query("goalIds")
	if goalIdsParam == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "goalIds is required"})
		return
	}

	goalIds := strings.Split(goalIdsParam, ",")

	// Parse date range from query parameters, default to last 14 days
	startDateParam := c.Query("startDate")
	endDateParam := c.Query("endDate")

	var startDate, endDate string
	if startDateParam != "" && endDateParam != "" {
		startDate = startDateParam
		endDate = endDateParam
	} else {
		now := time.Now()
		endDate = now.Format("2006-01-02")
		startDate = now.AddDate(0, 0, -14).Format("2006-01-02")
	}

	query := &domain.GoalStatsQuery{
		GoalIDs:   goalIds,
		StartDate: startDate,
		EndDate:   endDate,
	}

	stats, err := h.goalService.GetStats(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get total values for all goals (sum across all time)
	totalValues, err := h.goalService.GetTotalValues(c.Request.Context(), goalIds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Aggregate stats by goal ID
	goalStatsMap := make(map[string]*GoalStatsResponse)
	today := time.Now().Format("2006-01-02")

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

		// Set current value (today's value for daily_counter, or total)
		if stat.Date == today || stat.Date == "total" {
			gs.Value = stat.Value
		}
	}

	// Set total values and convert map to slice
	result := make([]GoalStatsResponse, 0, len(goalStatsMap))
	for goalID, gs := range goalStatsMap {
		gs.TotalValue = totalValues[goalID]
		sort.Slice(gs.DailyStats, func(i, j int) bool {
			return gs.DailyStats[i].Date < gs.DailyStats[j].Date
		})
		result = append(result, *gs)
	}

	// Add empty responses for goals with no stats
	for _, goalID := range goalIds {
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

	c.JSON(http.StatusOK, result)
}

// GetGoalStats returns stats for a single goal
func (h *GoalHandler) GetGoalStats(c *gin.Context) {
	goalID := c.Param("id")
	if goalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "goal id is required"})
		return
	}

	// Calculate date range for last 14 days
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -14)

	query := &domain.GoalStatsQuery{
		GoalIDs:   []string{goalID},
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endDate.Format("2006-01-02"),
	}

	stats, err := h.goalService.GetStats(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get total value for this goal
	totalValues, err := h.goalService.GetTotalValues(c.Request.Context(), []string{goalID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := GoalStatsResponse{
		GoalID:     goalID,
		Value:      0,
		TotalValue: totalValues[goalID],
		DailyStats: []DailyStatItem{},
	}

	today := time.Now().Format("2006-01-02")
	for _, stat := range stats {
		response.DailyStats = append(response.DailyStats, DailyStatItem{
			Date:  stat.Date,
			Value: stat.Value,
		})
		if stat.Date == today || stat.Date == "total" {
			response.Value = stat.Value
		}
	}

	// Sort dailyStats by date
	sort.Slice(response.DailyStats, func(i, j int) bool {
		return response.DailyStats[i].Date < response.DailyStats[j].Date
	})

	c.JSON(http.StatusOK, response)
}

// Project goals

func (h *GoalHandler) ListProject(c *gin.Context) {
	projectID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	user := middleware.GetCurrentUser(c)
	if !h.projectService.CanUserAccess(c.Request.Context(), user.ID, projectID, user.IsRoot) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	goals, err := h.goalService.GetProjectGoals(c.Request.Context(), projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, goals)
}

func (h *GoalHandler) CreateProject(c *gin.Context) {
	projectID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	user := middleware.GetCurrentUser(c)
	if !h.projectService.CanUserAccess(c.Request.Context(), user.ID, projectID, user.IsRoot) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req domain.CreateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	goal, err := h.goalService.CreateForProject(c.Request.Context(), projectID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, goal)
}

func (h *GoalHandler) UpdateProject(c *gin.Context) {
	projectID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	goalID, err := primitive.ObjectIDFromHex(c.Param("goalId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid goal id"})
		return
	}

	user := middleware.GetCurrentUser(c)
	if !h.projectService.CanUserAccess(c.Request.Context(), user.ID, projectID, user.IsRoot) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req domain.UpdateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	goal, err := h.goalService.Update(c.Request.Context(), goalID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, goal)
}

func (h *GoalHandler) DeleteProject(c *gin.Context) {
	projectID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	goalID, err := primitive.ObjectIDFromHex(c.Param("goalId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid goal id"})
		return
	}

	user := middleware.GetCurrentUser(c)
	if !h.projectService.CanUserAccess(c.Request.Context(), user.ID, projectID, user.IsRoot) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.goalService.Delete(c.Request.Context(), goalID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "goal deleted successfully"})
}

func (h *GoalHandler) ResetProjectGoal(c *gin.Context) {
	projectID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	goalID, err := primitive.ObjectIDFromHex(c.Param("goalId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid goal id"})
		return
	}

	user := middleware.GetCurrentUser(c)
	if !h.projectService.CanUserAccess(c.Request.Context(), user.ID, projectID, user.IsRoot) {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := h.goalService.ResetStats(c.Request.Context(), goalID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "goal stats reset successfully"})
}
