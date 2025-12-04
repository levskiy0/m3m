package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProjectStatus string

const (
	ProjectStatusStopped ProjectStatus = "stopped"
	ProjectStatusRunning ProjectStatus = "running"
)

type Project struct {
	ID            primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name          string               `bson:"name" json:"name"`
	Slug          string               `bson:"slug" json:"slug"`
	Color         string               `bson:"color" json:"color"`
	OwnerID       primitive.ObjectID   `bson:"owner_id" json:"owner_id"`
	Members       []primitive.ObjectID `bson:"members" json:"members"`
	APIKey        string               `bson:"api_key" json:"api_key"`
	Status        ProjectStatus        `bson:"status" json:"status"`
	AutoStart     bool                 `bson:"auto_start" json:"auto_start"`
	ActiveRelease string               `bson:"active_release" json:"active_release"`
	CreatedAt     time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time            `bson:"updated_at" json:"updated_at"`
}

type CreateProjectRequest struct {
	Name  string `json:"name" binding:"required"`
	Slug  string `json:"slug" binding:"required"`
	Color string `json:"color"`
}

type UpdateProjectRequest struct {
	Name      *string `json:"name"`
	Color     *string `json:"color"`
	AutoStart *bool   `json:"auto_start"`
}

type AddMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
}
