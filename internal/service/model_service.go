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
	// Validate model schema
	schemaValidator := NewModelSchemaValidator()
	if err := schemaValidator.ValidateModel(req); err != nil {
		return nil, err
	}

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

	// Validate update request
	schemaValidator := NewModelSchemaValidator()
	if err := schemaValidator.ValidateModelUpdate(req, model); err != nil {
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

	// Apply defaults for missing fields
	dataWithDefaults := s.applyDefaults(model, data)

	// Validate data against schema
	validator := NewDataValidator(model)
	validatedData, err := validator.ValidateAndCoerce(dataWithDefaults)
	if err != nil {
		return nil, err
	}

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

// GetDataAdvanced retrieves data with advanced filtering
func (s *ModelService) GetDataAdvanced(ctx context.Context, modelID primitive.ObjectID, query *domain.AdvancedDataQuery) ([]bson.M, int64, error) {
	model, err := s.modelRepo.FindByID(ctx, modelID)
	if err != nil {
		return nil, 0, err
	}

	return s.modelRepo.FindDataAdvanced(ctx, model, query)
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

	// Validate partial data (only provided fields)
	validator := NewDataValidator(model)
	if err := validator.ValidatePartial(data); err != nil {
		return err
	}

	// Coerce values to proper types
	coercedData := make(map[string]interface{})
	fieldMap := make(map[string]domain.ModelField)
	for _, f := range model.Fields {
		fieldMap[f.Key] = f
	}
	for key, value := range data {
		if field, exists := fieldMap[key]; exists {
			coercedData[key] = validator.CoerceValue(field, value)
		}
	}

	return s.modelRepo.UpdateData(ctx, model, dataID, coercedData)
}

func (s *ModelService) DeleteData(ctx context.Context, modelID, dataID primitive.ObjectID) error {
	model, err := s.modelRepo.FindByID(ctx, modelID)
	if err != nil {
		return err
	}

	return s.modelRepo.DeleteData(ctx, model, dataID)
}

func (s *ModelService) applyDefaults(model *domain.Model, data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy existing data
	for k, v := range data {
		result[k] = v
	}

	// Apply defaults for missing fields
	for _, field := range model.Fields {
		if _, exists := result[field.Key]; !exists && field.DefaultValue != nil {
			result[field.Key] = field.DefaultValue
		}
	}

	return result
}
