package repository

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"m3m/internal/domain"
)

// testMongoDB creates a test MongoDB connection
func setupTestMongoDB(t *testing.T) (*MongoDB, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		t.Fatalf("Failed to ping MongoDB: %v", err)
	}

	database := client.Database("m3m_test")

	db := &MongoDB{
		Client:   client,
		Database: database,
	}

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		database.Drop(ctx)
		client.Disconnect(ctx)
	}

	return db, cleanup
}

func createTestModel(projectID primitive.ObjectID) *domain.Model {
	return &domain.Model{
		ID:        primitive.NewObjectID(),
		ProjectID: projectID,
		Name:      "Test Model",
		Slug:      "test_model",
		Fields: []domain.ModelField{
			{Key: "name", Type: domain.FieldTypeString, Required: true},
			{Key: "email", Type: domain.FieldTypeString, Required: true},
			{Key: "counter", Type: domain.FieldTypeNumber, Required: false},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestUpsertData_Insert(t *testing.T) {
	db, cleanup := setupTestMongoDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModel(projectID)

	// Insert the model first
	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Test upsert (should insert new document)
	filter := bson.M{"email": "test@example.com"}
	data := map[string]interface{}{
		"name":  "Test User",
		"email": "test@example.com",
	}

	result, isNew, err := repo.UpsertData(ctx, model, filter, data)
	if err != nil {
		t.Fatalf("UpsertData failed: %v", err)
	}

	if !isNew {
		t.Error("Expected isNew to be true for new document")
	}

	if result == nil {
		t.Fatal("Expected result to not be nil")
	}

	if result.Data["name"] != "Test User" {
		t.Errorf("Expected name 'Test User', got '%v'", result.Data["name"])
	}

	if result.Data["email"] != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%v'", result.Data["email"])
	}
}

func TestUpsertData_Update(t *testing.T) {
	db, cleanup := setupTestMongoDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModel(projectID)

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// First insert
	filter := bson.M{"email": "test@example.com"}
	data := map[string]interface{}{
		"name":  "Original Name",
		"email": "test@example.com",
	}

	_, isNew, err := repo.UpsertData(ctx, model, filter, data)
	if err != nil {
		t.Fatalf("First UpsertData failed: %v", err)
	}
	if !isNew {
		t.Error("First upsert should create new document")
	}

	// Second upsert (should update)
	updatedData := map[string]interface{}{
		"name":  "Updated Name",
		"email": "test@example.com",
	}

	result, isNew, err := repo.UpsertData(ctx, model, filter, updatedData)
	if err != nil {
		t.Fatalf("Second UpsertData failed: %v", err)
	}

	if isNew {
		t.Error("Expected isNew to be false for update")
	}

	if result.Data["name"] != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%v'", result.Data["name"])
	}

	// Verify only one document exists
	collectionName := repo.dataCollectionName(model.ProjectID, model.Slug)
	count, err := db.Collection(collectionName).CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("Failed to count documents: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 document, got %d", count)
	}
}

func TestFindOneAndUpdateData_Basic(t *testing.T) {
	db, cleanup := setupTestMongoDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModel(projectID)

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Create initial document
	_, err := repo.CreateData(ctx, model, map[string]interface{}{
		"name":    "Counter Test",
		"email":   "counter@example.com",
		"counter": int64(0),
	})
	if err != nil {
		t.Fatalf("Failed to create initial data: %v", err)
	}

	// FindOneAndUpdate with $inc
	filter := bson.M{"email": "counter@example.com"}
	updateOps := map[string]interface{}{
		"$inc": bson.M{"counter": int64(1)},
	}

	result, err := repo.FindOneAndUpdateData(ctx, model, filter, updateOps, true)
	if err != nil {
		t.Fatalf("FindOneAndUpdateData failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to not be nil")
	}

	// With returnNew=true, should get updated value
	if result["counter"] != int64(1) {
		t.Errorf("Expected counter to be 1, got %v (type: %T)", result["counter"], result["counter"])
	}
}

