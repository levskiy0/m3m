package service

import (
	"context"
	"time"

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
		searchable := make([]string, 0)
		for i, f := range req.Fields {
			columns[i] = f.Key
			// Make string/text fields searchable by default
			if f.Type == domain.FieldTypeString || f.Type == domain.FieldTypeText {
				searchable = append(searchable, f.Key)
			}
		}
		model.TableConfig = domain.TableConfig{
			Columns:     columns,
			Filters:     []string{},
			SortColumns: columns,
			Searchable:  searchable,
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

// DeleteManyData deletes multiple data records by their IDs
func (s *ModelService) DeleteManyData(ctx context.Context, modelID primitive.ObjectID, dataIDs []primitive.ObjectID) (int64, error) {
	model, err := s.modelRepo.FindByID(ctx, modelID)
	if err != nil {
		return 0, err
	}

	return s.modelRepo.DeleteManyData(ctx, model, dataIDs)
}

// UpsertData inserts or updates a data record based on filter
func (s *ModelService) UpsertData(ctx context.Context, modelID primitive.ObjectID, filter bson.M, data map[string]interface{}) (map[string]interface{}, bool, error) {
	model, err := s.modelRepo.FindByID(ctx, modelID)
	if err != nil {
		return nil, false, err
	}

	// Apply defaults for missing fields (only on insert)
	dataWithDefaults := s.applyDefaults(model, data)

	// Validate data against schema
	validator := NewDataValidator(model)
	validatedData, err := validator.ValidateAndCoerce(dataWithDefaults)
	if err != nil {
		return nil, false, err
	}

	result, isNew, err := s.modelRepo.UpsertData(ctx, model, filter, validatedData)
	if err != nil {
		return nil, false, err
	}

	return bson.M{
		"id":         result.ID,
		"data":       result.Data,
		"created_at": result.CreatedAt,
		"updated_at": result.UpdatedAt,
	}, isNew, nil
}

// FindOneAndUpdateData finds and updates a document atomically
func (s *ModelService) FindOneAndUpdateData(ctx context.Context, modelID primitive.ObjectID, filter bson.M, updateOps map[string]interface{}, returnNew bool) (map[string]interface{}, error) {
	model, err := s.modelRepo.FindByID(ctx, modelID)
	if err != nil {
		return nil, err
	}

	return s.modelRepo.FindOneAndUpdateData(ctx, model, filter, updateOps, returnNew)
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
			result[field.Key] = s.resolveDefaultValue(field)
		}
	}

	return result
}

// resolveDefaultValue handles special default values like $now
func (s *ModelService) resolveDefaultValue(field domain.ModelField) interface{} {
	if strVal, ok := field.DefaultValue.(string); ok {
		if strVal == "$now" {
			now := time.Now()
			switch field.Type {
			case domain.FieldTypeDate:
				return now.Format("2006-01-02")
			case domain.FieldTypeDateTime:
				return now.Format(time.RFC3339)
			}
		}
	}
	return field.DefaultValue
}

// GetProjectDataSize returns total size of all data collections for a project in bytes
func (s *ModelService) GetProjectDataSize(ctx context.Context, projectID primitive.ObjectID) (int64, error) {
	return s.modelRepo.GetProjectDataSize(ctx, projectID)
}
