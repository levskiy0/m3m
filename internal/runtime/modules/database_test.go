package modules

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Note: Full database module tests require MongoDB integration.
// These tests verify basic structure and behavior with nil/zero states.

func TestDatabaseModule_NewDatabaseModule(t *testing.T) {
	projectID := primitive.NewObjectID()
	// nil modelService is acceptable for structure test
	db := NewDatabaseModule(nil, projectID)

	if db == nil {
		t.Fatal("NewDatabaseModule() returned nil")
	}

	if db.projectID != projectID {
		t.Error("projectID not set correctly")
	}
}

func TestDatabaseModule_Collection_NoService(t *testing.T) {
	// This test requires a valid modelService which requires MongoDB
	// Collection() will panic with nil modelService, so we skip
	t.Skip("Collection() requires valid modelService - skipping")
}

func TestCollectionWrapper_EmptyModelID(t *testing.T) {
	wrapper := &CollectionWrapper{
		modelService: nil,
		projectID:    primitive.NewObjectID(),
		modelSlug:    "test",
		modelID:      primitive.NilObjectID, // Zero/empty model ID
	}

	// Find should return empty array
	results := wrapper.Find(nil)
	if len(results) != 0 {
		t.Errorf("Find() with zero modelID should return empty array, got %d items", len(results))
	}

	// FindOne should return nil
	result := wrapper.FindOne(nil)
	if result != nil {
		t.Errorf("FindOne() with zero modelID should return nil, got %v", result)
	}

	// Insert should return nil
	insertResult := wrapper.Insert(map[string]interface{}{"key": "value"})
	if insertResult != nil {
		t.Errorf("Insert() with zero modelID should return nil, got %v", insertResult)
	}

	// Update should return false
	if wrapper.Update("507f1f77bcf86cd799439011", map[string]interface{}{"key": "value"}) {
		t.Error("Update() with zero modelID should return false")
	}

	// Delete should return false
	if wrapper.Delete("507f1f77bcf86cd799439011") {
		t.Error("Delete() with zero modelID should return false")
	}

	// Count should return 0
	if wrapper.Count(nil) != 0 {
		t.Error("Count() with zero modelID should return 0")
	}
}

func TestCollectionWrapper_InvalidObjectID(t *testing.T) {
	wrapper := &CollectionWrapper{
		modelService: nil,
		projectID:    primitive.NewObjectID(),
		modelSlug:    "test",
		modelID:      primitive.NewObjectID(),
	}

	// Update with invalid ID should return false
	if wrapper.Update("invalid-id", map[string]interface{}{"key": "value"}) {
		t.Error("Update() with invalid ObjectID should return false")
	}

	// Delete with invalid ID should return false
	if wrapper.Delete("invalid-id") {
		t.Error("Delete() with invalid ObjectID should return false")
	}
}
