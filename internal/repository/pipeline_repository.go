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
	ErrBranchNotFound      = errors.New("branch not found")
	ErrBranchAlreadyExists = errors.New("branch already exists")
	ErrReleaseNotFound     = errors.New("release not found")
	ErrReleaseExists       = errors.New("release version already exists")
)

type PipelineRepository struct {
	branchesCollection *mongo.Collection
	releasesCollection *mongo.Collection
}

func NewPipelineRepository(db *MongoDB) *PipelineRepository {
	branchesCollection := db.Collection("branches")
	releasesCollection := db.Collection("releases")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	branchesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
			{Key: "name", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})

	releasesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "project_id", Value: 1},
			{Key: "version", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})

	return &PipelineRepository{
		branchesCollection: branchesCollection,
		releasesCollection: releasesCollection,
	}
}

// Branch methods
func (r *PipelineRepository) CreateBranch(ctx context.Context, branch *domain.Branch) error {
	branch.ID = primitive.NewObjectID()
	branch.CreatedAt = time.Now()
	branch.UpdatedAt = time.Now()

	_, err := r.branchesCollection.InsertOne(ctx, branch)
	if mongo.IsDuplicateKeyError(err) {
		return ErrBranchAlreadyExists
	}
	return err
}

func (r *PipelineRepository) FindBranchByID(ctx context.Context, id primitive.ObjectID) (*domain.Branch, error) {
	var branch domain.Branch
	err := r.branchesCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&branch)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrBranchNotFound
	}
	return &branch, err
}

func (r *PipelineRepository) FindBranchByName(ctx context.Context, projectID primitive.ObjectID, name string) (*domain.Branch, error) {
	var branch domain.Branch
	err := r.branchesCollection.FindOne(ctx, bson.M{
		"project_id": projectID,
		"name":       name,
	}).Decode(&branch)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrBranchNotFound
	}
	return &branch, err
}

func (r *PipelineRepository) FindBranchesByProject(ctx context.Context, projectID primitive.ObjectID) ([]*domain.Branch, error) {
	cursor, err := r.branchesCollection.Find(ctx, bson.M{"project_id": projectID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	branches := make([]*domain.Branch, 0)
	if err := cursor.All(ctx, &branches); err != nil {
		return nil, err
	}
	return branches, nil
}

func (r *PipelineRepository) FindBranchSummariesByProject(ctx context.Context, projectID primitive.ObjectID) ([]*domain.BranchSummary, error) {
	opts := options.Find().SetProjection(bson.M{"files": 0})
	cursor, err := r.branchesCollection.Find(ctx, bson.M{"project_id": projectID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	branches := make([]*domain.BranchSummary, 0)
	if err := cursor.All(ctx, &branches); err != nil {
		return nil, err
	}
	return branches, nil
}

func (r *PipelineRepository) UpdateBranch(ctx context.Context, branch *domain.Branch) error {
	branch.UpdatedAt = time.Now()
	_, err := r.branchesCollection.UpdateOne(
		ctx,
		bson.M{"_id": branch.ID},
		bson.M{"$set": branch},
	)
	return err
}

func (r *PipelineRepository) DeleteBranch(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.branchesCollection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrBranchNotFound
	}
	return nil
}

// Release methods
func (r *PipelineRepository) CreateRelease(ctx context.Context, release *domain.Release) error {
	release.ID = primitive.NewObjectID()
	release.CreatedAt = time.Now()

	_, err := r.releasesCollection.InsertOne(ctx, release)
	if mongo.IsDuplicateKeyError(err) {
		return ErrReleaseExists
	}
	return err
}

func (r *PipelineRepository) FindReleaseByID(ctx context.Context, id primitive.ObjectID) (*domain.Release, error) {
	var release domain.Release
	err := r.releasesCollection.FindOne(ctx, bson.M{"_id": id}).Decode(&release)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrReleaseNotFound
	}
	return &release, err
}

func (r *PipelineRepository) FindReleaseByVersion(ctx context.Context, projectID primitive.ObjectID, version string) (*domain.Release, error) {
	var release domain.Release
	err := r.releasesCollection.FindOne(ctx, bson.M{
		"project_id": projectID,
		"version":    version,
	}).Decode(&release)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrReleaseNotFound
	}
	return &release, err
}

func (r *PipelineRepository) FindReleasesByProject(ctx context.Context, projectID primitive.ObjectID) ([]*domain.Release, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.releasesCollection.Find(ctx, bson.M{"project_id": projectID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	releases := make([]*domain.Release, 0)
	if err := cursor.All(ctx, &releases); err != nil {
		return nil, err
	}
	return releases, nil
}

func (r *PipelineRepository) FindReleaseSummariesByProject(ctx context.Context, projectID primitive.ObjectID) ([]*domain.ReleaseSummary, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetProjection(bson.M{"files": 0})
	cursor, err := r.releasesCollection.Find(ctx, bson.M{"project_id": projectID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	releases := make([]*domain.ReleaseSummary, 0)
	if err := cursor.All(ctx, &releases); err != nil {
		return nil, err
	}
	return releases, nil
}

func (r *PipelineRepository) FindActiveRelease(ctx context.Context, projectID primitive.ObjectID) (*domain.Release, error) {
	var release domain.Release
	err := r.releasesCollection.FindOne(ctx, bson.M{
		"project_id": projectID,
		"is_active":  true,
	}).Decode(&release)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrReleaseNotFound
	}
	return &release, err
}

func (r *PipelineRepository) FindLatestRelease(ctx context.Context, projectID primitive.ObjectID) (*domain.Release, error) {
	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})
	var release domain.Release
	err := r.releasesCollection.FindOne(ctx, bson.M{"project_id": projectID}, opts).Decode(&release)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, ErrReleaseNotFound
	}
	return &release, err
}

func (r *PipelineRepository) ActivateRelease(ctx context.Context, projectID primitive.ObjectID, version string) error {
	// Deactivate all releases for project
	_, err := r.releasesCollection.UpdateMany(
		ctx,
		bson.M{"project_id": projectID},
		bson.M{"$set": bson.M{"is_active": false}},
	)
	if err != nil {
		return err
	}

	// Activate specific release
	_, err = r.releasesCollection.UpdateOne(
		ctx,
		bson.M{"project_id": projectID, "version": version},
		bson.M{"$set": bson.M{"is_active": true}},
	)
	return err
}

func (r *PipelineRepository) DeleteRelease(ctx context.Context, id primitive.ObjectID) error {
	result, err := r.releasesCollection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrReleaseNotFound
	}
	return nil
}

