package modules

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"m3m/internal/domain"
	"m3m/internal/service"
)

type DatabaseModule struct {
	modelService *service.ModelService
	projectID    primitive.ObjectID
}

func NewDatabaseModule(modelService *service.ModelService, projectID primitive.ObjectID) *DatabaseModule {
	return &DatabaseModule{
		modelService: modelService,
		projectID:    projectID,
	}
}

type CollectionWrapper struct {
	modelService *service.ModelService
	projectID    primitive.ObjectID
	modelSlug    string
	modelID      primitive.ObjectID
}

func (d *DatabaseModule) Collection(name string) *CollectionWrapper {
	ctx := context.Background()
	model, err := d.modelService.GetBySlug(ctx, d.projectID, name)
	if err != nil {
		return &CollectionWrapper{
			modelService: d.modelService,
			projectID:    d.projectID,
			modelSlug:    name,
		}
	}

	return &CollectionWrapper{
		modelService: d.modelService,
		projectID:    d.projectID,
		modelSlug:    name,
		modelID:      model.ID,
	}
}

func (c *CollectionWrapper) Find(filter map[string]interface{}) []map[string]interface{} {
	if c.modelID.IsZero() {
		return []map[string]interface{}{}
	}

	ctx := context.Background()
	query := &domain.DataQuery{
		Page:    1,
		Limit:   100,
		Filters: make(map[string]string),
	}

	for k, v := range filter {
		if str, ok := v.(string); ok {
			query.Filters[k] = str
		}
	}

	data, _, err := c.modelService.GetData(ctx, c.modelID, query)
	if err != nil {
		return []map[string]interface{}{}
	}

	result := make([]map[string]interface{}, len(data))
	for i, d := range data {
		result[i] = map[string]interface{}(d)
	}
	return result
}

func (c *CollectionWrapper) FindOne(filter map[string]interface{}) map[string]interface{} {
	results := c.Find(filter)
	if len(results) > 0 {
		return results[0]
	}
	return nil
}

func (c *CollectionWrapper) Insert(data map[string]interface{}) map[string]interface{} {
	if c.modelID.IsZero() {
		return nil
	}

	ctx := context.Background()
	result, err := c.modelService.CreateData(ctx, c.modelID, data)
	if err != nil {
		return nil
	}

	return result
}

func (c *CollectionWrapper) Update(id string, data map[string]interface{}) bool {
	if c.modelID.IsZero() {
		return false
	}

	dataID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return false
	}

	ctx := context.Background()
	err = c.modelService.UpdateData(ctx, c.modelID, dataID, data)
	return err == nil
}

func (c *CollectionWrapper) Delete(id string) bool {
	if c.modelID.IsZero() {
		return false
	}

	dataID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return false
	}

	ctx := context.Background()
	err = c.modelService.DeleteData(ctx, c.modelID, dataID)
	return err == nil
}

func (c *CollectionWrapper) Count(filter map[string]interface{}) int64 {
	if c.modelID.IsZero() {
		return 0
	}

	ctx := context.Background()
	query := &domain.DataQuery{
		Page:    1,
		Limit:   1,
		Filters: make(map[string]string),
	}

	for k, v := range filter {
		if str, ok := v.(string); ok {
			query.Filters[k] = str
		}
	}

	_, total, err := c.modelService.GetData(ctx, c.modelID, query)
	if err != nil {
		return 0
	}
	return total
}
