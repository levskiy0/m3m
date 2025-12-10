package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/levskiy0/m3m/internal/domain"
)

var (
	ErrActionNotFound = errors.New("action not found")
	ErrActionExists   = errors.New("action with this slug already exists")
)

type ActionRepository struct {
	collection *mongo.Collection
}

func NewActionRepository(db *MongoDB) *ActionRepository {
	collection := db.Collection("actions")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create unique index on project_id + slug
	collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
			{Key: "slug", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})

	return &ActionRepository{collection: collection}
}

func (r *ActionRepository) Create(ctx context.Context, action *domain.Action) error {
	action.ID = primitive.NewObjectID()
	action.CreatedAt = time.Now()
	action.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, action)
	if mongo.IsDuplicateKeyError(err) {
		return ErrActionExists
	}
	return err
}

func (r *ActionRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Action, error) {
	var action domain.Action
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&action)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrActionNotFound
	}
	return &action, err
}

func (r *ActionRepository) FindBySlug(ctx context.Context, projectID primitive.ObjectID, slug string) (*domain.Action, error) {
	var action domain.Action
	err := r.collection.FindOne(ctx, bson.M{
		"project_id": projectID,
		"slug":       slug,
	}).Decode(&action)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrActionNotFound
	}
	return &action, err
}

func (r *ActionRepository) FindByProjectID(ctx context.Context, projectID primitive.ObjectID) ([]*domain.Action, error) {
	opts := options.Find().SetSort(bson.D{{Key: "order", Value: 1}})
	cursor, err := r.collection.Find(ctx, bson.M{"project_id": projectID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	actions := make([]*domain.Action, 0)
	if err := cursor.All(ctx, &actions); err != nil {
		return nil, err
	}
	return actions, nil
}

func (r *ActionRepository) Update(ctx context.Context, action *domain.Action) error {
	action.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": action.ID},
		bson.M{"$set": action},
	)
	return err
}

func (r *ActionRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrActionNotFound
	}
	return nil
}

func (r *ActionRepository) DeleteByProject(ctx context.Context, projectID primitive.ObjectID) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"project_id": projectID})
	return err
}

// GetMaxOrder returns the maximum order value for actions in a project
func (r *ActionRepository) GetMaxOrder(ctx context.Context, projectID primitive.ObjectID) (int, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "order", Value: -1}})
	var action domain.Action
	err := r.collection.FindOne(ctx, bson.M{"project_id": projectID}, opts).Decode(&action)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return action.Order, nil
}
