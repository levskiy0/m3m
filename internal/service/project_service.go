package service

import (
	"context"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"m3m/internal/config"
	"m3m/internal/domain"
	"m3m/internal/repository"
)

type ProjectService struct {
	projectRepo *repository.ProjectRepository
	config      *config.Config
}

func NewProjectService(projectRepo *repository.ProjectRepository, config *config.Config) *ProjectService {
	return &ProjectService{
		projectRepo: projectRepo,
		config:      config,
	}
}

func (s *ProjectService) Create(ctx context.Context, req *domain.CreateProjectRequest, ownerID primitive.ObjectID) (*domain.Project, error) {
	project := &domain.Project{
		Name:    req.Name,
		Slug:    req.Slug,
		Color:   req.Color,
		OwnerID: ownerID,
		Members: []primitive.ObjectID{},
		APIKey:  s.generateAPIKey(),
		Status:  domain.ProjectStatusStopped,
	}

	if err := s.projectRepo.Create(ctx, project); err != nil {
		return nil, err
	}

	// Create storage directory for project
	storagePath := filepath.Join(s.config.Storage.Path, project.ID.Hex(), "storage")
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		// Rollback project creation on storage error
		s.projectRepo.Delete(ctx, project.ID)
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Project, error) {
	return s.projectRepo.FindByID(ctx, id)
}

func (s *ProjectService) GetBySlug(ctx context.Context, slug string) (*domain.Project, error) {
	return s.projectRepo.FindBySlug(ctx, slug)
}

func (s *ProjectService) GetAll(ctx context.Context) ([]*domain.Project, error) {
	return s.projectRepo.FindAll(ctx)
}

func (s *ProjectService) GetByUser(ctx context.Context, userID primitive.ObjectID, isRoot bool) ([]*domain.Project, error) {
	if isRoot {
		return s.projectRepo.FindAll(ctx)
	}
	return s.projectRepo.FindByMember(ctx, userID)
}

func (s *ProjectService) Update(ctx context.Context, id primitive.ObjectID, req *domain.UpdateProjectRequest) (*domain.Project, error) {
	project, err := s.projectRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		project.Name = *req.Name
	}
	if req.Color != nil {
		project.Color = *req.Color
	}
	if req.AutoStart != nil {
		project.AutoStart = *req.AutoStart
	}

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) Delete(ctx context.Context, id primitive.ObjectID) error {
	project, err := s.projectRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete storage directory
	storagePath := filepath.Join(s.config.Storage.Path, project.ID.Hex())
	os.RemoveAll(storagePath)

	return s.projectRepo.Delete(ctx, id)
}

func (s *ProjectService) RegenerateAPIKey(ctx context.Context, id primitive.ObjectID) (*domain.Project, error) {
	project, err := s.projectRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	project.APIKey = s.generateAPIKey()

	if err := s.projectRepo.Update(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) AddMember(ctx context.Context, projectID, userID primitive.ObjectID) error {
	return s.projectRepo.AddMember(ctx, projectID, userID)
}

func (s *ProjectService) RemoveMember(ctx context.Context, projectID, userID primitive.ObjectID) error {
	return s.projectRepo.RemoveMember(ctx, projectID, userID)
}

func (s *ProjectService) CanUserAccess(ctx context.Context, userID primitive.ObjectID, projectID primitive.ObjectID, isRoot bool) bool {
	if isRoot {
		return true
	}

	project, err := s.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return false
	}

	if project.OwnerID == userID {
		return true
	}

	for _, member := range project.Members {
		if member == userID {
			return true
		}
	}

	return false
}

func (s *ProjectService) generateAPIKey() string {
	return uuid.New().String()
}

func (s *ProjectService) UpdateStatus(ctx context.Context, id primitive.ObjectID, status domain.ProjectStatus) error {
	return s.projectRepo.UpdateStatus(ctx, id, status)
}

func (s *ProjectService) SetActiveRelease(ctx context.Context, id primitive.ObjectID, version string) error {
	return s.projectRepo.SetActiveRelease(ctx, id, version)
}

func (s *ProjectService) GetByStatus(ctx context.Context, status domain.ProjectStatus) ([]*domain.Project, error) {
	return s.projectRepo.FindByStatus(ctx, status)
}
