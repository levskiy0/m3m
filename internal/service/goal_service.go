package service

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/levskiy0/m3m/internal/domain"
	"github.com/levskiy0/m3m/internal/repository"
)

type GoalService struct {
	goalRepo *repository.GoalRepository
}

func NewGoalService(goalRepo *repository.GoalRepository) *GoalService {
	return &GoalService{
		goalRepo: goalRepo,
	}
}

func (s *GoalService) CreateGlobal(ctx context.Context, req *domain.CreateGoalRequest) (*domain.Goal, error) {
	allowedProjects := make([]primitive.ObjectID, 0)
	for _, idStr := range req.AllowedProjects {
		oid, err := primitive.ObjectIDFromHex(idStr)
		if err == nil {
			allowedProjects = append(allowedProjects, oid)
		}
	}

	goal := &domain.Goal{
		Name:            req.Name,
		Slug:            req.Slug,
		Color:           req.Color,
		Type:            req.Type,
		Description:     req.Description,
		ProjectRef:      nil,
		AllowedProjects: allowedProjects,
	}

	if err := s.goalRepo.Create(ctx, goal); err != nil {
		return nil, err
	}

	return goal, nil
}

func (s *GoalService) CreateForProject(ctx context.Context, projectID primitive.ObjectID, req *domain.CreateGoalRequest) (*domain.Goal, error) {
	gridSpan := req.GridSpan
	if gridSpan == 0 {
		gridSpan = 1
	}

	goal := &domain.Goal{
		Name:            req.Name,
		Slug:            fmt.Sprintf("%s-%s", projectID.Hex()[:8], req.Slug),
		Color:           req.Color,
		Type:            req.Type,
		Description:     req.Description,
		ProjectRef:      &projectID,
		AllowedProjects: []primitive.ObjectID{projectID},
		GridSpan:        gridSpan,
		ShowTotal:       req.ShowTotal,
	}

	if err := s.goalRepo.Create(ctx, goal); err != nil {
		return nil, err
	}

	return goal, nil
}

func (s *GoalService) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Goal, error) {
	return s.goalRepo.FindByID(ctx, id)
}

func (s *GoalService) GetBySlug(ctx context.Context, slug string) (*domain.Goal, error) {
	return s.goalRepo.FindBySlug(ctx, slug)
}

func (s *GoalService) GetGlobalGoals(ctx context.Context) ([]*domain.Goal, error) {
	return s.goalRepo.FindGlobalGoals(ctx)
}

func (s *GoalService) GetProjectGoals(ctx context.Context, projectID primitive.ObjectID) ([]*domain.Goal, error) {
	return s.goalRepo.FindByProject(ctx, projectID)
}

func (s *GoalService) Update(ctx context.Context, id primitive.ObjectID, req *domain.UpdateGoalRequest) (*domain.Goal, error) {
	goal, err := s.goalRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		goal.Name = *req.Name
	}
	if req.Color != nil {
		goal.Color = *req.Color
	}
	if req.Description != nil {
		goal.Description = *req.Description
	}
	if req.AllowedProjects != nil {
		allowedProjects := make([]primitive.ObjectID, 0)
		for _, idStr := range *req.AllowedProjects {
			oid, err := primitive.ObjectIDFromHex(idStr)
			if err == nil {
				allowedProjects = append(allowedProjects, oid)
			}
		}
		goal.AllowedProjects = allowedProjects
	}
	if req.GridSpan != nil {
		goal.GridSpan = *req.GridSpan
	}
	if req.ShowTotal != nil {
		goal.ShowTotal = *req.ShowTotal
	}
	if req.Order != nil {
		goal.Order = *req.Order
	}

	if err := s.goalRepo.Update(ctx, goal); err != nil {
		return nil, err
	}

	return goal, nil
}

func (s *GoalService) Delete(ctx context.Context, id primitive.ObjectID) error {
	return s.goalRepo.Delete(ctx, id)
}

func (s *GoalService) Increment(ctx context.Context, goalSlug string, projectID primitive.ObjectID, value int64) error {
	goal, err := s.goalRepo.FindBySlug(ctx, goalSlug)
	if err != nil {
		return err
	}

	// Check if project has access to this goal
	hasAccess := false
	if goal.ProjectRef != nil && *goal.ProjectRef == projectID {
		hasAccess = true
	} else {
		for _, allowed := range goal.AllowedProjects {
			if allowed == projectID {
				hasAccess = true
				break
			}
		}
	}

	if !hasAccess {
		return fmt.Errorf("project does not have access to goal %s", goalSlug)
	}

	var date string
	if goal.Type == domain.GoalTypeDailyCounter {
		date = time.Now().UTC().Format("2006-01-02")
	} else {
		date = "total"
	}

	return s.goalRepo.IncrementStat(ctx, goal.ID, projectID, date, value)
}

func (s *GoalService) GetStats(ctx context.Context, query *domain.GoalStatsQuery) ([]*domain.GoalStat, error) {
	return s.goalRepo.GetStats(ctx, query)
}

func (s *GoalService) GetTotalValues(ctx context.Context, goalIDs []string) (map[string]int64, error) {
	return s.goalRepo.GetTotalValues(ctx, goalIDs)
}

func (s *GoalService) ResetStats(ctx context.Context, id primitive.ObjectID) error {
	return s.goalRepo.ResetStats(ctx, id)
}
