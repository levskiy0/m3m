package modules

import (
	"log/slog"
	"os"
	"testing"

	"github.com/dop251/goja"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestScheduleModule_NewScheduleModule(t *testing.T) {
	logger := newTestLogger()
	sched := NewScheduleModule(logger)

	if sched == nil {
		t.Fatal("NewScheduleModule() returned nil")
	}

	if sched.cron == nil {
		t.Error("cron field is nil")
	}

	if sched.started {
		t.Error("scheduler should not be started initially")
	}
}

func TestScheduleModule_Cron(t *testing.T) {
	logger := newTestLogger()
	sched := NewScheduleModule(logger)
	vm := goja.New()

	// Create a callable
	handler := vm.ToValue(func(call goja.FunctionCall) goja.Value {
		return goja.Undefined()
	})

	callable, ok := goja.AssertFunction(handler)
	if !ok {
		t.Fatal("Failed to create callable")
	}

	// Add cron job that runs every minute (standard 5-field cron)
	// For testing, we just verify the job is added, not executed
	sched.Cron("* * * * *", callable)

	// Check job was added
	if len(sched.jobs) != 1 {
		t.Errorf("Cron() should add 1 job, got %d", len(sched.jobs))
	}

	// Start scheduler
	sched.Start()
	if !sched.started {
		t.Error("scheduler should be started")
	}

	// Stop scheduler
	sched.Stop()
}

func TestScheduleModule_Daily(t *testing.T) {
	logger := newTestLogger()
	sched := NewScheduleModule(logger)
	vm := goja.New()

	handler := vm.ToValue(func(call goja.FunctionCall) goja.Value {
		return goja.Undefined()
	})

	callable, _ := goja.AssertFunction(handler)
	sched.Daily(callable)

	// Check job was added
	if len(sched.jobs) != 1 {
		t.Errorf("Daily() should add 1 job, got %d", len(sched.jobs))
	}
}

func TestScheduleModule_Hourly(t *testing.T) {
	logger := newTestLogger()
	sched := NewScheduleModule(logger)
	vm := goja.New()

	handler := vm.ToValue(func(call goja.FunctionCall) goja.Value {
		return goja.Undefined()
	})

	callable, _ := goja.AssertFunction(handler)
	sched.Hourly(callable)

	// Check job was added
	if len(sched.jobs) != 1 {
		t.Errorf("Hourly() should add 1 job, got %d", len(sched.jobs))
	}
}

func TestScheduleModule_Start_Stop(t *testing.T) {
	logger := newTestLogger()
	sched := NewScheduleModule(logger)
	vm := goja.New()

	handler := vm.ToValue(func(call goja.FunctionCall) goja.Value {
		return goja.Undefined()
	})

	callable, _ := goja.AssertFunction(handler)
	sched.Cron("0 * * * *", callable)

	// Start
	sched.Start()
	if !sched.started {
		t.Error("scheduler should be started after Start()")
	}

	// Stop
	sched.Stop()
	if sched.started {
		t.Error("scheduler should be stopped after Stop()")
	}
}

func TestScheduleModule_Start_NoJobs(t *testing.T) {
	logger := newTestLogger()
	sched := NewScheduleModule(logger)

	// Start with no jobs
	sched.Start()

	// Should not be started without jobs
	if sched.started {
		t.Error("scheduler should not start without jobs")
	}
}

func TestScheduleModule_MultipleJobs(t *testing.T) {
	logger := newTestLogger()
	sched := NewScheduleModule(logger)
	vm := goja.New()

	handler := vm.ToValue(func(call goja.FunctionCall) goja.Value {
		return goja.Undefined()
	})

	callable, _ := goja.AssertFunction(handler)

	sched.Daily(callable)
	sched.Hourly(callable)
	sched.Cron("0 0 * * *", callable)

	if len(sched.jobs) != 3 {
		t.Errorf("Expected 3 jobs, got %d", len(sched.jobs))
	}
}

func TestScheduleModule_InvalidCronExpression(t *testing.T) {
	logger := newTestLogger()
	sched := NewScheduleModule(logger)
	vm := goja.New()

	handler := vm.ToValue(func(call goja.FunctionCall) goja.Value {
		return goja.Undefined()
	})

	callable, _ := goja.AssertFunction(handler)

	// Add invalid cron expression - should not panic
	sched.Cron("invalid cron", callable)

	// Job should not be added
	if len(sched.jobs) != 0 {
		t.Errorf("Invalid cron expression should not add job, got %d", len(sched.jobs))
	}
}

func TestScheduleModule_HandlerError(t *testing.T) {
	logger := newTestLogger()
	sched := NewScheduleModule(logger)
	vm := goja.New()

	// Create handler that throws error
	handler := vm.ToValue(func(call goja.FunctionCall) goja.Value {
		panic("test error")
	})

	callable, _ := goja.AssertFunction(handler)

	// Add job with error-throwing handler (use valid 5-field cron)
	sched.Cron("* * * * *", callable)

	if len(sched.jobs) != 1 {
		t.Error("Job should be added even with potentially error-throwing handler")
	}

	// Start and immediately stop - we just verify it doesn't panic during setup
	sched.Start()
	sched.Stop()
}

func TestScheduleModule_DoubleStart(t *testing.T) {
	logger := newTestLogger()
	sched := NewScheduleModule(logger)
	vm := goja.New()

	handler := vm.ToValue(func(call goja.FunctionCall) goja.Value {
		return goja.Undefined()
	})

	callable, _ := goja.AssertFunction(handler)
	sched.Cron("0 * * * *", callable)

	sched.Start()
	sched.Start() // Should not panic or cause issues

	if !sched.started {
		t.Error("scheduler should still be started")
	}

	sched.Stop()
}

func TestScheduleModule_DoubleStop(t *testing.T) {
	logger := newTestLogger()
	sched := NewScheduleModule(logger)
	vm := goja.New()

	handler := vm.ToValue(func(call goja.FunctionCall) goja.Value {
		return goja.Undefined()
	})

	callable, _ := goja.AssertFunction(handler)
	sched.Cron("0 * * * *", callable)

	sched.Start()
	sched.Stop()
	sched.Stop() // Should not panic or cause issues

	if sched.started {
		t.Error("scheduler should still be stopped")
	}
}
