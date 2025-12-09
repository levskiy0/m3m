package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GoalType string

const (
	GoalTypeCounter      GoalType = "counter"
	GoalTypeDailyCounter GoalType = "daily_counter"
)

type Goal struct {
	ID              primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	Name            string               `bson:"name" json:"name"`
	Slug            string               `bson:"slug" json:"slug"`
	Color           string               `bson:"color" json:"color"`
	Type            GoalType             `bson:"type" json:"type"`
	Description     string               `bson:"description" json:"description"`
	ProjectRef      *primitive.ObjectID  `bson:"project_ref" json:"project_ref"`
	AllowedProjects []primitive.ObjectID `bson:"allowed_projects" json:"allowed_projects"`
	GridSpan        int                  `bson:"grid_span" json:"gridSpan"`
	ShowTotal       bool                 `bson:"show_total" json:"showTotal"`
	Order           int                  `bson:"order" json:"order"`
	CreatedAt       time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time            `bson:"updated_at" json:"updated_at"`
}

type GoalStat struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	GoalID    primitive.ObjectID `bson:"goal_id" json:"goal_id"`
	ProjectID primitive.ObjectID `bson:"project_id" json:"project_id"`
	Date      string             `bson:"date" json:"date"`
	Value     int64              `bson:"value" json:"value"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

type CreateGoalRequest struct {
	Name            string   `json:"name" binding:"required"`
	Slug            string   `json:"slug" binding:"required"`
	Color           string   `json:"color"`
	Type            GoalType `json:"type" binding:"required"`
	Description     string   `json:"description"`
	AllowedProjects []string `json:"allowed_projects"`
	GridSpan        int      `json:"gridSpan"`
	ShowTotal       bool     `json:"showTotal"`
}

type UpdateGoalRequest struct {
	Name            *string   `json:"name"`
	Color           *string   `json:"color"`
	Description     *string   `json:"description"`
	AllowedProjects *[]string `json:"allowed_projects"`
	GridSpan        *int      `json:"gridSpan"`
	ShowTotal       *bool     `json:"showTotal"`
	Order           *int      `json:"order"`
}

type GoalStatsQuery struct {
	GoalIDs   []string `form:"goal_ids"`
	ProjectID string   `form:"project_id"`
	StartDate string   `form:"start_date"`
	EndDate   string   `form:"end_date"`
}

type IncrementGoalRequest struct {
	Value int64 `json:"value"`
}
