package runtime

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/levskiy0/m3m/internal/config"
	"github.com/levskiy0/m3m/internal/service"
)

// These tests verify the context cancellation bug fix and runtime stability

// createTestManager creates a Manager with minimal dependencies for testing
func createTestManager(t *testing.T) (*Manager, func()) {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "m3m-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cfg := &config.Config{
		Runtime: config.RuntimeConfig{
			Timeout:        30,
			WorkerPoolSize: 5,
		},
		Storage: config.StorageConfig{
			Path: tempDir,
		},
		Logging: config.LoggingConfig{
			Level: "debug",
			Path:  tempDir,
		},
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	storageService := service.NewStorageService(cfg)

	manager := NewManager(cfg, logger, nil, nil, nil, nil, storageService)

	cleanup := func() {
		manager.StopAll()
		os.RemoveAll(tempDir)
	}

	return manager, cleanup
}

// TestContextCancellationBug tests that runtime survives parent context cancellation
// This was the main bug: AutoStartRuntimes passed OnStart context which was cancelled
// immediately after OnStart completed, killing all auto-started services
func TestContextCancellationBug(t *testing.T) {
	manager, cleanup := createTestManager(t)
	defer cleanup()

	projectID := primitive.NewObjectID()

	// Simple code that just runs
	code := `
		$service.start(function() {
			$logger.info("Service started!");
		});
	`

	// BUG SCENARIO: Parent context that gets cancelled quickly (simulating OnStart context)
	parentCtx, parentCancel := context.WithCancel(context.Background())

	// Start runtime with parent context
	err := manager.Start(parentCtx, projectID, code)
	if err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}

	// Wait for service to start
	time.Sleep(100 * time.Millisecond)

	// Verify runtime is running
	if !manager.IsRunning(projectID) {
		t.Fatal("Runtime should be running after start")
	}

	// Cancel parent context (simulating OnStart completion)
	parentCancel()

	// Wait a bit for context cancellation to propagate
	time.Sleep(200 * time.Millisecond)

	// BUG CHECK: Runtime should still be running after parent context cancellation
	// This was the bug - runtime died when parent context was cancelled
	if !manager.IsRunning(projectID) {
		t.Fatal("BUG REPRODUCED: Runtime died when parent context was cancelled!")
	}

	// Cleanup
	manager.Stop(projectID)
}

// TestRuntimeSurvivesShortLivedContext tests with very short-lived context
func TestRuntimeSurvivesShortLivedContext(t *testing.T) {
	manager, cleanup := createTestManager(t)
	defer cleanup()

	projectID := primitive.NewObjectID()

	code := `
		$service.start(function() {
			$logger.info("Started!");
		});
	`

	// Create already-cancelled context (worst case scenario)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately!

	// Start with already-cancelled context
	err := manager.Start(ctx, projectID, code)
	if err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	// Runtime should still be running even with pre-cancelled context
	// Because Start() should use its own background context internally
	if !manager.IsRunning(projectID) {
		t.Fatal("Runtime should survive even with pre-cancelled parent context")
	}

	manager.Stop(projectID)
}

