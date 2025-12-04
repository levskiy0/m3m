package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"m3m/internal/domain"
)

var (
	ErrEnvVarNotFound = errors.New("environment variable not found")
	ErrEnvVarExists   = errors.New("environment variable already exists")
)

type EnvironmentRepository struct {
	collection *mongo.Collection
}

func NewEnvironmentRepository(db *MongoDB) *EnvironmentRepository {
	collection := db.Collection("environment")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
			{Key: "key", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})

	return &EnvironmentRepository{collection: collection}
}

func (r *EnvironmentRepository) Create(ctx context.Context, envVar *domain.EnvVar) error {
	envVar.ID = primitive.NewObjectID()
	envVar.CreatedAt = time.Now()
	envVar.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, envVar)
	if mongo.IsDuplicateKeyError(err) {
		return ErrEnvVarExists
	}
	return err
}

func (r *EnvironmentRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.EnvVar, error) {
	var envVar domain.EnvVar
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&envVar)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrEnvVarNotFound
	}
	return &envVar, err
}

func (r *EnvironmentRepository) FindByKey(ctx context.Context, projectID primitive.ObjectID, key string) (*domain.EnvVar, error) {
	var envVar domain.EnvVar
	err := r.collection.FindOne(ctx, bson.M{
		"project_id": projectID,
		"key":        key,
	}).Decode(&envVar)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrEnvVarNotFound
	}
	return &envVar, err
}

func (r *EnvironmentRepository) FindByProject(ctx context.Context, projectID primitive.ObjectID) ([]*domain.EnvVar, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"project_id": projectID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var envVars []*domain.EnvVar
	if err := cursor.All(ctx, &envVars); err != nil {
		return nil, err
	}
	return envVars, nil
}

func (r *EnvironmentRepository) Update(ctx context.Context, envVar *domain.EnvVar) error {
	envVar.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": envVar.ID},
		bson.M{"$set": envVar},
	)
	return err
}

func (r *EnvironmentRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrEnvVarNotFound
	}
	return nil
}

func (r *EnvironmentRepository) DeleteByKey(ctx context.Context, projectID primitive.ObjectID, key string) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{
		"project_id": projectID,
		"key":        key,
	})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrEnvVarNotFound
	}
	return nil
}

func (r *EnvironmentRepository) DeleteByProject(ctx context.Context, projectID primitive.ObjectID) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"project_id": projectID})
	return err
}
