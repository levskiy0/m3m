package service

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"m3m/internal/domain"
	"m3m/internal/repository"
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
	goalID, err := primitive.ObjectIDFromHex(req.GoalID)
	if err != nil {
		return nil, err
	}

	// Verify goal exists
	if _, err := s.goalRepo.FindByID(ctx, goalID); err != nil {
		return nil, err
	}

	gridSpan := req.GridSpan
	if gridSpan < 1 || gridSpan > 5 {
		gridSpan = 1
	}

	widget := &domain.Widget{
		ProjectID: projectID,
		GoalID:    goalID,
		Variant:   req.Variant,
		GridSpan:  gridSpan,
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

func (s *WidgetService) Update(ctx context.Context, id primitive.ObjectID, req *domain.UpdateWidgetRequest) (*domain.Widget, error) {
	widget, err := s.widgetRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
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

func (s *WidgetService) Delete(ctx context.Context, id primitive.ObjectID) error {
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
