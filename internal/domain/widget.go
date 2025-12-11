package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WidgetType string

const (
	WidgetTypeGoal     WidgetType = "goal"     // Goal-based widget
	WidgetTypeMemory   WidgetType = "memory"   // Memory usage widget
	WidgetTypeRequests WidgetType = "requests" // Request count widget
	WidgetTypeCPU      WidgetType = "cpu"      // CPU usage widget
	WidgetTypeStorage  WidgetType = "storage"  // Storage usage widget
	WidgetTypeDatabase WidgetType = "database" // Database size widget
	WidgetTypeUptime   WidgetType = "uptime"   // Uptime widget
	WidgetTypeJobs     WidgetType = "jobs"     // Scheduled jobs widget
	WidgetTypeAction   WidgetType = "action"   // Action button widget
)

type WidgetVariant string

const (
	WidgetVariantMini     WidgetVariant = "mini"     // Small card with sparkline
	WidgetVariantDetailed WidgetVariant = "detailed" // Full card with chart
	WidgetVariantSimple   WidgetVariant = "simple"   // Simple counter without chart
)

type Widget struct {
	ID        primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	ProjectID primitive.ObjectID  `bson:"project_id" json:"projectId"`
	Type      WidgetType          `bson:"type" json:"type"`
	GoalID    *primitive.ObjectID `bson:"goal_id,omitempty" json:"goalId,omitempty"`     // Only for goal widgets
	ActionID  *primitive.ObjectID `bson:"action_id,omitempty" json:"actionId,omitempty"` // Only for action widgets
	Variant   WidgetVariant       `bson:"variant" json:"variant"`
	GridSpan  int                 `bson:"grid_span" json:"gridSpan"`
	Order     int                 `bson:"order" json:"order"`
	CreatedAt time.Time           `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time           `bson:"updated_at" json:"updatedAt"`
}

type CreateWidgetRequest struct {
	Type     WidgetType    `json:"type" binding:"required"`
	GoalID   string        `json:"goalId"`   // Required only for goal widgets
	ActionID string        `json:"actionId"` // Required only for action widgets
	Variant  WidgetVariant `json:"variant" binding:"required"`
	GridSpan int           `json:"gridSpan"`
}

type UpdateWidgetRequest struct {
	Variant  *WidgetVariant `json:"variant"`
	GridSpan *int           `json:"gridSpan"`
	Order    *int           `json:"order"`
}

type ReorderWidgetsRequest struct {
	WidgetIDs []string `json:"widgetIds" binding:"required"`
}
