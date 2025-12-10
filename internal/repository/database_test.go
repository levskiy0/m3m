package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/levskiy0/m3m/internal/domain"
)

// getTestDBURI returns the database URI based on TEST_DB_DRIVER env var
func getTestDBURI() (uri string, dbName string) {
	driver := os.Getenv("TEST_DB_DRIVER")
	switch driver {
	case "ferretdb":
		uri = os.Getenv("TEST_FERRETDB_URI")
		if uri == "" {
			uri = "mongodb://localhost:27018"
		}
		dbName = "m3m_test"
	default: // mongodb
		uri = os.Getenv("TEST_MONGODB_URI")
		if uri == "" {
			uri = "mongodb://localhost:27017"
		}
		dbName = "m3m_test"
	}
	return
}

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) (*MongoDB, func()) {
	uri, dbName := getTestDBURI()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		t.Skipf("Skipping test: could not connect to database at %s: %v", uri, err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		t.Skipf("Skipping test: could not ping database at %s: %v", uri, err)
	}

	database := client.Database(dbName)

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

func createTestModelForDB(projectID primitive.ObjectID, slug string) *domain.Model {
	return &domain.Model{
		ID:        primitive.NewObjectID(),
		ProjectID: projectID,
		Name:      "Test Model",
		Slug:      slug,
		Fields: []domain.ModelField{
			{Key: "name", Type: domain.FieldTypeString, Required: true},
			{Key: "email", Type: domain.FieldTypeString, Required: true},
			{Key: "age", Type: domain.FieldTypeNumber, Required: false},
			{Key: "score", Type: domain.FieldTypeFloat, Required: false},
			{Key: "active", Type: domain.FieldTypeBool, Required: false},
			{Key: "tags", Type: domain.FieldTypeMultiSelect, Required: false},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// =============================================================================
// COLLECTION METHODS TESTS
// =============================================================================

func TestDatabase_Insert(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModelForDB(projectID, "test_insert")

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Test insert
	data := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   int64(25),
	}

	result, err := repo.CreateData(ctx, model, data)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to not be nil")
	}

	if result.ID.IsZero() {
		t.Error("Expected ID to be set")
	}
}

func TestDatabase_Find(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModelForDB(projectID, "test_find")

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Insert test data
	testData := []map[string]interface{}{
		{"name": "Alice", "email": "alice@example.com", "age": int64(30)},
		{"name": "Bob", "email": "bob@example.com", "age": int64(25)},
		{"name": "Charlie", "email": "charlie@example.com", "age": int64(35)},
	}

	for _, data := range testData {
		if _, err := repo.CreateData(ctx, model, data); err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// Test find all
	results, total, err := repo.FindData(ctx, model, &domain.DataQuery{
		Limit: 100,
	})
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}

	if total != 3 {
		t.Errorf("Expected 3 documents, got %d", total)
	}

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}
}

func TestDatabase_FindWithOptions(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModelForDB(projectID, "test_find_options")

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Insert test data
	for i := 0; i < 15; i++ {
		data := map[string]interface{}{
			"name":  "User",
			"email": "user@example.com",
			"age":   int64(20 + i),
		}
		if _, err := repo.CreateData(ctx, model, data); err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// Test pagination
	results, total, err := repo.FindData(ctx, model, &domain.DataQuery{
		Page:  2,
		Limit: 5,
		Sort:  "age",
		Order: "asc",
	})
	if err != nil {
		t.Fatalf("FindWithOptions failed: %v", err)
	}

	if total != 15 {
		t.Errorf("Expected total 15, got %d", total)
	}

	if len(results) != 5 {
		t.Errorf("Expected 5 results on page 2, got %d", len(results))
	}

	// First result on page 2 should have age 25 (skipped 5 items: 20,21,22,23,24)
	if results[0]["age"] != int64(25) {
		t.Errorf("Expected first age on page 2 to be 25, got %v", results[0]["age"])
	}
}

func TestDatabase_FindOne(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModelForDB(projectID, "test_findone")

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Insert test data
	created, err := repo.CreateData(ctx, model, map[string]interface{}{
		"name":  "Unique User",
		"email": "unique@example.com",
		"age":   int64(42),
	})
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Test findOne by ID
	result, err := repo.FindDataByID(ctx, model, created.ID)
	if err != nil {
		t.Fatalf("FindOne failed: %v", err)
	}

	if result["name"] != "Unique User" {
		t.Errorf("Expected name 'Unique User', got '%v'", result["name"])
	}
}

func TestDatabase_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModelForDB(projectID, "test_update")

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Insert test data
	created, err := repo.CreateData(ctx, model, map[string]interface{}{
		"name":  "Original Name",
		"email": "original@example.com",
	})
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Update
	err = repo.UpdateData(ctx, model, created.ID, map[string]interface{}{
		"name": "Updated Name",
	})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// Verify update
	result, err := repo.FindDataByID(ctx, model, created.ID)
	if err != nil {
		t.Fatalf("FindDataByID failed: %v", err)
	}

	if result["name"] != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%v'", result["name"])
	}
}

