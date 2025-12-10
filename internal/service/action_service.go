package service

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/levskiy0/m3m/internal/domain"
	"github.com/levskiy0/m3m/internal/repository"
)

type ActionService struct {
	actionRepo *repository.ActionRepository
}

func NewActionService(actionRepo *repository.ActionRepository) *ActionService {
	return &ActionService{
		actionRepo: actionRepo,
	}
}

func (s *ActionService) Create(ctx context.Context, projectID primitive.ObjectID, req *domain.CreateActionRequest) (*domain.Action, error) {
	// Get max order to add new action at the end
	maxOrder, err := s.actionRepo.GetMaxOrder(ctx, projectID)
	if err != nil {
		return nil, err
	}

	action := &domain.Action{
		ProjectID: projectID,
		Name:      req.Name,
		Slug:      req.Slug,
		Order:     maxOrder + 1,
	}

	if err := s.actionRepo.Create(ctx, action); err != nil {
		return nil, err
	}

	return action, nil
}

func (s *ActionService) GetByProject(ctx context.Context, projectID primitive.ObjectID) ([]*domain.Action, error) {
	return s.actionRepo.FindByProjectID(ctx, projectID)
}

func (s *ActionService) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Action, error) {
	return s.actionRepo.FindByID(ctx, id)
}

func (s *ActionService) GetBySlug(ctx context.Context, projectID primitive.ObjectID, slug string) (*domain.Action, error) {
	return s.actionRepo.FindBySlug(ctx, projectID, slug)
}

func (s *ActionService) Update(ctx context.Context, id primitive.ObjectID, req *domain.UpdateActionRequest) (*domain.Action, error) {
	action, err := s.actionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		action.Name = *req.Name
	}
	if req.Order != nil {
		action.Order = *req.Order
	}

	if err := s.actionRepo.Update(ctx, action); err != nil {
		return nil, err
	}

	return action, nil
}

func (s *ActionService) Delete(ctx context.Context, id primitive.ObjectID) error {
	return s.actionRepo.Delete(ctx, id)
}

func (s *ActionService) DeleteByProject(ctx context.Context, projectID primitive.ObjectID) error {
	return s.actionRepo.DeleteByProject(ctx, projectID)
}
