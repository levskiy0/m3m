package repository

import (
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"m3m/internal/domain"
)

const testMongoURI = "mongodb://localhost:27017"
const testDatabaseName = "m3m_testing_database"

// setupTestDB creates a test database connection and cleans up existing data
func setupTestDB(t *testing.T) (*MongoDB, func()) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(testMongoURI))
	if err != nil {
		t.Skipf("Skipping test: MongoDB not available at %s: %v", testMongoURI, err)
		return nil, func() {}
	}

	// Check connection
	if err := client.Ping(ctx, nil); err != nil {
		t.Skipf("Skipping test: MongoDB ping failed: %v", err)
		return nil, func() {}
	}

	db := client.Database(testDatabaseName)

	// Drop all collections for clean state
	collections, _ := db.ListCollectionNames(ctx, map[string]interface{}{})
	for _, col := range collections {
		db.Collection(col).Drop(ctx)
	}

	mongodb := &MongoDB{
		Client:   client,
		Database: db,
	}

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		// Drop database after tests
		db.Drop(ctx)
		client.Disconnect(ctx)
	}

	return mongodb, cleanup
}

func TestModelRepository_CRUD(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	projectID := primitive.NewObjectID()

	// Test Create
	t.Run("Create", func(t *testing.T) {
		model := &domain.Model{
			ProjectID: projectID,
			Name:      "Test Model",
			Slug:      "test-model",
			Fields: []domain.ModelField{
				{Key: "name", Type: domain.FieldTypeString, Required: true},
				{Key: "age", Type: domain.FieldTypeNumber, Required: false},
			},
			TableConfig: domain.TableConfig{
				Columns:     []string{"name", "age"},
				Filters:     []string{"name"},
				SortColumns: []string{"name"},
			},
			FormConfig: domain.FormConfig{
				FieldOrder:   []string{"name", "age"},
				HiddenFields: []string{},
				FieldViews:   map[string]string{},
			},
		}

		err := repo.Create(ctx, model)
		if err != nil {
			t.Fatalf("Failed to create model: %v", err)
		}

		if model.ID.IsZero() {
			t.Error("Model ID should be set after create")
		}
		if model.CreatedAt.IsZero() {
			t.Error("CreatedAt should be set after create")
		}
	})

	// Test FindBySlug
	t.Run("FindBySlug", func(t *testing.T) {
		model, err := repo.FindBySlug(ctx, projectID, "test-model")
		if err != nil {
			t.Fatalf("Failed to find model by slug: %v", err)
		}

		if model.Name != "Test Model" {
			t.Errorf("Expected name 'Test Model', got %q", model.Name)
		}
		if len(model.Fields) != 2 {
			t.Errorf("Expected 2 fields, got %d", len(model.Fields))
		}
	})

	// Test FindByProject
	t.Run("FindByProject", func(t *testing.T) {
		models, err := repo.FindByProject(ctx, projectID)
		if err != nil {
			t.Fatalf("Failed to find models by project: %v", err)
		}

		if len(models) != 1 {
			t.Errorf("Expected 1 model, got %d", len(models))
		}
	})

	// Test duplicate slug error
	t.Run("DuplicateSlug", func(t *testing.T) {
		model := &domain.Model{
			ProjectID: projectID,
			Name:      "Duplicate Model",
			Slug:      "test-model", // Same slug as before
			Fields: []domain.ModelField{
				{Key: "field1", Type: domain.FieldTypeString},
			},
		}

		err := repo.Create(ctx, model)
		if err != ErrModelSlugExists {
			t.Errorf("Expected ErrModelSlugExists, got %v", err)
		}
	})

	// Test Update
	t.Run("Update", func(t *testing.T) {
		model, _ := repo.FindBySlug(ctx, projectID, "test-model")
		model.Name = "Updated Model"
		model.Fields = append(model.Fields, domain.ModelField{
			Key:  "email",
			Type: domain.FieldTypeString,
		})

		err := repo.Update(ctx, model)
		if err != nil {
			t.Fatalf("Failed to update model: %v", err)
		}

		updated, _ := repo.FindByID(ctx, model.ID)
		if updated.Name != "Updated Model" {
			t.Errorf("Expected name 'Updated Model', got %q", updated.Name)
		}
		if len(updated.Fields) != 3 {
			t.Errorf("Expected 3 fields, got %d", len(updated.Fields))
		}
	})

	// Test Delete
	t.Run("Delete", func(t *testing.T) {
		model, _ := repo.FindBySlug(ctx, projectID, "test-model")

		err := repo.Delete(ctx, model.ID)
		if err != nil {
			t.Fatalf("Failed to delete model: %v", err)
		}

		_, err = repo.FindByID(ctx, model.ID)
		if err != ErrModelNotFound {
			t.Errorf("Expected ErrModelNotFound after delete, got %v", err)
		}
	})
}

