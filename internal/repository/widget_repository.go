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
	ErrWidgetNotFound = errors.New("widget not found")
)

type WidgetRepository struct {
	collection *mongo.Collection
}

func NewWidgetRepository(db *MongoDB) *WidgetRepository {
	collection := db.Collection("widgets")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "project_id", Value: 1}, {Key: "order", Value: 1}}},
		{Keys: bson.D{{Key: "project_id", Value: 1}, {Key: "goal_id", Value: 1}}},
	})

	return &WidgetRepository{collection: collection}
}

func (r *WidgetRepository) Create(ctx context.Context, widget *domain.Widget) error {
	widget.ID = primitive.NewObjectID()
	widget.CreatedAt = time.Now()
	widget.UpdatedAt = time.Now()

	// Get max order for project
	var lastWidget domain.Widget
	opts := options.FindOne().SetSort(bson.D{{Key: "order", Value: -1}})
	err := r.collection.FindOne(ctx, bson.M{"project_id": widget.ProjectID}, opts).Decode(&lastWidget)
	if err == nil {
		widget.Order = lastWidget.Order + 1
	} else {
		widget.Order = 0
	}

	_, err = r.collection.InsertOne(ctx, widget)
	return err
}

func (r *WidgetRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Widget, error) {
	var widget domain.Widget
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&widget)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrWidgetNotFound
	}
	return &widget, err
}

func (r *WidgetRepository) FindByProject(ctx context.Context, projectID primitive.ObjectID) ([]*domain.Widget, error) {
	opts := options.Find().SetSort(bson.D{{Key: "order", Value: 1}})
	cursor, err := r.collection.Find(ctx, bson.M{"project_id": projectID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	widgets := make([]*domain.Widget, 0)
	if err := cursor.All(ctx, &widgets); err != nil {
		return nil, err
	}
	return widgets, nil
}

func (r *WidgetRepository) Update(ctx context.Context, widget *domain.Widget) error {
	widget.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": widget.ID},
		bson.M{"$set": widget},
	)
	return err
}

func (r *WidgetRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrWidgetNotFound
	}
	return nil
}

func (r *WidgetRepository) DeleteByProject(ctx context.Context, projectID primitive.ObjectID) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"project_id": projectID})
	return err
}

func (r *WidgetRepository) Reorder(ctx context.Context, projectID primitive.ObjectID, widgetIDs []primitive.ObjectID) error {
	for i, id := range widgetIDs {
		_, err := r.collection.UpdateOne(
			ctx,
			bson.M{"_id": id, "project_id": projectID},
			bson.M{"$set": bson.M{"order": i, "updated_at": time.Now()}},
		)
		if err != nil {
			return err
		}
	}
	return nil
}
