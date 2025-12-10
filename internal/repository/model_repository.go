package repository

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/levskiy0/m3m/internal/domain"
)

var (
	ErrModelNotFound   = errors.New("model not found")
	ErrModelSlugExists = errors.New("model slug already exists")
	ErrDataNotFound    = errors.New("data not found")
)

type ModelRepository struct {
	db         *MongoDB
	collection *mongo.Collection
}

func NewModelRepository(db *MongoDB) *ModelRepository {
	collection := db.Collection("models")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
			{Key: "slug", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})

	return &ModelRepository{
		db:         db,
		collection: collection,
	}
}

func (r *ModelRepository) Create(ctx context.Context, model *domain.Model) error {
	model.ID = primitive.NewObjectID()
	model.CreatedAt = time.Now()
	model.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, model)
	if mongo.IsDuplicateKeyError(err) {
		return ErrModelSlugExists
	}
	return err
}

func (r *ModelRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Model, error) {
	var model domain.Model
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&model)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrModelNotFound
	}
	return &model, err
}

func (r *ModelRepository) FindBySlug(ctx context.Context, projectID primitive.ObjectID, slug string) (*domain.Model, error) {
	var model domain.Model
	err := r.collection.FindOne(ctx, bson.M{
		"project_id": projectID,
		"slug":       slug,
	}).Decode(&model)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrModelNotFound
	}
	return &model, err
}

func (r *ModelRepository) FindByProject(ctx context.Context, projectID primitive.ObjectID) ([]*domain.Model, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"project_id": projectID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	models := make([]*domain.Model, 0)
	if err := cursor.All(ctx, &models); err != nil {
		return nil, err
	}
	return models, nil
}

func (r *ModelRepository) Update(ctx context.Context, model *domain.Model) error {
	model.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": model.ID},
		bson.M{"$set": model},
	)
	return err
}

func (r *ModelRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrModelNotFound
	}
	return nil
}

// Data collection name for a model
func (r *ModelRepository) dataCollectionName(projectID primitive.ObjectID, modelSlug string) string {
	return fmt.Sprintf("data_%s_%s", projectID.Hex(), modelSlug)
}

// Data methods
func (r *ModelRepository) CreateData(ctx context.Context, model *domain.Model, data map[string]interface{}) (*domain.ModelData, error) {
	collectionName := r.dataCollectionName(model.ProjectID, model.Slug)
	collection := r.db.Collection(collectionName)

	modelData := &domain.ModelData{
		ID:        primitive.NewObjectID(),
		ModelID:   model.ID,
		Data:      data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	doc := bson.M{
		"_id":         modelData.ID,
		"_model_id":   modelData.ModelID,
		"_created_at": modelData.CreatedAt,
		"_updated_at": modelData.UpdatedAt,
	}
	for k, v := range data {
		doc[k] = v
	}

	_, err := collection.InsertOne(ctx, doc)
	return modelData, err
}

func (r *ModelRepository) FindDataByID(ctx context.Context, model *domain.Model, id primitive.ObjectID) (map[string]interface{}, error) {
	collectionName := r.dataCollectionName(model.ProjectID, model.Slug)
	collection := r.db.Collection(collectionName)

	var result bson.M
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrDataNotFound
	}
	return result, err
}

func (r *ModelRepository) FindData(ctx context.Context, model *domain.Model, query *domain.DataQuery) ([]bson.M, int64, error) {
	collectionName := r.dataCollectionName(model.ProjectID, model.Slug)
	collection := r.db.Collection(collectionName)

	filter := bson.M{}
	for k, v := range query.Filters {
		filter[k] = v
	}

	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find()
	if query.Limit > 0 {
		opts.SetLimit(int64(query.Limit))
	} else {
		opts.SetLimit(50)
	}
	if query.Page > 0 {
		opts.SetSkip(int64((query.Page - 1) * query.Limit))
	}
	if query.Sort != "" {
		order := 1
		if query.Order == "desc" {
			order = -1
		}
		opts.SetSort(bson.D{{Key: query.Sort, Value: order}})
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	results := make([]bson.M, 0)
	if err := cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}
	return results, total, nil
}

