package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"m3m/internal/domain"
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

	var models []*domain.Model
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

	var results []bson.M
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

func (r *ModelRepository) DropDataCollection(ctx context.Context, model *domain.Model) error {
	collectionName := r.dataCollectionName(model.ProjectID, model.Slug)
	return r.db.Collection(collectionName).Drop(ctx)
}