func TestDatabase_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModelForDB(projectID, "test_delete")

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Insert test data
	created, err := repo.CreateData(ctx, model, map[string]interface{}{
		"name":  "To Delete",
		"email": "delete@example.com",
	})
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Delete
	err = repo.DeleteData(ctx, model, created.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	_, err = repo.FindDataByID(ctx, model, created.ID)
	if err != ErrDataNotFound {
		t.Errorf("Expected ErrDataNotFound, got %v", err)
	}
}

func TestDatabase_Count(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModelForDB(projectID, "test_count")

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Insert test data
	for i := 0; i < 5; i++ {
		data := map[string]interface{}{
			"name":  "User",
			"email": "user@example.com",
			"age":   int64(20 + i),
		}
		if _, err := repo.CreateData(ctx, model, data); err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// Count all
	collName := repo.dataCollectionName(model.ProjectID, model.Slug)
	count, err := db.Collection(collName).CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}

	if count != 5 {
		t.Errorf("Expected count 5, got %d", count)
	}
}

func TestDatabase_Upsert_Insert(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModelForDB(projectID, "test_upsert_insert")

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Upsert (should insert)
	filter := bson.M{"email": "upsert@example.com"}
	data := map[string]interface{}{
		"name":  "Upsert User",
		"email": "upsert@example.com",
	}

	result, isNew, err := repo.UpsertData(ctx, model, filter, data)
	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	if !isNew {
		t.Error("Expected isNew to be true for new document")
	}

	if result.Data["name"] != "Upsert User" {
		t.Errorf("Expected name 'Upsert User', got '%v'", result.Data["name"])
	}
}

func TestDatabase_Upsert_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModelForDB(projectID, "test_upsert_update")

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	filter := bson.M{"email": "upsert2@example.com"}

	// First upsert (insert)
	_, _, err := repo.UpsertData(ctx, model, filter, map[string]interface{}{
		"name":  "Original",
		"email": "upsert2@example.com",
	})
	if err != nil {
		t.Fatalf("First upsert failed: %v", err)
	}

	// Second upsert (update)
	result, isNew, err := repo.UpsertData(ctx, model, filter, map[string]interface{}{
		"name":  "Updated",
		"email": "upsert2@example.com",
	})
	if err != nil {
		t.Fatalf("Second upsert failed: %v", err)
	}

	if isNew {
		t.Error("Expected isNew to be false for update")
	}

	if result.Data["name"] != "Updated" {
		t.Errorf("Expected name 'Updated', got '%v'", result.Data["name"])
	}
}

// =============================================================================
// FILTER OPERATORS TESTS
// =============================================================================

func TestDatabase_Filter_Equals(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModelForDB(projectID, "test_filter_eq")

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Insert test data
	repo.CreateData(ctx, model, map[string]interface{}{"name": "Alice", "email": "alice@example.com", "age": int64(30)})
	repo.CreateData(ctx, model, map[string]interface{}{"name": "Bob", "email": "bob@example.com", "age": int64(25)})

	// Test $eq (implicit)
	results, _, err := repo.FindDataAdvanced(ctx, model, &domain.AdvancedDataQuery{
		Filters: []domain.FilterCondition{
			{Field: "name", Operator: domain.FilterOpEquals, Value: "Alice"},
		},
		Limit: 100,
	})
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if results[0]["name"] != "Alice" {
		t.Errorf("Expected name 'Alice', got '%v'", results[0]["name"])
	}
}

