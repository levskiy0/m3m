package modules

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Note: GoalsModule tests are limited because GoalService requires MongoDB.
// These tests verify module construction and basic behavior.
// Full integration tests would require MongoDB connection.

func TestGoalsModule_NewGoalsModule(t *testing.T) {
	projectID := primitive.NewObjectID()

	// Test with nil service (should not panic during construction)
	module := NewGoalsModule(nil, projectID)

	if module == nil {
		t.Fatal("NewGoalsModule() returned nil")
	}

	if module.projectID != projectID {
		t.Error("projectID should be set correctly")
	}
}

func TestGoalsModule_NewGoalsModule_ZeroProjectID(t *testing.T) {
	var zeroID primitive.ObjectID

	module := NewGoalsModule(nil, zeroID)

	if module == nil {
		t.Fatal("NewGoalsModule() with zero ID returned nil")
	}

	if module.projectID != zeroID {
		t.Error("zero projectID should be preserved")
	}
}

func TestGoalsModule_Increment_NilService(t *testing.T) {
	projectID := primitive.NewObjectID()
	module := NewGoalsModule(nil, projectID)

	// Should return false (fail gracefully) with nil service
	result := module.Increment("test-slug")

	if result {
		t.Error("Increment with nil service should return false")
	}
}

func TestGoalsModule_Increment_WithValue(t *testing.T) {
	projectID := primitive.NewObjectID()
	module := NewGoalsModule(nil, projectID)

	// Test with custom value - should handle gracefully with nil service
	result := module.Increment("test-slug", 5)

	if result {
		t.Error("Increment with nil service should return false")
	}
}

func TestGoalsModule_Increment_ZeroValue(t *testing.T) {
	projectID := primitive.NewObjectID()
	module := NewGoalsModule(nil, projectID)

	result := module.Increment("test-slug", 0)

	if result {
		t.Error("Increment with nil service should return false")
	}
}

func TestGoalsModule_Increment_NegativeValue(t *testing.T) {
	projectID := primitive.NewObjectID()
	module := NewGoalsModule(nil, projectID)

	result := module.Increment("test-slug", -10)

	if result {
		t.Error("Increment with nil service should return false")
	}
}

func TestGoalsModule_Increment_EmptySlug(t *testing.T) {
	projectID := primitive.NewObjectID()
	module := NewGoalsModule(nil, projectID)

	result := module.Increment("")

	if result {
		t.Error("Increment with nil service should return false")
	}
}

func TestGoalsModule_Increment_SpecialCharSlug(t *testing.T) {
	projectID := primitive.NewObjectID()
	module := NewGoalsModule(nil, projectID)

	slugs := []string{
		"slug-with-dash",
		"slug_with_underscore",
		"slug.with.dot",
		"project-abc12345-goal",
		"UPPERCASE-SLUG",
	}

	for _, slug := range slugs {
		t.Run(slug, func(t *testing.T) {
			result := module.Increment(slug)

			if result {
				t.Errorf("Increment with nil service should return false for slug %q", slug)
			}
		})
	}
}

func TestGoalsModule_Increment_MultipleValues(t *testing.T) {
	projectID := primitive.NewObjectID()
	module := NewGoalsModule(nil, projectID)

	// When multiple values are passed, only the first should be used
	result := module.Increment("test-slug", 5, 10, 15)

	if result {
		t.Error("Increment with nil service should return false")
	}
}

func TestGoalsModule_Increment_LargeValue(t *testing.T) {
	projectID := primitive.NewObjectID()
	module := NewGoalsModule(nil, projectID)

	result := module.Increment("test-slug", int64(9223372036854775807)) // Max int64

	if result {
		t.Error("Increment with nil service should return false")
	}
}

func TestGoalsModule_DefaultValue(t *testing.T) {
	projectID := primitive.NewObjectID()
	module := NewGoalsModule(nil, projectID)

	// Verify that calling Increment without value doesn't panic
	result := module.Increment("test-slug")

	if result {
		t.Error("Increment with nil service should return false")
	}
}

func TestGoalsModule_ProjectIDPreserved(t *testing.T) {
	projectID := primitive.NewObjectID()
	module := NewGoalsModule(nil, projectID)

	// Verify projectID is accessible and preserved
	if module.projectID.Hex() != projectID.Hex() {
		t.Errorf("projectID = %s, want %s", module.projectID.Hex(), projectID.Hex())
	}
}