func (r *ModelRepository) UpdateData(ctx context.Context, model *domain.Model, id primitive.ObjectID, data map[string]interface{}) error {
	collectionName := r.dataCollectionName(model.ProjectID, model.Slug)
	collection := r.db.Collection(collectionName)

	update := bson.M{
		"$set": bson.M{"_updated_at": time.Now()},
	}
	for k, v := range data {
		update["$set"].(bson.M)[k] = v
	}

	result, err := collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrDataNotFound
	}
	return nil
}

func (r *ModelRepository) DeleteData(ctx context.Context, model *domain.Model, id primitive.ObjectID) error {
	collectionName := r.dataCollectionName(model.ProjectID, model.Slug)
	collection := r.db.Collection(collectionName)

	result, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrDataNotFound
	}
	return nil
}

// DeleteManyData deletes multiple documents by their IDs
func (r *ModelRepository) DeleteManyData(ctx context.Context, model *domain.Model, ids []primitive.ObjectID) (int64, error) {
	collectionName := r.dataCollectionName(model.ProjectID, model.Slug)
	collection := r.db.Collection(collectionName)

	result, err := collection.DeleteMany(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}

// UpsertData inserts or updates a document based on filter
func (r *ModelRepository) UpsertData(ctx context.Context, model *domain.Model, filter bson.M, data map[string]interface{}) (*domain.ModelData, bool, error) {
	collectionName := r.dataCollectionName(model.ProjectID, model.Slug)
	collection := r.db.Collection(collectionName)

	now := time.Now()
	newID := primitive.NewObjectID()

	update := bson.M{
		"$set": bson.M{"_updated_at": now},
		"$setOnInsert": bson.M{
			"_id":         newID,
			"_model_id":   model.ID,
			"_created_at": now,
		},
	}

	for k, v := range data {
		update["$set"].(bson.M)[k] = v
	}

	opts := options.Update().SetUpsert(true)
	result, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return nil, false, err
	}

	isNew := result.UpsertedCount > 0

	// Find the document to return
	var doc bson.M
	if isNew {
		err = collection.FindOne(ctx, bson.M{"_id": newID}).Decode(&doc)
	} else {
		err = collection.FindOne(ctx, filter).Decode(&doc)
	}
	if err != nil {
		return nil, isNew, err
	}

	modelData := &domain.ModelData{
		ID:        doc["_id"].(primitive.ObjectID),
		ModelID:   model.ID,
		Data:      r.extractUserData(doc),
		CreatedAt: doc["_created_at"].(primitive.DateTime).Time(),
		UpdatedAt: doc["_updated_at"].(primitive.DateTime).Time(),
	}

	return modelData, isNew, nil
}

// FindOneAndUpdateData finds and updates a document atomically
func (r *ModelRepository) FindOneAndUpdateData(ctx context.Context, model *domain.Model, filter bson.M, updateOps map[string]interface{}, returnNew bool) (map[string]interface{}, error) {
	collectionName := r.dataCollectionName(model.ProjectID, model.Slug)
	collection := r.db.Collection(collectionName)

	update := r.buildUpdateDocument(updateOps)
	update["$set"].(bson.M)["_updated_at"] = time.Now()

	opts := options.FindOneAndUpdate()
	if returnNew {
		opts.SetReturnDocument(options.After)
	} else {
		opts.SetReturnDocument(options.Before)
	}

	var result bson.M
	err := collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrDataNotFound
		}
		return nil, err
	}

	return result, nil
}

// buildUpdateDocument converts update operations to MongoDB update document
func (r *ModelRepository) buildUpdateDocument(updateOps map[string]interface{}) bson.M {
	update := bson.M{
		"$set": bson.M{},
	}

	for op, value := range updateOps {
		switch op {
		case "$set":
			if setMap, ok := value.(map[string]interface{}); ok {
				for k, v := range setMap {
					update["$set"].(bson.M)[k] = v
				}
			}
		case "$inc":
			update["$inc"] = value
		case "$unset":
			update["$unset"] = value
		case "$push":
			update["$push"] = value
		case "$pull":
			update["$pull"] = value
		case "$addToSet":
			update["$addToSet"] = value
		default:
			// Treat as $set field
			update["$set"].(bson.M)[op] = value
		}
	}

	return update
}