func TestDatabase_Filter_NotEquals(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModelForDB(projectID, "test_filter_ne")

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	repo.CreateData(ctx, model, map[string]interface{}{"name": "Alice", "email": "alice@example.com"})
	repo.CreateData(ctx, model, map[string]interface{}{"name": "Bob", "email": "bob@example.com"})

	results, _, err := repo.FindDataAdvanced(ctx, model, &domain.AdvancedDataQuery{
		Filters: []domain.FilterCondition{
			{Field: "name", Operator: domain.FilterOpNotEquals, Value: "Alice"},
		},
		Limit: 100,
	})
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if results[0]["name"] != "Bob" {
		t.Errorf("Expected name 'Bob', got '%v'", results[0]["name"])
	}
}

func TestDatabase_Filter_Comparison(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModelForDB(projectID, "test_filter_cmp")

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Insert test data with different ages
	for i := 18; i <= 65; i += 5 {
		repo.CreateData(ctx, model, map[string]interface{}{
			"name":  "User",
			"email": "user@example.com",
			"age":   int64(i),
		})
	}

	// Test $gt
	results, _, err := repo.FindDataAdvanced(ctx, model, &domain.AdvancedDataQuery{
		Filters: []domain.FilterCondition{
			{Field: "age", Operator: domain.FilterOpGreater, Value: int64(50)},
		},
		Limit: 100,
	})
	if err != nil {
		t.Fatalf("Find with $gt failed: %v", err)
	}

	for _, r := range results {
		if r["age"].(int64) <= 50 {
			t.Errorf("Expected age > 50, got %v", r["age"])
		}
	}

	// Test $gte
	results, _, err = repo.FindDataAdvanced(ctx, model, &domain.AdvancedDataQuery{
		Filters: []domain.FilterCondition{
			{Field: "age", Operator: domain.FilterOpGreaterEq, Value: int64(53)},
		},
		Limit: 100,
	})
	if err != nil {
		t.Fatalf("Find with $gte failed: %v", err)
	}

	for _, r := range results {
		if r["age"].(int64) < 53 {
			t.Errorf("Expected age >= 53, got %v", r["age"])
		}
	}

	// Test $lt
	results, _, err = repo.FindDataAdvanced(ctx, model, &domain.AdvancedDataQuery{
		Filters: []domain.FilterCondition{
			{Field: "age", Operator: domain.FilterOpLess, Value: int64(25)},
		},
		Limit: 100,
	})
	if err != nil {
		t.Fatalf("Find with $lt failed: %v", err)
	}

	for _, r := range results {
		if r["age"].(int64) >= 25 {
			t.Errorf("Expected age < 25, got %v", r["age"])
		}
	}

	// Test $lte
	results, _, err = repo.FindDataAdvanced(ctx, model, &domain.AdvancedDataQuery{
		Filters: []domain.FilterCondition{
			{Field: "age", Operator: domain.FilterOpLessEq, Value: int64(23)},
		},
		Limit: 100,
	})
	if err != nil {
		t.Fatalf("Find with $lte failed: %v", err)
	}

	for _, r := range results {
		if r["age"].(int64) > 23 {
			t.Errorf("Expected age <= 23, got %v", r["age"])
		}
	}
}

func TestDatabase_Filter_Contains(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModelForDB(projectID, "test_filter_contains")

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	repo.CreateData(ctx, model, map[string]interface{}{"name": "John Doe", "email": "john@example.com"})
	repo.CreateData(ctx, model, map[string]interface{}{"name": "Jane Doe", "email": "jane@example.com"})
	repo.CreateData(ctx, model, map[string]interface{}{"name": "Bob Smith", "email": "bob@example.com"})

	results, _, err := repo.FindDataAdvanced(ctx, model, &domain.AdvancedDataQuery{
		Filters: []domain.FilterCondition{
			{Field: "name", Operator: domain.FilterOpContains, Value: "Doe"},
		},
		Limit: 100,
	})
	if err != nil {
		t.Fatalf("Find with $contains failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results with 'Doe', got %d", len(results))
	}
}

