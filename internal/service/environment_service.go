package service

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/levskiy0/m3m/internal/domain"
	"github.com/levskiy0/m3m/internal/repository"
)

type EnvironmentService struct {
	envRepo *repository.EnvironmentRepository
}

func NewEnvironmentService(envRepo *repository.EnvironmentRepository) *EnvironmentService {
	return &EnvironmentService{
		envRepo: envRepo,
	}
}

func (s *EnvironmentService) Create(ctx context.Context, projectID primitive.ObjectID, req *domain.CreateEnvVarRequest) (*domain.EnvVar, error) {
	envVar := &domain.EnvVar{
		ProjectID: projectID,
		Key:       req.Key,
		Type:      req.Type,
		Value:     req.Value,
	}

	if err := s.envRepo.Create(ctx, envVar); err != nil {
		return nil, err
	}

	return envVar, nil
}

func (s *EnvironmentService) GetByProject(ctx context.Context, projectID primitive.ObjectID) ([]*domain.EnvVar, error) {
	return s.envRepo.FindByProject(ctx, projectID)
}

func (s *EnvironmentService) GetByKey(ctx context.Context, projectID primitive.ObjectID, key string) (*domain.EnvVar, error) {
	return s.envRepo.FindByKey(ctx, projectID, key)
}

func (s *EnvironmentService) Update(ctx context.Context, projectID primitive.ObjectID, key string, req *domain.UpdateEnvVarRequest) (*domain.EnvVar, error) {
	envVar, err := s.envRepo.FindByKey(ctx, projectID, key)
	if err != nil {
		return nil, err
	}

	if req.Type != nil {
		envVar.Type = *req.Type
	}
	if req.Value != nil {
		envVar.Value = *req.Value
	}

	if err := s.envRepo.Update(ctx, envVar); err != nil {
		return nil, err
	}

	return envVar, nil
}

func (s *EnvironmentService) Delete(ctx context.Context, projectID primitive.ObjectID, key string) error {
	return s.envRepo.DeleteByKey(ctx, projectID, key)
}

func (s *EnvironmentService) GetEnvMap(ctx context.Context, projectID primitive.ObjectID) (map[string]interface{}, error) {
	envVars, err := s.envRepo.FindByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	for _, v := range envVars {
		result[v.Key] = v.Value
	}

	return result, nil
}

func (s *EnvironmentService) BulkUpdate(ctx context.Context, projectID primitive.ObjectID, req *domain.BulkUpdateEnvVarRequest) ([]*domain.EnvVar, error) {
	envVars := make([]*domain.EnvVar, len(req.Items))
	for i, item := range req.Items {
		envVars[i] = &domain.EnvVar{
			Key:   item.Key,
			Type:  item.Type,
			Value: item.Value,
			Order: item.Order,
		}
	}

	return s.envRepo.BulkUpdate(ctx, projectID, envVars)
}
