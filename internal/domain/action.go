package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ActionState represents the current UI state of an action button
type ActionState string

const (
	ActionStateEnabled  ActionState = "enabled"
	ActionStateDisabled ActionState = "disabled"
	ActionStateLoading  ActionState = "loading"
)

// Action represents a project action definition stored in MongoDB
type Action struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProjectID primitive.ObjectID `bson:"project_id" json:"project_id"`
	Name      string             `bson:"name" json:"name"`
	Slug      string             `bson:"slug" json:"slug"`
	Order     int                `bson:"order" json:"order"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// ActionRuntimeState represents the runtime state of an action (not persisted)
type ActionRuntimeState struct {
	Slug  string      `json:"slug"`
	State ActionState `json:"state"`
}

// CreateActionRequest is the request for creating a new action
type CreateActionRequest struct {
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug" binding:"required"`
}

// UpdateActionRequest is the request for updating an action
type UpdateActionRequest struct {
	Name  *string `json:"name"`
	Order *int    `json:"order"`
}