func TestDatabase_Filter_StartsWith(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModelForDB(projectID, "test_filter_starts")

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	repo.CreateData(ctx, model, map[string]interface{}{"name": "John", "email": "john@gmail.com"})
	repo.CreateData(ctx, model, map[string]interface{}{"name": "Jane", "email": "jane@yahoo.com"})

	results, _, err := repo.FindDataAdvanced(ctx, model, &domain.AdvancedDataQuery{
		Filters: []domain.FilterCondition{
			{Field: "name", Operator: domain.FilterOpStartsWith, Value: "Jo"},
		},
		Limit: 100,
	})
	if err != nil {
		t.Fatalf("Find with $startsWith failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result starting with 'Jo', got %d", len(results))
	}
}

func TestDatabase_Filter_EndsWith(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModelForDB(projectID, "test_filter_ends")

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	repo.CreateData(ctx, model, map[string]interface{}{"name": "John", "email": "john@gmail.com"})
	repo.CreateData(ctx, model, map[string]interface{}{"name": "Jane", "email": "jane@yahoo.com"})

	results, _, err := repo.FindDataAdvanced(ctx, model, &domain.AdvancedDataQuery{
		Filters: []domain.FilterCondition{
			{Field: "email", Operator: domain.FilterOpEndsWith, Value: "@gmail.com"},
		},
		Limit: 100,
	})
	if err != nil {
		t.Fatalf("Find with $endsWith failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result ending with '@gmail.com', got %d", len(results))
	}
}

func TestDatabase_Filter_In(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModelForDB(projectID, "test_filter_in")

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	repo.CreateData(ctx, model, map[string]interface{}{"name": "Alice", "email": "alice@example.com"})
	repo.CreateData(ctx, model, map[string]interface{}{"name": "Bob", "email": "bob@example.com"})
	repo.CreateData(ctx, model, map[string]interface{}{"name": "Charlie", "email": "charlie@example.com"})

	results, _, err := repo.FindDataAdvanced(ctx, model, &domain.AdvancedDataQuery{
		Filters: []domain.FilterCondition{
			{Field: "name", Operator: domain.FilterOpIn, Value: []interface{}{"Alice", "Charlie"}},
		},
		Limit: 100,
	})
	if err != nil {
		t.Fatalf("Find with $in failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results with $in, got %d", len(results))
	}
}

func TestDatabase_Filter_NotIn(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModelForDB(projectID, "test_filter_nin")

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	repo.CreateData(ctx, model, map[string]interface{}{"name": "Alice", "email": "alice@example.com"})
	repo.CreateData(ctx, model, map[string]interface{}{"name": "Bob", "email": "bob@example.com"})
	repo.CreateData(ctx, model, map[string]interface{}{"name": "Charlie", "email": "charlie@example.com"})

	results, _, err := repo.FindDataAdvanced(ctx, model, &domain.AdvancedDataQuery{
		Filters: []domain.FilterCondition{
			{Field: "name", Operator: domain.FilterOpNotIn, Value: []interface{}{"Alice", "Charlie"}},
		},
		Limit: 100,
	})
	if err != nil {
		t.Fatalf("Find with $nin failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result with $nin, got %d", len(results))
	}

	if results[0]["name"] != "Bob" {
		t.Errorf("Expected name 'Bob', got '%v'", results[0]["name"])
	}
}

// =============================================================================
// UPDATE OPERATORS TESTS
// =============================================================================

func TestDatabase_Update_Set(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModelForDB(projectID, "test_update_set")

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	repo.CreateData(ctx, model, map[string]interface{}{
		"name":  "Original",
		"email": "set@example.com",
	})

	result, err := repo.FindOneAndUpdateData(ctx, model, bson.M{"email": "set@example.com"}, map[string]interface{}{
		"$set": map[string]interface{}{"name": "Updated via $set"},
	}, true)
	if err != nil {
		t.Fatalf("FindOneAndUpdate with $set failed: %v", err)
	}

	if result["name"] != "Updated via $set" {
		t.Errorf("Expected name 'Updated via $set', got '%v'", result["name"])
	}
}

