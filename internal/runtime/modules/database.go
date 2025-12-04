package modules

import (
	"context"

	"github.com/dop251/goja"
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

// Name returns the module name for JavaScript
func (d *DatabaseModule) Name() string {
	return "database"
}

// Register registers the module into the JavaScript VM
func (d *DatabaseModule) Register(vm interface{}) {
	vm.(*goja.Runtime).Set(d.Name(), map[string]interface{}{
		"collection": d.Collection,
	})
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
	return c.FindWithOptions(filter, nil)
}

// FindWithOptions supports advanced query options
func (c *CollectionWrapper) FindWithOptions(filter map[string]interface{}, options map[string]interface{}) []map[string]interface{} {
	if c.modelID.IsZero() {
		return []map[string]interface{}{}
	}

	ctx := context.Background()
	query := &domain.AdvancedDataQuery{
		Page:    1,
		Limit:   100,
		Filters: []domain.FilterCondition{},
	}

	// Parse options
	if options != nil {
		if page, ok := options["page"].(float64); ok {
			query.Page = int(page)
		}
		if limit, ok := options["limit"].(float64); ok {
			query.Limit = int(limit)
		}
		if sort, ok := options["sort"].(string); ok {
			query.Sort = sort
		}
		if order, ok := options["order"].(string); ok {
			query.Order = order
		}
	}

	// Convert filter to FilterConditions
	for k, v := range filter {
		// Check if value is an operator object like {$gt: 10}
		if opMap, ok := v.(map[string]interface{}); ok {
			for op, val := range opMap {
				var operator domain.FilterOperator
				switch op {
				case "$eq":
					operator = domain.FilterOpEquals
				case "$ne":
					operator = domain.FilterOpNotEquals
				case "$gt":
					operator = domain.FilterOpGreater
				case "$gte":
					operator = domain.FilterOpGreaterEq
				case "$lt":
					operator = domain.FilterOpLess
				case "$lte":
					operator = domain.FilterOpLessEq
				case "$contains":
					operator = domain.FilterOpContains
				case "$startsWith":
					operator = domain.FilterOpStartsWith
				case "$endsWith":
					operator = domain.FilterOpEndsWith
				case "$in":
					operator = domain.FilterOpIn
				case "$nin":
					operator = domain.FilterOpNotIn
				default:
					continue
				}
				query.Filters = append(query.Filters, domain.FilterCondition{
					Field:    k,
					Operator: operator,
					Value:    val,
				})
			}
		} else {
			// Simple equality filter
			query.Filters = append(query.Filters, domain.FilterCondition{
				Field:    k,
				Operator: domain.FilterOpEquals,
				Value:    v,
			})
		}
	}

	data, _, err := c.modelService.GetDataAdvanced(ctx, c.modelID, query)
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
	query := &domain.AdvancedDataQuery{
		Page:    1,
		Limit:   1,
		Filters: []domain.FilterCondition{},
	}

	// Convert filter to FilterConditions (same logic as Find)
	for k, v := range filter {
		if opMap, ok := v.(map[string]interface{}); ok {
			for op, val := range opMap {
				var operator domain.FilterOperator
				switch op {
				case "$eq":
					operator = domain.FilterOpEquals
				case "$ne":
					operator = domain.FilterOpNotEquals
				case "$gt":
					operator = domain.FilterOpGreater
				case "$gte":
					operator = domain.FilterOpGreaterEq
				case "$lt":
					operator = domain.FilterOpLess
				case "$lte":
					operator = domain.FilterOpLessEq
				case "$contains":
					operator = domain.FilterOpContains
				default:
					continue
				}
				query.Filters = append(query.Filters, domain.FilterCondition{
					Field:    k,
					Operator: operator,
					Value:    val,
				})
			}
		} else {
			query.Filters = append(query.Filters, domain.FilterCondition{
				Field:    k,
				Operator: domain.FilterOpEquals,
				Value:    v,
			})
		}
	}

	_, total, err := c.modelService.GetDataAdvanced(ctx, c.modelID, query)
	if err != nil {
		return 0
	}
	return total
}

// GetSchema implements JSSchemaProvider
func (d *DatabaseModule) GetSchema() JSModuleSchema {
	return JSModuleSchema{
		Name:        "database",
		Description: "Database operations for model-based data storage. Supports MongoDB-style filter operators: $eq, $ne, $gt, $gte, $lt, $lte, $contains, $startsWith, $endsWith, $in, $nin",
		Types: []JSTypeSchema{
			{
				Name:        "QueryOptions",
				Description: "Options for find queries",
				Fields: []JSParamSchema{
					{Name: "page", Type: "number", Description: "Page number (default: 1)"},
					{Name: "limit", Type: "number", Description: "Results per page (default: 100)"},
					{Name: "sort", Type: "string", Description: "Field to sort by"},
					{Name: "order", Type: "'asc' | 'desc'", Description: "Sort order"},
				},
			},
			{
				Name:        "Collection",
				Description: "A database collection for a model",
				Fields: []JSParamSchema{
					{Name: "find", Type: "(filter?: object) => object[]", Description: "Find documents matching filter. Supports operators: {field: {$gt: value}}"},
					{Name: "findWithOptions", Type: "(filter?: object, options?: QueryOptions) => object[]", Description: "Find with pagination and sorting"},
					{Name: "findOne", Type: "(filter?: object) => object | null", Description: "Find first document matching filter"},
					{Name: "insert", Type: "(data: object) => object | null", Description: "Insert a new document"},
					{Name: "update", Type: "(id: string, data: object) => boolean", Description: "Update a document by ID"},
					{Name: "delete", Type: "(id: string) => boolean", Description: "Delete a document by ID"},
					{Name: "count", Type: "(filter?: object) => number", Description: "Count documents matching filter"},
				},
			},
		},
		Methods: []JSMethodSchema{
			{
				Name:        "collection",
				Description: "Get a collection wrapper for a model",
				Params:      []JSParamSchema{{Name: "name", Type: "string", Description: "Model slug name"}},
				Returns:     &JSParamSchema{Type: "Collection"},
			},
		},
	}
}

// GetDatabaseSchema returns the database schema (static version)
func GetDatabaseSchema() JSModuleSchema {
	return (&DatabaseModule{}).GetSchema()
}