// extractUserData extracts user data fields from a document (excludes system fields)
func (r *ModelRepository) extractUserData(doc bson.M) map[string]interface{} {
	data := make(map[string]interface{})
	for k, v := range doc {
		if k != "_id" && k != "_model_id" && k != "_created_at" && k != "_updated_at" {
			data[k] = v
		}
	}
	return data
}

func (r *ModelRepository) DropDataCollection(ctx context.Context, model *domain.Model) error {
	collectionName := r.dataCollectionName(model.ProjectID, model.Slug)
	return r.db.Collection(collectionName).Drop(ctx)
}

// ExistsDataByFilter checks if a document exists matching the filter
func (r *ModelRepository) ExistsDataByFilter(ctx context.Context, model *domain.Model, filter bson.M) (bool, error) {
	collectionName := r.dataCollectionName(model.ProjectID, model.Slug)
	collection := r.db.Collection(collectionName)

	count, err := collection.CountDocuments(ctx, filter, options.Count().SetLimit(1))
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// FindDataAdvanced finds data with advanced filtering support
func (r *ModelRepository) FindDataAdvanced(ctx context.Context, model *domain.Model, query *domain.AdvancedDataQuery) ([]bson.M, int64, error) {
	collectionName := r.dataCollectionName(model.ProjectID, model.Slug)
	collection := r.db.Collection(collectionName)

	filter := r.buildAdvancedFilter(query, model)

	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	opts := options.Find()
	limit := query.Limit
	if limit <= 0 {
		limit = 50
	}
	opts.SetLimit(int64(limit))

	if query.Page > 0 {
		opts.SetSkip(int64((query.Page - 1) * limit))
	}
	if query.Sort != "" {
		order := 1
		if query.Order == "desc" {
			order = -1
		}
		opts.SetSort(bson.D{{Key: query.Sort, Value: order}})
	}

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	results := make([]bson.M, 0)
	if err := cursor.All(ctx, &results); err != nil {
		return nil, 0, err
	}
	return results, total, nil
}

// buildAdvancedFilter builds MongoDB filter from AdvancedDataQuery
func (r *ModelRepository) buildAdvancedFilter(query *domain.AdvancedDataQuery, model *domain.Model) bson.M {
	filter := bson.M{}
	conditions := []bson.M{}

	// Build field type map for type-aware filtering
	fieldTypes := make(map[string]domain.FieldType)
	for _, f := range model.Fields {
		fieldTypes[f.Key] = f.Type
	}

	// Process filter conditions
	for _, cond := range query.Filters {
		mongoFilter := r.buildFilterCondition(cond, fieldTypes[cond.Field])
		if mongoFilter != nil {
			conditions = append(conditions, mongoFilter)
		}
	}

	// Add search filter if provided
	if query.Search != "" && len(query.SearchIn) > 0 {
		searchConditions := make([]bson.M, 0, len(query.SearchIn))
		// Escape regex special characters to prevent ReDoS
		escapedSearch := regexp.QuoteMeta(query.Search)
		for _, field := range query.SearchIn {
			searchConditions = append(searchConditions, bson.M{
				field: bson.M{"$regex": escapedSearch, "$options": "i"},
			})
		}
		conditions = append(conditions, bson.M{"$or": searchConditions})
	}

	if len(conditions) > 0 {
		filter["$and"] = conditions
	}

	return filter
}

// buildFilterCondition converts a FilterCondition to MongoDB filter
func (r *ModelRepository) buildFilterCondition(cond domain.FilterCondition, fieldType domain.FieldType) bson.M {
	field := cond.Field
	value := r.coerceFilterValue(cond.Value, fieldType)

	switch cond.Operator {
	case domain.FilterOpEquals, "":
		return bson.M{field: value}
	case domain.FilterOpNotEquals:
		return bson.M{field: bson.M{"$ne": value}}
	case domain.FilterOpGreater:
		return bson.M{field: bson.M{"$gt": value}}
	case domain.FilterOpGreaterEq:
		return bson.M{field: bson.M{"$gte": value}}
	case domain.FilterOpLess:
		return bson.M{field: bson.M{"$lt": value}}
	case domain.FilterOpLessEq:
		return bson.M{field: bson.M{"$lte": value}}
	case domain.FilterOpContains:
		if str, ok := value.(string); ok {
			// Escape regex special characters to prevent ReDoS
			return bson.M{field: bson.M{"$regex": regexp.QuoteMeta(str), "$options": "i"}}
		}
		return nil
	case domain.FilterOpStartsWith:
		if str, ok := value.(string); ok {
			// Escape regex special characters to prevent ReDoS
			return bson.M{field: bson.M{"$regex": "^" + regexp.QuoteMeta(str), "$options": "i"}}
		}
		return nil
	case domain.FilterOpEndsWith:
		if str, ok := value.(string); ok {
			// Escape regex special characters to prevent ReDoS
			return bson.M{field: bson.M{"$regex": regexp.QuoteMeta(str) + "$", "$options": "i"}}
		}
		return nil
	case domain.FilterOpIn:
		if arr, ok := value.([]interface{}); ok {
			return bson.M{field: bson.M{"$in": arr}}
		}
		return nil
	case domain.FilterOpNotIn:
		if arr, ok := value.([]interface{}); ok {
			return bson.M{field: bson.M{"$nin": arr}}
		}
		return nil
	case domain.FilterOpIsNull:
		return bson.M{"$or": []bson.M{
			{field: nil},
			{field: bson.M{"$exists": false}},
		}}
	case domain.FilterOpIsNotNull:
		return bson.M{field: bson.M{"$exists": true, "$ne": nil}}
	default:
		return bson.M{field: value}
	}
}

// coerceFilterValue attempts to convert filter value to appropriate type
func (r *ModelRepository) coerceFilterValue(value interface{}, fieldType domain.FieldType) interface{} {
	if value == nil {
		return nil
	}

	switch fieldType {
	case domain.FieldTypeNumber:
		if str, ok := value.(string); ok {
			if i, err := parseInt(str); err == nil {
				return i
			}
		}
		if f, ok := value.(float64); ok {
			return int64(f)
		}
	case domain.FieldTypeFloat:
		if str, ok := value.(string); ok {
			if f, err := parseFloat(str); err == nil {
				return f
			}
		}
	case domain.FieldTypeBool:
		if str, ok := value.(string); ok {
			return str == "true" || str == "1"
		}
	case domain.FieldTypeRef:
		if str, ok := value.(string); ok {
			if oid, err := primitive.ObjectIDFromHex(str); err == nil {
				return oid
			}
		}
	case domain.FieldTypeDate, domain.FieldTypeDateTime:
		if str, ok := value.(string); ok {
			if t, err := time.Parse(time.RFC3339, str); err == nil {
				return t
			}
			// Try without timezone
			if t, err := time.Parse("2006-01-02T15:04:05", str); err == nil {
				return t
			}
			// Try date only
			if t, err := time.Parse("2006-01-02", str); err == nil {
				return t
			}
		}
	}

	return value
}

func parseInt(s string) (int64, error) {
	var i int64
	_, err := fmt.Sscanf(s, "%d", &i)
	return i, err
}

func parseFloat(s string) (float64, error) {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	return f, err
}

// GetProjectDataSize calculates total size of all data collections for a project in bytes
func (r *ModelRepository) GetProjectDataSize(ctx context.Context, projectID primitive.ObjectID) (int64, error) {
	// Find all models for this project
	models, err := r.FindByProject(ctx, projectID)
	if err != nil {
		return 0, err
	}

	var totalSize int64
	for _, model := range models {
		collectionName := r.dataCollectionName(model.ProjectID, model.Slug)

		// Run collStats command to get collection size
		result := r.db.Database.RunCommand(ctx, bson.D{
			{Key: "collStats", Value: collectionName},
		})

		var stats bson.M
		if err := result.Decode(&stats); err != nil {
			// Collection might not exist yet, skip
			continue
		}

		// Get storageSize (actual disk size including indexes)
		if size, ok := stats["storageSize"].(int64); ok {
			totalSize += size
		} else if size, ok := stats["storageSize"].(int32); ok {
			totalSize += int64(size)
		}
	}

	return totalSize, nil
}
