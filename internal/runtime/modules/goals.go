package modules

import (
	"context"
	"time"

	"github.com/dop251/goja"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"m3m/internal/domain"
	"m3m/internal/service"
)

// GoalInfo represents goal information returned to JS
type GoalInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

// GoalStatInfo represents goal statistics returned to JS
type GoalStatInfo struct {
	Date  string `json:"date"`
	Value int64  `json:"value"`
}

type GoalsModule struct {
	goalService *service.GoalService
	projectID   primitive.ObjectID
}

func NewGoalsModule(goalService *service.GoalService, projectID primitive.ObjectID) *GoalsModule {
	return &GoalsModule{
		goalService: goalService,
		projectID:   projectID,
	}
}

// Name returns the module name for JavaScript
func (g *GoalsModule) Name() string {
	return "$goals"
}

// Register registers the module into the JavaScript VM
func (g *GoalsModule) Register(vm interface{}) {
	vm.(*goja.Runtime).Set(g.Name(), map[string]interface{}{
		"increment": g.Increment,
		"getValue":  g.GetValue,
		"getStats":  g.GetStats,
		"list":      g.List,
		"get":       g.Get,
	})
}

// Increment increments a goal counter by the specified value (default 1)
func (g *GoalsModule) Increment(slug string, value ...int64) bool {
	if g.goalService == nil {
		return false
	}

	v := int64(1)
	if len(value) > 0 {
		v = value[0]
	}

	ctx := context.Background()
	err := g.goalService.Increment(ctx, slug, g.projectID, v)
	return err == nil
}

// GetValue returns the current value of a goal (total for counter, today's value for daily_counter)
func (g *GoalsModule) GetValue(slug string) int64 {
	if g.goalService == nil {
		return 0
	}

	ctx := context.Background()
	goal, err := g.goalService.GetBySlug(ctx, slug)
	if err != nil {
		return 0
	}

	// Determine the date to query
	var date string
	if goal.Type == domain.GoalTypeDailyCounter {
		date = time.Now().Format("2006-01-02")
	} else {
		date = "total"
	}

	// Get stats for this goal
	query := &domain.GoalStatsQuery{
		GoalIDs:   []string{goal.ID.Hex()},
		ProjectID: g.projectID.Hex(),
		StartDate: date,
		EndDate:   date,
	}

	stats, err := g.goalService.GetStats(ctx, query)
	if err != nil || len(stats) == 0 {
		return 0
	}

	return stats[0].Value
}

// GetStats returns statistics for a goal over a period of days
func (g *GoalsModule) GetStats(slug string, days int) []GoalStatInfo {
	if g.goalService == nil {
		return []GoalStatInfo{}
	}

	if days <= 0 {
		days = 7 // Default to 7 days
	}

	ctx := context.Background()
	goal, err := g.goalService.GetBySlug(ctx, slug)
	if err != nil {
		return []GoalStatInfo{}
	}

	// Calculate date range
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days+1)

	query := &domain.GoalStatsQuery{
		GoalIDs:   []string{goal.ID.Hex()},
		ProjectID: g.projectID.Hex(),
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endDate.Format("2006-01-02"),
	}

	stats, err := g.goalService.GetStats(ctx, query)
	if err != nil {
		return []GoalStatInfo{}
	}

	result := make([]GoalStatInfo, 0, len(stats))
	for _, stat := range stats {
		result = append(result, GoalStatInfo{
			Date:  stat.Date,
			Value: stat.Value,
		})
	}

	return result
}

// List returns all goals accessible by this project
func (g *GoalsModule) List() []GoalInfo {
	if g.goalService == nil {
		return []GoalInfo{}
	}

	ctx := context.Background()

	// Get project-specific goals
	projectGoals, err := g.goalService.GetProjectGoals(ctx, g.projectID)
	if err != nil {
		projectGoals = []*domain.Goal{}
	}

	// Get global goals that this project has access to
	globalGoals, err := g.goalService.GetGlobalGoals(ctx)
	if err != nil {
		globalGoals = []*domain.Goal{}
	}

	// Combine and filter
	result := make([]GoalInfo, 0)

	// Add project goals
	for _, goal := range projectGoals {
		result = append(result, GoalInfo{
			ID:          goal.ID.Hex(),
			Name:        goal.Name,
			Slug:        goal.Slug,
			Type:        string(goal.Type),
			Description: goal.Description,
			Color:       goal.Color,
		})
	}

	// Add global goals that this project has access to
	for _, goal := range globalGoals {
		hasAccess := false
		for _, allowed := range goal.AllowedProjects {
			if allowed == g.projectID {
				hasAccess = true
				break
			}
		}
		if hasAccess {
			result = append(result, GoalInfo{
				ID:          goal.ID.Hex(),
				Name:        goal.Name,
				Slug:        goal.Slug,
				Type:        string(goal.Type),
				Description: goal.Description,
				Color:       goal.Color,
			})
		}
	}

	return result
}

