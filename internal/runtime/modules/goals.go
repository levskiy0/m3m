package modules

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"m3m/internal/service"
)

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

func (g *GoalsModule) Increment(slug string, value ...int64) bool {
	v := int64(1)
	if len(value) > 0 {
		v = value[0]
	}

	ctx := context.Background()
	err := g.goalService.Increment(ctx, slug, g.projectID, v)
	return err == nil
}
