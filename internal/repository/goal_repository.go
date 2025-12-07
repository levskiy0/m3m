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
	ErrGoalNotFound   = errors.New("goal not found")
	ErrGoalSlugExists = errors.New("goal slug already exists")
)

type GoalRepository struct {
	goalsCollection *mongo.Collection
	statsCollection *mongo.Collection
}

func NewGoalRepository(db *MongoDB) *GoalRepository {
	goalsCollection := db.Collection("goals")
	statsCollection := db.Collection("goal_stats")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	goalsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "slug", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	statsCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "goal_id", Value: 1}, {Key: "date", Value: 1}}},
		{Keys: bson.D{{Key: "project_id", Value: 1}}},
	})

	return &GoalRepository{
		goalsCollection: goalsCollection,
		statsCollection: statsCollection,
	}
}

func (r *GoalRepository) Create(ctx context.Context, goal *domain.Goal) error {
	goal.ID = primitive.NewObjectID()
	goal.CreatedAt = time.Now()
	goal.UpdatedAt = time.Now()

	_, err := r.goalsCollection.InsertOne(ctx, goal)
	if mongo.IsDuplicateKeyError(err) {
		return ErrGoalSlugExists
	}
	return err
}

func (r *GoalRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Goal, error) {
	var goal domain.Goal
	err := r.goalsCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&goal)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrGoalNotFound
	}
	return &goal, err
}

func (r *GoalRepository) FindBySlug(ctx context.Context, slug string) (*domain.Goal, error) {
	var goal domain.Goal
	err := r.goalsCollection.FindOne(ctx, bson.M{"slug": slug}).Decode(&goal)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrGoalNotFound
	}
	return &goal, err
}

func (r *GoalRepository) FindGlobalGoals(ctx context.Context) ([]*domain.Goal, error) {
	cursor, err := r.goalsCollection.Find(ctx, bson.M{"project_ref": nil})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	goals := make([]*domain.Goal, 0)
	if err := cursor.All(ctx, &goals); err != nil {
		return nil, err
	}
	return goals, nil
}

func (r *GoalRepository) FindByProject(ctx context.Context, projectID primitive.ObjectID) ([]*domain.Goal, error) {
	cursor, err := r.goalsCollection.Find(ctx, bson.M{"project_ref": projectID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	goals := make([]*domain.Goal, 0)
	if err := cursor.All(ctx, &goals); err != nil {
		return nil, err
	}
	return goals, nil
}

func (r *GoalRepository) Update(ctx context.Context, goal *domain.Goal) error {
	goal.UpdatedAt = time.Now()
	_, err := r.goalsCollection.UpdateOne(
		ctx,
		bson.M{"_id": goal.ID},
		bson.M{"$set": goal},
	)
	return err
}

func (r *GoalRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.goalsCollection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrGoalNotFound
	}
	return nil
}

func (r *GoalRepository) IncrementStat(ctx context.Context, goalID, projectID primitive.ObjectID, date string, value int64) error {
	filter := bson.M{
		"goal_id":    goalID,
		"project_id": projectID,
		"date":       date,
	}
	update := bson.M{
		"$inc": bson.M{"value": value},
		"$set": bson.M{"updated_at": time.Now()},
	}
	opts := options.Update().SetUpsert(true)
	_, err := r.statsCollection.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *GoalRepository) GetStats(ctx context.Context, query *domain.GoalStatsQuery) ([]*domain.GoalStat, error) {
	filter := bson.M{}

	if len(query.GoalIDs) > 0 {
		goalOIDs := make([]primitive.ObjectID, 0, len(query.GoalIDs))
		for _, id := range query.GoalIDs {
			oid, err := primitive.ObjectIDFromHex(id)
			if err == nil {
				goalOIDs = append(goalOIDs, oid)
			}
		}
		filter["goal_id"] = bson.M{"$in": goalOIDs}
	}

	if query.ProjectID != "" {
		oid, err := primitive.ObjectIDFromHex(query.ProjectID)
		if err == nil {
			filter["project_id"] = oid
		}
	}

	if query.StartDate != "" && query.EndDate != "" {
		filter["date"] = bson.M{
			"$gte": query.StartDate,
			"$lte": query.EndDate,
		}
	} else if query.StartDate != "" {
		filter["date"] = bson.M{"$gte": query.StartDate}
	} else if query.EndDate != "" {
		filter["date"] = bson.M{"$lte": query.EndDate}
	}

	cursor, err := r.statsCollection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	stats := make([]*domain.GoalStat, 0)
	if err := cursor.All(ctx, &stats); err != nil {
		return nil, err
	}
	return stats, nil
}

// GetTotalValues returns the sum of all stats for each goal ID (across all dates)
func (r *GoalRepository) GetTotalValues(ctx context.Context, goalIDs []string) (map[string]int64, error) {
	goalOIDs := make([]primitive.ObjectID, 0, len(goalIDs))
	for _, id := range goalIDs {
		oid, err := primitive.ObjectIDFromHex(id)
		if err == nil {
			goalOIDs = append(goalOIDs, oid)
		}
	}

	if len(goalOIDs) == 0 {
		return make(map[string]int64), nil
	}

	pipeline := []bson.M{
		{"$match": bson.M{"goal_id": bson.M{"$in": goalOIDs}}},
		{"$group": bson.M{
			"_id":   "$goal_id",
			"total": bson.M{"$sum": "$value"},
		}},
	}

	cursor, err := r.statsCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := make(map[string]int64)
	for cursor.Next(ctx) {
		var doc struct {
			ID    primitive.ObjectID `bson:"_id"`
			Total int64              `bson:"total"`
		}
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		result[doc.ID.Hex()] = doc.Total
	}

	return result, nil
}