// Get returns information about a specific goal by slug
func (g *GoalsModule) Get(slug string) *GoalInfo {
	if g.goalService == nil {
		return nil
	}

	ctx := context.Background()
	goal, err := g.goalService.GetBySlug(ctx, slug)
	if err != nil {
		return nil
	}

	// Check access
	hasAccess := false
	if goal.ProjectRef != nil && *goal.ProjectRef == g.projectID {
		hasAccess = true
	} else {
		for _, allowed := range goal.AllowedProjects {
			if allowed == g.projectID {
				hasAccess = true
				break
			}
		}
	}

	if !hasAccess {
		return nil
	}

	return &GoalInfo{
		ID:          goal.ID.Hex(),
		Name:        goal.Name,
		Slug:        goal.Slug,
		Type:        string(goal.Type),
		Description: goal.Description,
		Color:       goal.Color,
	}
}

// GetSchema implements JSSchemaProvider
func (g *GoalsModule) GetSchema() JSModuleSchema {
	return JSModuleSchema{
		Name:        "$goals",
		Description: "Goal tracking and metrics management",
		Types: []JSTypeSchema{
			{
				Name:        "GoalInfo",
				Description: "Information about a goal",
				Fields: []JSParamSchema{
					{Name: "id", Type: "string", Description: "Goal ID"},
					{Name: "name", Type: "string", Description: "Goal name"},
					{Name: "slug", Type: "string", Description: "Goal slug identifier"},
					{Name: "type", Type: "string", Description: "Goal type (counter or daily_counter)"},
					{Name: "description", Type: "string", Description: "Goal description"},
					{Name: "color", Type: "string", Description: "Goal color"},
				},
			},
			{
				Name:        "GoalStatInfo",
				Description: "Goal statistics for a date",
				Fields: []JSParamSchema{
					{Name: "date", Type: "string", Description: "Date in YYYY-MM-DD format"},
					{Name: "value", Type: "number", Description: "Value for the date"},
				},
			},
		},
		Methods: []JSMethodSchema{
			{
				Name:        "increment",
				Description: "Increment a goal counter",
				Params: []JSParamSchema{
					{Name: "slug", Type: "string", Description: "Goal slug identifier"},
					{Name: "value", Type: "number", Description: "Amount to increment (default 1)", Optional: true},
				},
				Returns: &JSParamSchema{Type: "boolean"},
			},
			{
				Name:        "getValue",
				Description: "Get current value of a goal",
				Params:      []JSParamSchema{{Name: "slug", Type: "string", Description: "Goal slug identifier"}},
				Returns:     &JSParamSchema{Type: "number"},
			},
			{
				Name:        "getStats",
				Description: "Get statistics for a goal over a period",
				Params: []JSParamSchema{
					{Name: "slug", Type: "string", Description: "Goal slug identifier"},
					{Name: "days", Type: "number", Description: "Number of days (default 7)"},
				},
				Returns: &JSParamSchema{Type: "GoalStatInfo[]"},
			},
			{
				Name:        "list",
				Description: "List all accessible goals",
				Returns:     &JSParamSchema{Type: "GoalInfo[]"},
			},
			{
				Name:        "get",
				Description: "Get goal information by slug",
				Params:      []JSParamSchema{{Name: "slug", Type: "string", Description: "Goal slug identifier"}},
				Returns:     &JSParamSchema{Type: "GoalInfo | null"},
			},
		},
	}
}

// GetGoalsSchema returns the goals schema (static version)
func GetGoalsSchema() JSModuleSchema {
	return (&GoalsModule{}).GetSchema()
}
