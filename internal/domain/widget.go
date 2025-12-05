package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type WidgetVariant string

const (
	WidgetVariantMini     WidgetVariant = "mini"     // Small card with sparkline
	WidgetVariantDetailed WidgetVariant = "detailed" // Full card with chart
	WidgetVariantSimple   WidgetVariant = "simple"   // Simple counter without chart
)

type Widget struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID primitive.ObjectID `bson:"project_id" json:"projectId"`
	GoalID    primitive.ObjectID `bson:"goal_id" json:"goalId"`
	Variant   WidgetVariant      `bson:"variant" json:"variant"`
	Order     int                `bson:"order" json:"order"`
	CreatedAt time.Time          `bson:"created_at" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updatedAt"`
}

type CreateWidgetRequest struct {
	GoalID  string        `json:"goalId" binding:"required"`
	Variant WidgetVariant `json:"variant" binding:"required"`
}

type UpdateWidgetRequest struct {
	Variant *WidgetVariant `json:"variant"`
	Order   *int           `json:"order"`
}

type ReorderWidgetsRequest struct {
	WidgetIDs []string `json:"widgetIds" binding:"required"`
}