func TestDatabase_Update_Inc(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModelForDB(projectID, "test_update_inc")

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	repo.CreateData(ctx, model, map[string]interface{}{
		"name":  "Counter",
		"email": "inc@example.com",
		"age":   int64(10),
	})

	result, err := repo.FindOneAndUpdateData(ctx, model, bson.M{"email": "inc@example.com"}, map[string]interface{}{
		"$inc": bson.M{"age": int64(5)},
	}, true)
	if err != nil {
		t.Fatalf("FindOneAndUpdate with $inc failed: %v", err)
	}

	if result["age"] != int64(15) {
		t.Errorf("Expected age 15, got %v", result["age"])
	}

	// Test negative increment
	result, err = repo.FindOneAndUpdateData(ctx, model, bson.M{"email": "inc@example.com"}, map[string]interface{}{
		"$inc": bson.M{"age": int64(-3)},
	}, true)
	if err != nil {
		t.Fatalf("FindOneAndUpdate with negative $inc failed: %v", err)
	}

	if result["age"] != int64(12) {
		t.Errorf("Expected age 12, got %v", result["age"])
	}
}

func TestDatabase_Update_Unset(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModelForDB(projectID, "test_update_unset")

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	repo.CreateData(ctx, model, map[string]interface{}{
		"name":  "With Field",
		"email": "unset@example.com",
		"age":   int64(30),
	})

	result, err := repo.FindOneAndUpdateData(ctx, model, bson.M{"email": "unset@example.com"}, map[string]interface{}{
		"$unset": bson.M{"age": ""},
	}, true)
	if err != nil {
		t.Fatalf("FindOneAndUpdate with $unset failed: %v", err)
	}

	if _, exists := result["age"]; exists {
		t.Errorf("Expected 'age' field to be removed, but it still exists: %v", result["age"])
	}
}

func TestDatabase_Update_Push(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModelForDB(projectID, "test_update_push")

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	repo.CreateData(ctx, model, map[string]interface{}{
		"name":  "With Tags",
		"email": "push@example.com",
		"tags":  []interface{}{"tag1"},
	})

	result, err := repo.FindOneAndUpdateData(ctx, model, bson.M{"email": "push@example.com"}, map[string]interface{}{
		"$push": bson.M{"tags": "tag2"},
	}, true)
	if err != nil {
		t.Fatalf("FindOneAndUpdate with $push failed: %v", err)
	}

	tags, ok := result["tags"].(primitive.A)
	if !ok {
		t.Fatalf("Expected tags to be array, got %T", result["tags"])
	}

	if len(tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(tags))
	}
}

func TestDatabase_Update_Pull(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModelForDB(projectID, "test_update_pull")

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	repo.CreateData(ctx, model, map[string]interface{}{
		"name":  "With Tags",
		"email": "pull@example.com",
		"tags":  []interface{}{"tag1", "tag2", "tag3"},
	})

	result, err := repo.FindOneAndUpdateData(ctx, model, bson.M{"email": "pull@example.com"}, map[string]interface{}{
		"$pull": bson.M{"tags": "tag2"},
	}, true)
	if err != nil {
		t.Fatalf("FindOneAndUpdate with $pull failed: %v", err)
	}

	tags, ok := result["tags"].(primitive.A)
	if !ok {
		t.Fatalf("Expected tags to be array, got %T", result["tags"])
	}

	if len(tags) != 2 {
		t.Errorf("Expected 2 tags after pull, got %d", len(tags))
	}

	for _, tag := range tags {
		if tag == "tag2" {
			t.Error("tag2 should have been removed")
		}
	}
}