func TestModelRepository_DataCRUD(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	// Create a model first
	projectID := primitive.NewObjectID()
	model := &domain.Model{
		ProjectID: projectID,
		Name:      "Users",
		Slug:      "users",
		Fields: []domain.ModelField{
			{Key: "name", Type: domain.FieldTypeString, Required: true},
			{Key: "age", Type: domain.FieldTypeNumber, Required: false},
			{Key: "active", Type: domain.FieldTypeBool, Required: false},
		},
	}
	repo.Create(ctx, model)

	var dataID primitive.ObjectID

	// Test CreateData
	t.Run("CreateData", func(t *testing.T) {
		data := map[string]interface{}{
			"name":   "John Doe",
			"age":    30,
			"active": true,
		}

		result, err := repo.CreateData(ctx, model, data)
		if err != nil {
			t.Fatalf("Failed to create data: %v", err)
		}

		if result.ID.IsZero() {
			t.Error("Data ID should be set after create")
		}
		dataID = result.ID

		if result.Data["name"] != "John Doe" {
			t.Errorf("Expected name 'John Doe', got %v", result.Data["name"])
		}
	})

	// Test FindDataByID
	t.Run("FindDataByID", func(t *testing.T) {
		data, err := repo.FindDataByID(ctx, model, dataID)
		if err != nil {
			t.Fatalf("Failed to find data by ID: %v", err)
		}

		if data["name"] != "John Doe" {
			t.Errorf("Expected name 'John Doe', got %v", data["name"])
		}
	})

	// Test FindData with query
	t.Run("FindData", func(t *testing.T) {
		// Add more test data
		repo.CreateData(ctx, model, map[string]interface{}{"name": "Jane Doe", "age": 25, "active": true})
		repo.CreateData(ctx, model, map[string]interface{}{"name": "Bob Smith", "age": 35, "active": false})

		query := &domain.DataQuery{
			Page:  1,
			Limit: 10,
		}

		results, total, err := repo.FindData(ctx, model, query)
		if err != nil {
			t.Fatalf("Failed to find data: %v", err)
		}

		if total != 3 {
			t.Errorf("Expected total 3, got %d", total)
		}
		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}
	})

	// Test FindData with filter
	t.Run("FindData_WithFilter", func(t *testing.T) {
		query := &domain.DataQuery{
			Page:    1,
			Limit:   10,
			Filters: map[string]string{"name": "John Doe"},
		}

		results, total, err := repo.FindData(ctx, model, query)
		if err != nil {
			t.Fatalf("Failed to find data with filter: %v", err)
		}

		if total != 1 {
			t.Errorf("Expected total 1, got %d", total)
		}
		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}
	})

	// Test FindData with sorting
	t.Run("FindData_WithSorting", func(t *testing.T) {
		query := &domain.DataQuery{
			Page:  1,
			Limit: 10,
			Sort:  "age",
			Order: "asc",
		}

		results, _, err := repo.FindData(ctx, model, query)
		if err != nil {
			t.Fatalf("Failed to find data with sorting: %v", err)
		}

		// First result should have lowest age (25)
		if results[0]["age"].(int32) != 25 {
			t.Errorf("Expected first result age 25, got %v", results[0]["age"])
		}
	})

	// Test FindData with pagination
	t.Run("FindData_WithPagination", func(t *testing.T) {
		query := &domain.DataQuery{
			Page:  1,
			Limit: 2,
		}

		results, total, err := repo.FindData(ctx, model, query)
		if err != nil {
			t.Fatalf("Failed to find data with pagination: %v", err)
		}

		if total != 3 {
			t.Errorf("Expected total 3, got %d", total)
		}
		if len(results) != 2 {
			t.Errorf("Expected 2 results (limit 2), got %d", len(results))
		}
	})

	// Test UpdateData
	t.Run("UpdateData", func(t *testing.T) {
		err := repo.UpdateData(ctx, model, dataID, map[string]interface{}{
			"age": 31,
		})
		if err != nil {
			t.Fatalf("Failed to update data: %v", err)
		}

		data, _ := repo.FindDataByID(ctx, model, dataID)
		if data["age"].(int32) != 31 {
			t.Errorf("Expected age 31, got %v", data["age"])
		}
	})

	// Test DeleteData
	t.Run("DeleteData", func(t *testing.T) {
		err := repo.DeleteData(ctx, model, dataID)
		if err != nil {
			t.Fatalf("Failed to delete data: %v", err)
		}

		_, err = repo.FindDataByID(ctx, model, dataID)
		if err != ErrDataNotFound {
			t.Errorf("Expected ErrDataNotFound after delete, got %v", err)
		}
	})

	// Test DropDataCollection
	t.Run("DropDataCollection", func(t *testing.T) {
		err := repo.DropDataCollection(ctx, model)
		if err != nil {
			t.Fatalf("Failed to drop data collection: %v", err)
		}

		query := &domain.DataQuery{Page: 1, Limit: 10}
		results, total, _ := repo.FindData(ctx, model, query)
		if total != 0 || len(results) != 0 {
			t.Error("Expected empty collection after drop")
		}
	})
}