func TestFindOneAndUpdateData_ReturnOld(t *testing.T) {
	db, cleanup := setupTestMongoDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModel(projectID)

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Create initial document with counter = 5
	_, err := repo.CreateData(ctx, model, map[string]interface{}{
		"name":    "Counter Test",
		"email":   "counter@example.com",
		"counter": int64(5),
	})
	if err != nil {
		t.Fatalf("Failed to create initial data: %v", err)
	}

	// FindOneAndUpdate with returnNew=false (return old document)
	filter := bson.M{"email": "counter@example.com"}
	updateOps := map[string]interface{}{
		"$inc": bson.M{"counter": int64(1)},
	}

	result, err := repo.FindOneAndUpdateData(ctx, model, filter, updateOps, false)
	if err != nil {
		t.Fatalf("FindOneAndUpdateData failed: %v", err)
	}

	// With returnNew=false, should get old value (5)
	if result["counter"] != int64(5) {
		t.Errorf("Expected counter to be 5 (old value), got %v", result["counter"])
	}

	// Verify the document was actually updated
	doc, err := repo.FindDataByID(ctx, model, result["_id"].(primitive.ObjectID))
	if err != nil {
		t.Fatalf("Failed to find document: %v", err)
	}
	if doc["counter"] != int64(6) {
		t.Errorf("Expected counter to be 6 in database, got %v", doc["counter"])
	}
}

func TestFindOneAndUpdateData_SetOperator(t *testing.T) {
	db, cleanup := setupTestMongoDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModel(projectID)

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Create initial document
	_, err := repo.CreateData(ctx, model, map[string]interface{}{
		"name":  "Original Name",
		"email": "set@example.com",
	})
	if err != nil {
		t.Fatalf("Failed to create initial data: %v", err)
	}

	// FindOneAndUpdate with $set
	filter := bson.M{"email": "set@example.com"}
	updateOps := map[string]interface{}{
		"$set": map[string]interface{}{
			"name": "Updated Name",
		},
	}

	result, err := repo.FindOneAndUpdateData(ctx, model, filter, updateOps, true)
	if err != nil {
		t.Fatalf("FindOneAndUpdateData failed: %v", err)
	}

	if result["name"] != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%v'", result["name"])
	}
}

func TestFindOneAndUpdateData_NotFound(t *testing.T) {
	db, cleanup := setupTestMongoDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModel(projectID)

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Try to update non-existent document
	filter := bson.M{"email": "nonexistent@example.com"}
	updateOps := map[string]interface{}{
		"$inc": bson.M{"counter": int64(1)},
	}

	_, err := repo.FindOneAndUpdateData(ctx, model, filter, updateOps, true)
	if err != ErrDataNotFound {
		t.Errorf("Expected ErrDataNotFound, got %v", err)
	}
}

func TestFindOneAndUpdateData_MultipleIncrements(t *testing.T) {
	db, cleanup := setupTestMongoDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModel(projectID)

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Create initial document
	_, err := repo.CreateData(ctx, model, map[string]interface{}{
		"name":    "Counter Test",
		"email":   "multi@example.com",
		"counter": int64(0),
	})
	if err != nil {
		t.Fatalf("Failed to create initial data: %v", err)
	}

	filter := bson.M{"email": "multi@example.com"}

	// Simulate atomic counter increments
	for i := 0; i < 5; i++ {
		updateOps := map[string]interface{}{
			"$inc": bson.M{"counter": int64(1)},
		}
		_, err := repo.FindOneAndUpdateData(ctx, model, filter, updateOps, true)
		if err != nil {
			t.Fatalf("Increment %d failed: %v", i, err)
		}
	}

	// Verify final counter value
	results, _, err := repo.FindData(ctx, model, &domain.DataQuery{
		Filters: map[string]string{"email": "multi@example.com"},
		Limit:   1,
	})
	if err != nil {
		t.Fatalf("FindData failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	if results[0]["counter"] != int64(5) {
		t.Errorf("Expected counter to be 5 after 5 increments, got %v", results[0]["counter"])
	}
}