func TestDatabase_Update_AddToSet(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModelForDB(projectID, "test_update_addtoset")

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	repo.CreateData(ctx, model, map[string]interface{}{
		"name":  "With Tags",
		"email": "addtoset@example.com",
		"tags":  []interface{}{"tag1"},
	})

	// Add new unique element
	result, err := repo.FindOneAndUpdateData(ctx, model, bson.M{"email": "addtoset@example.com"}, map[string]interface{}{
		"$addToSet": bson.M{"tags": "tag2"},
	}, true)
	if err != nil {
		t.Fatalf("FindOneAndUpdate with $addToSet failed: %v", err)
	}

	tags, ok := result["tags"].(primitive.A)
	if !ok {
		t.Fatalf("Expected tags to be array, got %T", result["tags"])
	}

	if len(tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(tags))
	}

	// Try adding duplicate (should not add)
	result, err = repo.FindOneAndUpdateData(ctx, model, bson.M{"email": "addtoset@example.com"}, map[string]interface{}{
		"$addToSet": bson.M{"tags": "tag1"},
	}, true)
	if err != nil {
		t.Fatalf("FindOneAndUpdate with duplicate $addToSet failed: %v", err)
	}

	tags, _ = result["tags"].(primitive.A)
	if len(tags) != 2 {
		t.Errorf("Expected 2 tags (no duplicate), got %d", len(tags))
	}
}

// =============================================================================
// COMPLEX QUERIES TESTS
// =============================================================================

func TestDatabase_ComplexQuery(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModelForDB(projectID, "test_complex")

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Insert test data
	testData := []map[string]interface{}{
		{"name": "Alice", "email": "alice@gmail.com", "age": int64(25), "active": true},
		{"name": "Bob", "email": "bob@gmail.com", "age": int64(30), "active": false},
		{"name": "Charlie", "email": "charlie@yahoo.com", "age": int64(35), "active": true},
		{"name": "Diana", "email": "diana@gmail.com", "age": int64(28), "active": true},
	}

	for _, data := range testData {
		if _, err := repo.CreateData(ctx, model, data); err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// Complex query: active users, age >= 25, age <= 30
	results, _, err := repo.FindDataAdvanced(ctx, model, &domain.AdvancedDataQuery{
		Filters: []domain.FilterCondition{
			{Field: "active", Operator: domain.FilterOpEquals, Value: true},
			{Field: "age", Operator: domain.FilterOpGreaterEq, Value: int64(25)},
			{Field: "age", Operator: domain.FilterOpLessEq, Value: int64(30)},
		},
		Sort:  "age",
		Order: "asc",
		Limit: 100,
	})
	if err != nil {
		t.Fatalf("Complex query failed: %v", err)
	}

	// Should return Alice (25) and Diana (28), not Bob (inactive) or Charlie (35)
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	if results[0]["name"] != "Alice" {
		t.Errorf("First result should be Alice (sorted by age), got %v", results[0]["name"])
	}

	if results[1]["name"] != "Diana" {
		t.Errorf("Second result should be Diana, got %v", results[1]["name"])
	}
}

func TestDatabase_AtomicCounter(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()
	model := createTestModelForDB(projectID, "test_atomic")

	if err := repo.Create(ctx, model); err != nil {
		t.Fatalf("Failed to create model: %v", err)
	}

	// Create document with counter
	repo.CreateData(ctx, model, map[string]interface{}{
		"name":  "Counter Doc",
		"email": "counter@example.com",
		"age":   int64(0),
	})

	// Simulate concurrent increments
	for i := 0; i < 10; i++ {
		_, err := repo.FindOneAndUpdateData(ctx, model, bson.M{"email": "counter@example.com"}, map[string]interface{}{
			"$inc": bson.M{"age": int64(1)},
		}, true)
		if err != nil {
			t.Fatalf("Increment %d failed: %v", i, err)
		}
	}

	// Verify final value
	results, _, err := repo.FindDataAdvanced(ctx, model, &domain.AdvancedDataQuery{
		Filters: []domain.FilterCondition{
			{Field: "email", Operator: domain.FilterOpEquals, Value: "counter@example.com"},
		},
		Limit: 1,
	})
	if err != nil {
		t.Fatalf("Find failed: %v", err)
	}

	if results[0]["age"] != int64(10) {
		t.Errorf("Expected counter to be 10 after 10 increments, got %v", results[0]["age"])
	}
}