func TestModelRepository_AdvancedFiltering(t *testing.T) {
	db, cleanup := setupTestDB(t)
	if db == nil {
		return
	}
	defer cleanup()

	repo := NewModelRepository(db)
	ctx := context.Background()

	// Create a model
	projectID := primitive.NewObjectID()
	model := &domain.Model{
		ProjectID: projectID,
		Name:      "Products",
		Slug:      "products",
		Fields: []domain.ModelField{
			{Key: "name", Type: domain.FieldTypeString},
			{Key: "price", Type: domain.FieldTypeFloat},
			{Key: "quantity", Type: domain.FieldTypeNumber},
			{Key: "active", Type: domain.FieldTypeBool},
		},
	}
	repo.Create(ctx, model)

	// Add test data
	testData := []map[string]interface{}{
		{"name": "Apple", "price": 1.5, "quantity": 100, "active": true},
		{"name": "Banana", "price": 0.75, "quantity": 150, "active": true},
		{"name": "Orange", "price": 2.0, "quantity": 80, "active": false},
		{"name": "Grapes", "price": 3.5, "quantity": 50, "active": true},
	}
	for _, data := range testData {
		repo.CreateData(ctx, model, data)
	}

	// Test greater than filter
	t.Run("GreaterThan", func(t *testing.T) {
		query := &domain.AdvancedDataQuery{
			Page:  1,
			Limit: 10,
			Filters: []domain.FilterCondition{
				{Field: "price", Operator: domain.FilterOpGreater, Value: 1.5},
			},
		}

		results, total, err := repo.FindDataAdvanced(ctx, model, query)
		if err != nil {
			t.Fatalf("Failed: %v", err)
		}

		if total != 2 { // Orange (2.0) and Grapes (3.5)
			t.Errorf("Expected 2 results, got %d", total)
		}
		_ = results
	})

	// Test less than or equal filter
	t.Run("LessThanOrEqual", func(t *testing.T) {
		query := &domain.AdvancedDataQuery{
			Page:  1,
			Limit: 10,
			Filters: []domain.FilterCondition{
				{Field: "price", Operator: domain.FilterOpLessEq, Value: 1.5},
			},
		}

		results, total, err := repo.FindDataAdvanced(ctx, model, query)
		if err != nil {
			t.Fatalf("Failed: %v", err)
		}

		if total != 2 { // Apple (1.5) and Banana (0.75)
			t.Errorf("Expected 2 results, got %d", total)
		}
		_ = results
	})

	// Test contains filter
	t.Run("Contains", func(t *testing.T) {
		query := &domain.AdvancedDataQuery{
			Page:  1,
			Limit: 10,
			Filters: []domain.FilterCondition{
				{Field: "name", Operator: domain.FilterOpContains, Value: "an"},
			},
		}

		results, total, err := repo.FindDataAdvanced(ctx, model, query)
		if err != nil {
			t.Fatalf("Failed: %v", err)
		}

		if total != 2 { // Banana and Orange
			t.Errorf("Expected 2 results, got %d", total)
		}
		_ = results
	})

	// Test multiple filters (AND)
	t.Run("MultipleFilters", func(t *testing.T) {
		query := &domain.AdvancedDataQuery{
			Page:  1,
			Limit: 10,
			Filters: []domain.FilterCondition{
				{Field: "price", Operator: domain.FilterOpGreater, Value: 1.0},
				{Field: "active", Operator: domain.FilterOpEquals, Value: true},
			},
		}

		results, total, err := repo.FindDataAdvanced(ctx, model, query)
		if err != nil {
			t.Fatalf("Failed: %v", err)
		}

		if total != 2 { // Apple (1.5, active) and Grapes (3.5, active)
			t.Errorf("Expected 2 results, got %d", total)
		}
		_ = results
	})

	// Test search
	t.Run("Search", func(t *testing.T) {
		query := &domain.AdvancedDataQuery{
			Page:     1,
			Limit:    10,
			Search:   "ap",
			SearchIn: []string{"name"},
		}

		results, total, err := repo.FindDataAdvanced(ctx, model, query)
		if err != nil {
			t.Fatalf("Failed: %v", err)
		}

		if total != 2 { // Apple and Grapes
			t.Errorf("Expected 2 results (Apple, Grapes), got %d", total)
		}
		_ = results
	})
}
