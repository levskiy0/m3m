package service

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"m3m/internal/domain"
	"m3m/internal/repository"
)

type ModelService struct {
	modelRepo *repository.ModelRepository
}

func NewModelService(modelRepo *repository.ModelRepository) *ModelService {
	return &ModelService{
		modelRepo: modelRepo,
	}
}

func (s *ModelService) Create(ctx context.Context, projectID primitive.ObjectID, req *domain.CreateModelRequest) (*domain.Model, error) {
	model := &domain.Model{
		ProjectID: projectID,
		Name:      req.Name,
		Slug:      req.Slug,
		Fields:    req.Fields,
	}

	if req.TableConfig != nil {
		model.TableConfig = *req.TableConfig
	} else {
		// Default table config - show all fields
		columns := make([]string, len(req.Fields))
		for i, f := range req.Fields {
			columns[i] = f.Key
		}
		model.TableConfig = domain.TableConfig{
			Columns:     columns,
			Filters:     []string{},
			SortColumns: columns,
		}
	}

	if req.FormConfig != nil {
		model.FormConfig = *req.FormConfig
	} else {
		// Default form config
		fieldOrder := make([]string, len(req.Fields))
		for i, f := range req.Fields {
			fieldOrder[i] = f.Key
		}
		model.FormConfig = domain.FormConfig{
			FieldOrder:   fieldOrder,
			HiddenFields: []string{},
			FieldViews:   make(map[string]string),
		}
	}

	if err := s.modelRepo.Create(ctx, model); err != nil {
		return nil, err
	}

	return model, nil
}

func (s *ModelService) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Model, error) {
	return s.modelRepo.FindByID(ctx, id)
}

func (s *ModelService) GetBySlug(ctx context.Context, projectID primitive.ObjectID, slug string) (*domain.Model, error) {
	return s.modelRepo.FindBySlug(ctx, projectID, slug)
}

func (s *ModelService) GetByProject(ctx context.Context, projectID primitive.ObjectID) ([]*domain.Model, error) {
	return s.modelRepo.FindByProject(ctx, projectID)
}

func (s *ModelService) Update(ctx context.Context, id primitive.ObjectID, req *domain.UpdateModelRequest) (*domain.Model, error) {
	model, err := s.modelRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		model.Name = *req.Name
	}
	if req.Fields != nil {
		model.Fields = *req.Fields
	}
	if req.TableConfig != nil {
		model.TableConfig = *req.TableConfig
	}
	if req.FormConfig != nil {
		model.FormConfig = *req.FormConfig
	}

	if err := s.modelRepo.Update(ctx, model); err != nil {
		return nil, err
	}

	return model, nil
}

func (s *ModelService) Delete(ctx context.Context, id primitive.ObjectID) error {
	model, err := s.modelRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Drop data collection first
	if err := s.modelRepo.DropDataCollection(ctx, model); err != nil {
		// Ignore error if collection doesn't exist
	}

	return s.modelRepo.Delete(ctx, id)
}

// Data methods
func (s *ModelService) CreateData(ctx context.Context, modelID primitive.ObjectID, data map[string]interface{}) (map[string]interface{}, error) {
	model, err := s.modelRepo.FindByID(ctx, modelID)
	if err != nil {
		return nil, err
	}

	// Validate and apply defaults
	validatedData := s.validateAndApplyDefaults(model, data)

	result, err := s.modelRepo.CreateData(ctx, model, validatedData)
	if err != nil {
		return nil, err
	}

	return bson.M{
		"id":         result.ID,
		"data":       result.Data,
		"created_at": result.CreatedAt,
		"updated_at": result.UpdatedAt,
	}, nil
}

func (s *ModelService) GetData(ctx context.Context, modelID primitive.ObjectID, query *domain.DataQuery) ([]bson.M, int64, error) {
	model, err := s.modelRepo.FindByID(ctx, modelID)
	if err != nil {
		return nil, 0, err
	}

	return s.modelRepo.FindData(ctx, model, query)
}

func (s *ModelService) GetDataByID(ctx context.Context, modelID, dataID primitive.ObjectID) (map[string]interface{}, error) {
	model, err := s.modelRepo.FindByID(ctx, modelID)
	if err != nil {
		return nil, err
	}

	return s.modelRepo.FindDataByID(ctx, model, dataID)
}

func (s *ModelService) UpdateData(ctx context.Context, modelID, dataID primitive.ObjectID, data map[string]interface{}) error {
	model, err := s.modelRepo.FindByID(ctx, modelID)
	if err != nil {
		return err
	}

	return s.modelRepo.UpdateData(ctx, model, dataID, data)
}

func (s *ModelService) DeleteData(ctx context.Context, modelID, dataID primitive.ObjectID) error {
	model, err := s.modelRepo.FindByID(ctx, modelID)
	if err != nil {
		return err
	}

	return s.modelRepo.DeleteData(ctx, model, dataID)
}

func (s *ModelService) validateAndApplyDefaults(model *domain.Model, data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for _, field := range model.Fields {
		if val, ok := data[field.Key]; ok {
			result[field.Key] = val
		} else if field.DefaultValue != nil {
			result[field.Key] = field.DefaultValue
		}
	}

	return result
}
