package service

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"m3m/internal/domain"
	"m3m/internal/repository"
)

var (
	ErrGoalIDRequired     = errors.New("goal ID is required for goal widgets")
	ErrWidgetNotInProject = errors.New("widget does not belong to this project")
)

type WidgetService struct {
	widgetRepo *repository.WidgetRepository
	goalRepo   *repository.GoalRepository
}

func NewWidgetService(widgetRepo *repository.WidgetRepository, goalRepo *repository.GoalRepository) *WidgetService {
	return &WidgetService{
		widgetRepo: widgetRepo,
		goalRepo:   goalRepo,
	}
}

func (s *WidgetService) Create(ctx context.Context, projectID primitive.ObjectID, req *domain.CreateWidgetRequest) (*domain.Widget, error) {
	gridSpan := req.GridSpan
	if gridSpan < 1 || gridSpan > 5 {
		gridSpan = 1
	}

	// Default type to goal for backward compatibility
	widgetType := req.Type
	if widgetType == "" {
		widgetType = domain.WidgetTypeGoal
	}

	widget := &domain.Widget{
		ProjectID: projectID,
		Type:      widgetType,
		Variant:   req.Variant,
		GridSpan:  gridSpan,
	}

	// Only parse and verify goal for goal-type widgets
	if widgetType == domain.WidgetTypeGoal {
		if req.GoalID == "" {
			return nil, ErrGoalIDRequired
		}
		goalID, err := primitive.ObjectIDFromHex(req.GoalID)
		if err != nil {
			return nil, err
		}
		// Verify goal exists
		if _, err := s.goalRepo.FindByID(ctx, goalID); err != nil {
			return nil, err
		}
		widget.GoalID = &goalID
	}

	if err := s.widgetRepo.Create(ctx, widget); err != nil {
		return nil, err
	}

	return widget, nil
}

func (s *WidgetService) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Widget, error) {
	return s.widgetRepo.FindByID(ctx, id)
}

func (s *WidgetService) GetByProject(ctx context.Context, projectID primitive.ObjectID) ([]*domain.Widget, error) {
	return s.widgetRepo.FindByProject(ctx, projectID)
}

func (s *WidgetService) Update(ctx context.Context, projectID, id primitive.ObjectID, req *domain.UpdateWidgetRequest) (*domain.Widget, error) {
	widget, err := s.widgetRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Verify widget belongs to the specified project
	if widget.ProjectID != projectID {
		return nil, ErrWidgetNotInProject
	}

	if req.Variant != nil {
		widget.Variant = *req.Variant
	}
	if req.GridSpan != nil {
		gridSpan := *req.GridSpan
		if gridSpan >= 1 && gridSpan <= 5 {
			widget.GridSpan = gridSpan
		}
	}
	if req.Order != nil {
		widget.Order = *req.Order
	}

	if err := s.widgetRepo.Update(ctx, widget); err != nil {
		return nil, err
	}

	return widget, nil
}

func (s *WidgetService) Delete(ctx context.Context, projectID, id primitive.ObjectID) error {
	widget, err := s.widgetRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Verify widget belongs to the specified project
	if widget.ProjectID != projectID {
		return ErrWidgetNotInProject
	}

	return s.widgetRepo.Delete(ctx, id)
}

func (s *WidgetService) Reorder(ctx context.Context, projectID primitive.ObjectID, req *domain.ReorderWidgetsRequest) error {
	widgetIDs := make([]primitive.ObjectID, 0, len(req.WidgetIDs))
	for _, idStr := range req.WidgetIDs {
		oid, err := primitive.ObjectIDFromHex(idStr)
		if err != nil {
			return err
		}
		widgetIDs = append(widgetIDs, oid)
	}

	return s.widgetRepo.Reorder(ctx, projectID, widgetIDs)
}
