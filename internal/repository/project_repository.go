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
	ErrProjectNotFound   = errors.New("project not found")
	ErrProjectSlugExists = errors.New("project slug already exists")
)

type ProjectRepository struct {
	collection *mongo.Collection
}

func NewProjectRepository(db *MongoDB) *ProjectRepository {
	collection := db.Collection("projects")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "slug", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	return &ProjectRepository{collection: collection}
}

func (r *ProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	project.ID = primitive.NewObjectID()
	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()
	project.Status = domain.ProjectStatusStopped

	_, err := r.collection.InsertOne(ctx, project)
	if mongo.IsDuplicateKeyError(err) {
		return ErrProjectSlugExists
	}
	return err
}

func (r *ProjectRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.Project, error) {
	var project domain.Project
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&project)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrProjectNotFound
	}
	return &project, err
}

func (r *ProjectRepository) FindBySlug(ctx context.Context, slug string) (*domain.Project, error) {
	var project domain.Project
	err := r.collection.FindOne(ctx, bson.M{"slug": slug}).Decode(&project)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrProjectNotFound
	}
	return &project, err
}

func (r *ProjectRepository) FindAll(ctx context.Context) ([]*domain.Project, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	projects := make([]*domain.Project, 0)
	if err := cursor.All(ctx, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *ProjectRepository) FindByOwner(ctx context.Context, ownerID primitive.ObjectID) ([]*domain.Project, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"owner_id": ownerID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	projects := make([]*domain.Project, 0)
	if err := cursor.All(ctx, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *ProjectRepository) FindByMember(ctx context.Context, userID primitive.ObjectID) ([]*domain.Project, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"$or": []bson.M{
			{"owner_id": userID},
			{"members": userID},
		},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	projects := make([]*domain.Project, 0)
	if err := cursor.All(ctx, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *ProjectRepository) Update(ctx context.Context, project *domain.Project) error {
	project.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": project.ID},
		bson.M{"$set": project},
	)
	return err
}

func (r *ProjectRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrProjectNotFound
	}
	return nil
}

func (r *ProjectRepository) AddMember(ctx context.Context, projectID, userID primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": projectID},
		bson.M{
			"$addToSet": bson.M{"members": userID},
			"$set":      bson.M{"updated_at": time.Now()},
		},
	)
	return err
}

func (r *ProjectRepository) RemoveMember(ctx context.Context, projectID, userID primitive.ObjectID) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": projectID},
		bson.M{
			"$pull": bson.M{"members": userID},
			"$set":  bson.M{"updated_at": time.Now()},
		},
	)
	return err
}

func (r *ProjectRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status domain.ProjectStatus) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"status": status, "updated_at": time.Now()}},
	)
	return err
}

func (r *ProjectRepository) SetActiveRelease(ctx context.Context, id primitive.ObjectID, version string) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"active_release": version, "updated_at": time.Now()}},
	)
	return err
}

func (r *ProjectRepository) FindByStatus(ctx context.Context, status domain.ProjectStatus) ([]*domain.Project, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"status": status})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	projects := make([]*domain.Project, 0)
	if err := cursor.All(ctx, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *ProjectRepository) SetAutoStart(ctx context.Context, id primitive.ObjectID, autoStart bool) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"auto_start": autoStart, "updated_at": time.Now()}},
	)
	return err
}

func (r *ProjectRepository) SetRunningSource(ctx context.Context, id primitive.ObjectID, source string) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"running_source": source, "updated_at": time.Now()}},
	)
	return err
}