func TestGoalsModule_MultipleIncrements(t *testing.T) {
	projectID := primitive.NewObjectID()
	module := NewGoalsModule(nil, projectID)

	// Multiple increments should all fail gracefully with nil service
	for i := 0; i < 100; i++ {
		result := module.Increment("test-slug", int64(i))
		if result {
			t.Errorf("Increment #%d with nil service should return false", i)
		}
	}
}

func TestGoalsModule_DifferentProjectIDs(t *testing.T) {
	projectID1 := primitive.NewObjectID()
	projectID2 := primitive.NewObjectID()

	module1 := NewGoalsModule(nil, projectID1)
	module2 := NewGoalsModule(nil, projectID2)

	if module1.projectID == module2.projectID {
		t.Error("different modules should have different project IDs")
	}
}

// Tests for GetValue method
func TestGoalsModule_GetValue_NilService(t *testing.T) {
	projectID := primitive.NewObjectID()
	module := NewGoalsModule(nil, projectID)

	result := module.GetValue("test-slug")

	if result != 0 {
		t.Errorf("GetValue with nil service = %d, want 0", result)
	}
}

func TestGoalsModule_GetValue_EmptySlug(t *testing.T) {
	projectID := primitive.NewObjectID()
	module := NewGoalsModule(nil, projectID)

	result := module.GetValue("")

	if result != 0 {
		t.Errorf("GetValue with empty slug = %d, want 0", result)
	}
}

// Tests for GetStats method
func TestGoalsModule_GetStats_NilService(t *testing.T) {
	projectID := primitive.NewObjectID()
	module := NewGoalsModule(nil, projectID)

	result := module.GetStats("test-slug", 7)

	if result == nil {
		t.Error("GetStats should return empty slice, not nil")
	}

	if len(result) != 0 {
		t.Errorf("GetStats with nil service length = %d, want 0", len(result))
	}
}

func TestGoalsModule_GetStats_DefaultDays(t *testing.T) {
	projectID := primitive.NewObjectID()
	module := NewGoalsModule(nil, projectID)

	// Zero days should default to 7
	result := module.GetStats("test-slug", 0)

	if result == nil {
		t.Error("GetStats should return empty slice, not nil")
	}
}

func TestGoalsModule_GetStats_NegativeDays(t *testing.T) {
	projectID := primitive.NewObjectID()
	module := NewGoalsModule(nil, projectID)

	// Negative days should default to 7
	result := module.GetStats("test-slug", -5)

	if result == nil {
		t.Error("GetStats should return empty slice, not nil")
	}
}

// Tests for List method
func TestGoalsModule_List_NilService(t *testing.T) {
	projectID := primitive.NewObjectID()
	module := NewGoalsModule(nil, projectID)

	result := module.List()

	if result == nil {
		t.Error("List should return empty slice, not nil")
	}

	if len(result) != 0 {
		t.Errorf("List with nil service length = %d, want 0", len(result))
	}
}

// Tests for Get method
func TestGoalsModule_Get_NilService(t *testing.T) {
	projectID := primitive.NewObjectID()
	module := NewGoalsModule(nil, projectID)

	result := module.Get("test-slug")

	if result != nil {
		t.Error("Get with nil service should return nil")
	}
}

func TestGoalsModule_Get_EmptySlug(t *testing.T) {
	projectID := primitive.NewObjectID()
	module := NewGoalsModule(nil, projectID)

	result := module.Get("")

	if result != nil {
		t.Error("Get with empty slug should return nil")
	}
}

// Note: To properly test GoalsModule with full functionality,
// integration tests with MongoDB or mock implementation of GoalService are needed.
// The following tests would be added with a proper test infrastructure:
//
// func TestGoalsModule_Increment_Success(t *testing.T) { ... }
// func TestGoalsModule_Increment_GoalNotFound(t *testing.T) { ... }
// func TestGoalsModule_Increment_NoAccessToGoal(t *testing.T) { ... }
// func TestGoalsModule_Increment_DailyCounter(t *testing.T) { ... }
// func TestGoalsModule_Increment_TotalCounter(t *testing.T) { ... }
// func TestGoalsModule_GetValue_Success(t *testing.T) { ... }
// func TestGoalsModule_GetStats_Success(t *testing.T) { ... }
// func TestGoalsModule_List_Success(t *testing.T) { ... }
// func TestGoalsModule_Get_Success(t *testing.T) { ... }