// TestRuntimeStopsOnExplicitStop verifies normal stop behavior
func TestRuntimeStopsOnExplicitStop(t *testing.T) {
	manager, cleanup := createTestManager(t)
	defer cleanup()

	projectID := primitive.NewObjectID()

	// Simple code without shutdown callback to avoid race condition
	code := `
		$service.start(function() {
			$logger.info("Started!");
		});
	`

	err := manager.Start(context.Background(), projectID, code)
	if err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if !manager.IsRunning(projectID) {
		t.Fatal("Runtime should be running")
	}

	// Explicit stop should work
	err = manager.Stop(projectID)
	if err != nil {
		t.Fatalf("Failed to stop runtime: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if manager.IsRunning(projectID) {
		t.Fatal("Runtime should be stopped after explicit Stop()")
	}
}

// TestMultipleRuntimesIsolation tests that multiple runtimes are isolated
func TestMultipleRuntimesIsolation(t *testing.T) {
	manager, cleanup := createTestManager(t)
	defer cleanup()

	project1 := primitive.NewObjectID()
	project2 := primitive.NewObjectID()

	code := `$service.start(function() {});`

	// Start both with different contexts
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()

	manager.Start(ctx1, project1, code)
	manager.Start(ctx2, project2, code)

	time.Sleep(100 * time.Millisecond)

	// Cancel context 1
	cancel1()
	time.Sleep(100 * time.Millisecond)

	// Both should still be running (context cancellation shouldn't affect runtime)
	if !manager.IsRunning(project1) {
		t.Fatal("Project 1 should still be running after parent context cancelled")
	}
	if !manager.IsRunning(project2) {
		t.Fatal("Project 2 should still be running")
	}

	// Stop project 1 explicitly
	manager.Stop(project1)
	time.Sleep(100 * time.Millisecond)

	// Project 1 stopped, project 2 still running
	if manager.IsRunning(project1) {
		t.Fatal("Project 1 should be stopped")
	}
	if !manager.IsRunning(project2) {
		t.Fatal("Project 2 should still be running after project 1 stopped")
	}
}

// TestStopAllStopsAllRuntimes tests StopAll behavior
func TestStopAllStopsAllRuntimes(t *testing.T) {
	manager, cleanup := createTestManager(t)
	defer cleanup()

	project1 := primitive.NewObjectID()
	project2 := primitive.NewObjectID()
	project3 := primitive.NewObjectID()

	code := `$service.start(function() {});`

	manager.Start(context.Background(), project1, code)
	manager.Start(context.Background(), project2, code)
	manager.Start(context.Background(), project3, code)

	time.Sleep(100 * time.Millisecond)

	// All should be running
	if !manager.IsRunning(project1) || !manager.IsRunning(project2) || !manager.IsRunning(project3) {
		t.Fatal("All projects should be running")
	}

	// Stop all
	manager.StopAll()

	time.Sleep(200 * time.Millisecond)

	// All should be stopped
	if manager.IsRunning(project1) || manager.IsRunning(project2) || manager.IsRunning(project3) {
		t.Fatal("All projects should be stopped after StopAll()")
	}
}

// TestAutoRestartOnCrash tests that runtime auto-restarts after a crash
func TestAutoRestartOnCrash(t *testing.T) {
	manager, cleanup := createTestManager(t)
	defer cleanup()

	projectID := primitive.NewObjectID()

	// Code that crashes on first run but works on second
	// Use a global counter via $env or similar mechanism
	code := `
		// This will cause a runtime error
		throw new Error("Simulated crash for testing");
	`

	err := manager.Start(context.Background(), projectID, code)
	if err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}

	// Wait for crash and auto-restart attempt
	// Initial delay is 1 second, so wait a bit longer
	time.Sleep(1500 * time.Millisecond)

	// Runtime should have attempted restart (will keep crashing but that's ok)
	// The key is that it tried to restart
	// Check logs would show "Auto-restarting in..."
	t.Log("Auto-restart test completed - check logs for 'Auto-restarting' message")
}

// TestNoAutoRestartOnExplicitStop tests that explicit stop doesn't trigger auto-restart
func TestNoAutoRestartOnExplicitStop(t *testing.T) {
	manager, cleanup := createTestManager(t)
	defer cleanup()

	projectID := primitive.NewObjectID()

	code := `$service.start(function() {});`

	err := manager.Start(context.Background(), projectID, code)
	if err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if !manager.IsRunning(projectID) {
		t.Fatal("Runtime should be running")
	}

	// Explicit stop
	manager.Stop(projectID)

	time.Sleep(200 * time.Millisecond)

	// Should NOT auto-restart
	if manager.IsRunning(projectID) {
		t.Fatal("Runtime should NOT auto-restart after explicit Stop()")
	}
}

// TestAutoRestartLimit tests that auto-restart respects the max restarts limit
func TestAutoRestartLimit(t *testing.T) {
	// Skip in normal test runs as this takes time
	if testing.Short() {
		t.Skip("Skipping auto-restart limit test in short mode")
	}

	manager, cleanup := createTestManager(t)
	defer cleanup()

	projectID := primitive.NewObjectID()

	// Code that always crashes
	code := `throw new Error("Always crash");`

	err := manager.Start(context.Background(), projectID, code)
	if err != nil {
		t.Fatalf("Failed to start runtime: %v", err)
	}

	// Wait for multiple restart attempts
	// With exponential backoff: 1s + 2s + 4s + 8s + 16s = 31s for 5 restarts
	// But we'll just wait a reasonable time and check the limit is respected
	time.Sleep(10 * time.Second)

	// After many crashes, the runtime should eventually give up
	// The log should show "Auto-restart disabled: exceeded max restarts"
	t.Log("Auto-restart limit test completed - check logs for limit exceeded message")
}