// UpdateBranchFile updates a single file's code in a branch
func (r *PipelineRepository) UpdateBranchFile(ctx context.Context, branchID primitive.ObjectID, fileName string, code string) error {
	result, err := r.branchesCollection.UpdateOne(
		ctx,
		bson.M{
			"_id":        branchID,
			"files.name": fileName,
		},
		bson.M{
			"$set": bson.M{
				"files.$.code": code,
				"updated_at":   time.Now(),
			},
		},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrBranchNotFound
	}
	return nil
}

// AddFileToBranch adds a new file to a branch
func (r *PipelineRepository) AddFileToBranch(ctx context.Context, branchID primitive.ObjectID, file domain.CodeFile) error {
	result, err := r.branchesCollection.UpdateOne(
		ctx,
		bson.M{"_id": branchID},
		bson.M{
			"$push": bson.M{"files": file},
			"$set":  bson.M{"updated_at": time.Now()},
		},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrBranchNotFound
	}
	return nil
}

// DeleteFileFromBranch removes a file from a branch
func (r *PipelineRepository) DeleteFileFromBranch(ctx context.Context, branchID primitive.ObjectID, fileName string) error {
	result, err := r.branchesCollection.UpdateOne(
		ctx,
		bson.M{"_id": branchID},
		bson.M{
			"$pull": bson.M{"files": bson.M{"name": fileName}},
			"$set":  bson.M{"updated_at": time.Now()},
		},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrBranchNotFound
	}
	return nil
}

// RenameFileInBranch renames a file in a branch
func (r *PipelineRepository) RenameFileInBranch(ctx context.Context, branchID primitive.ObjectID, oldName, newName string) error {
	result, err := r.branchesCollection.UpdateOne(
		ctx,
		bson.M{
			"_id":        branchID,
			"files.name": oldName,
		},
		bson.M{
			"$set": bson.M{
				"files.$.name": newName,
				"updated_at":   time.Now(),
			},
		},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrBranchNotFound
	}
	return nil
}
