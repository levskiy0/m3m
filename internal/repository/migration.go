package repository

import (
	"context"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// MigrateCodeToFiles migrates old code field to new files array format
// This runs automatically on app startup and is idempotent
func MigrateCodeToFiles(db *mongo.Database, logger *slog.Logger) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Migrate branches
	branchCount, err := migrateBranches(ctx, db.Collection("branches"), logger)
	if err != nil {
		return err
	}

	// Migrate releases
	releaseCount, err := migrateReleases(ctx, db.Collection("releases"), logger)
	if err != nil {
		return err
	}

	if branchCount > 0 || releaseCount > 0 {
		logger.Info("Migration completed",
			"branches_migrated", branchCount,
			"releases_migrated", releaseCount,
		)
	}

	return nil
}

func migrateBranches(ctx context.Context, collection *mongo.Collection, logger *slog.Logger) (int, error) {
	// Find documents that have 'code' field but no 'files' field
	filter := bson.M{
		"code":  bson.M{"$exists": true},
		"files": bson.M{"$exists": false},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	migrated := 0
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			logger.Error("Failed to decode branch document", "error", err)
			continue
		}

		id := doc["_id"]
		code, _ := doc["code"].(string)

		// Create files array with main file
		files := []bson.M{
			{"name": "main", "code": code},
		}

		// Update document: add files, remove code
		_, err := collection.UpdateOne(
			ctx,
			bson.M{"_id": id},
			bson.M{
				"$set":   bson.M{"files": files},
				"$unset": bson.M{"code": ""},
			},
		)
		if err != nil {
			logger.Error("Failed to migrate branch", "id", id, "error", err)
			continue
		}

		migrated++
		logger.Debug("Migrated branch", "id", id)
	}

	return migrated, cursor.Err()
}

func migrateReleases(ctx context.Context, collection *mongo.Collection, logger *slog.Logger) (int, error) {
	// Find documents that have 'code' field but no 'files' field
	filter := bson.M{
		"code":  bson.M{"$exists": true},
		"files": bson.M{"$exists": false},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	migrated := 0
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			logger.Error("Failed to decode release document", "error", err)
			continue
		}

		id := doc["_id"]
		code, _ := doc["code"].(string)

		// Create files array with main file
		files := []bson.M{
			{"name": "main", "code": code},
		}

		// Update document: add files, remove code
		_, err := collection.UpdateOne(
			ctx,
			bson.M{"_id": id},
			bson.M{
				"$set":   bson.M{"files": files},
				"$unset": bson.M{"code": ""},
			},
		)
		if err != nil {
			logger.Error("Failed to migrate release", "id", id, "error", err)
			continue
		}

		migrated++
		logger.Debug("Migrated release", "id", id)
	}

	return migrated, cursor.Err()
}
