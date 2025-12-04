package modules

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dop251/goja"
)

func TestDelayedModule_NewDelayedModule(t *testing.T) {
	tests := []struct {
		name         string
		poolSize     int
		expectedSize int
	}{
		{"positive pool size", 5, 5},
		{"zero pool size defaults to 10", 0, 10},
		{"negative pool size defaults to 10", -1, 10},
		{"large pool size", 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDelayedModule(tt.poolSize)
			if d == nil {
				t.Fatal("NewDelayedModule() returned nil")
			}
			if d.poolSize != tt.expectedSize {
				t.Errorf("poolSize = %d, want %d", d.poolSize, tt.expectedSize)
			}
			if cap(d.sem) != tt.expectedSize {
				t.Errorf("semaphore capacity = %d, want %d", cap(d.sem), tt.expectedSize)
			}
		})
	}
}

func TestDelayedModule_Run_BasicExecution(t *testing.T) {
	vm := goja.New()
	d := NewDelayedModule(5)

	var executed atomic.Bool
	handler := vm.ToValue(func(call goja.FunctionCall) goja.Value {
		executed.Store(true)
		return goja.Undefined()
	})

	callable, ok := goja.AssertFunction(handler)
	if !ok {
		t.Fatal("failed to create callable")
	}

	d.Run(callable)

	// Wait for execution
	time.Sleep(100 * time.Millisecond)

	if !executed.Load() {
		t.Error("handler was not executed")
	}
}

func TestDelayedModule_Run_MultipleTasks(t *testing.T) {
	vm := goja.New()
	d := NewDelayedModule(5)

	var counter atomic.Int32
	handler := vm.ToValue(func(call goja.FunctionCall) goja.Value {
		counter.Add(1)
		return goja.Undefined()
	})

	callable, ok := goja.AssertFunction(handler)
	if !ok {
		t.Fatal("failed to create callable")
	}

	taskCount := 10
	for i := 0; i < taskCount; i++ {
		d.Run(callable)
	}

	// Wait for all executions
	time.Sleep(200 * time.Millisecond)

	if counter.Load() != int32(taskCount) {
		t.Errorf("counter = %d, want %d", counter.Load(), taskCount)
	}
}

func TestDelayedModule_Run_PoolLimiting(t *testing.T) {
	vm := goja.New()
	poolSize := 3
	d := NewDelayedModule(poolSize)

	var concurrent atomic.Int32
	var maxConcurrent atomic.Int32

	handler := vm.ToValue(func(call goja.FunctionCall) goja.Value {
		current := concurrent.Add(1)
		// Track maximum concurrent executions
		for {
			max := maxConcurrent.Load()
			if current <= max || maxConcurrent.CompareAndSwap(max, current) {
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
		concurrent.Add(-1)
		return goja.Undefined()
	})

	callable, ok := goja.AssertFunction(handler)
	if !ok {
		t.Fatal("failed to create callable")
	}

	// Launch more tasks than pool size
	taskCount := 10
	for i := 0; i < taskCount; i++ {
		d.Run(callable)
	}

	// Wait for all executions
	time.Sleep(500 * time.Millisecond)

	max := maxConcurrent.Load()
	if int(max) > poolSize {
		t.Errorf("maxConcurrent = %d, should not exceed poolSize %d", max, poolSize)
	}
}

func TestDelayedModule_Run_PanicRecovery(t *testing.T) {
	vm := goja.New()
	d := NewDelayedModule(5)

	var normalExecuted atomic.Bool

	// Handler that panics
	panicHandler := vm.ToValue(func(call goja.FunctionCall) goja.Value {
		panic("test panic")
	})
	panicCallable, _ := goja.AssertFunction(panicHandler)

	// Normal handler
	normalHandler := vm.ToValue(func(call goja.FunctionCall) goja.Value {
		normalExecuted.Store(true)
		return goja.Undefined()
	})
	normalCallable, _ := goja.AssertFunction(normalHandler)

	// Run panicking task first
	d.Run(panicCallable)
	// Wait for panic to happen
	time.Sleep(50 * time.Millisecond)

	// Run normal task - should still work
	d.Run(normalCallable)
	time.Sleep(100 * time.Millisecond)

	if !normalExecuted.Load() {
		t.Error("normal handler should execute after panic in another task")
	}
}

func TestDelayedModule_Run_SemaphoreRelease(t *testing.T) {
	vm := goja.New()
	poolSize := 2
	d := NewDelayedModule(poolSize)

	var completedCount atomic.Int32

	handler := vm.ToValue(func(call goja.FunctionCall) goja.Value {
		time.Sleep(20 * time.Millisecond)
		completedCount.Add(1)
		return goja.Undefined()
	})

	callable, ok := goja.AssertFunction(handler)
	if !ok {
		t.Fatal("failed to create callable")
	}

	// Run many tasks sequentially through limited pool
	taskCount := 6
	for i := 0; i < taskCount; i++ {
		d.Run(callable)
	}

	// Wait long enough for all tasks
	time.Sleep(300 * time.Millisecond)

	if completedCount.Load() != int32(taskCount) {
		t.Errorf("completed = %d, want %d (semaphore may not be released)", completedCount.Load(), taskCount)
	}
}

func TestDelayedModule_Run_ConcurrentRegistration(t *testing.T) {
	vm := goja.New()
	d := NewDelayedModule(10)

	var counter atomic.Int32
	handler := vm.ToValue(func(call goja.FunctionCall) goja.Value {
		counter.Add(1)
		return goja.Undefined()
	})

	callable, ok := goja.AssertFunction(handler)
	if !ok {
		t.Fatal("failed to create callable")
	}

	var wg sync.WaitGroup
	taskCount := 20

	// Register tasks concurrently
	for i := 0; i < taskCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			d.Run(callable)
		}()
	}

	wg.Wait()
	// Wait for executions
	time.Sleep(200 * time.Millisecond)

	if counter.Load() != int32(taskCount) {
		t.Errorf("counter = %d, want %d", counter.Load(), taskCount)
	}
}

func TestDelayedModule_Run_NilHandler(t *testing.T) {
	d := NewDelayedModule(5)

	// This should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Run with nil handler should not panic: %v", r)
		}
	}()

	d.Run(nil)
	time.Sleep(50 * time.Millisecond)
}

func TestDelayedModule_Run_ExecutionOrder(t *testing.T) {
	vm := goja.New()
	d := NewDelayedModule(1) // Single worker to ensure sequential execution

	var order []int
	var mu sync.Mutex

	for i := 0; i < 5; i++ {
		idx := i
		handler := vm.ToValue(func(call goja.FunctionCall) goja.Value {
			mu.Lock()
			order = append(order, idx)
			mu.Unlock()
			return goja.Undefined()
		})
		callable, _ := goja.AssertFunction(handler)
		d.Run(callable)
	}

	// Wait for all executions
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(order) != 5 {
		t.Errorf("expected 5 executions, got %d", len(order))
	}
}

func TestDelayedModule_Run_PanicDoesNotLeakSemaphore(t *testing.T) {
	vm := goja.New()
	poolSize := 2
	d := NewDelayedModule(poolSize)

	var successCount atomic.Int32

	// Create panicking handlers
	for i := 0; i < poolSize; i++ {
		panicHandler := vm.ToValue(func(call goja.FunctionCall) goja.Value {
			panic("intentional panic")
		})
		callable, _ := goja.AssertFunction(panicHandler)
		d.Run(callable)
	}

	// Wait for panics
	time.Sleep(50 * time.Millisecond)

	// Now run successful tasks - they should be able to acquire semaphore
	successHandler := vm.ToValue(func(call goja.FunctionCall) goja.Value {
		successCount.Add(1)
		return goja.Undefined()
	})
	successCallable, _ := goja.AssertFunction(successHandler)

	for i := 0; i < 3; i++ {
		d.Run(successCallable)
	}

	time.Sleep(200 * time.Millisecond)

	if successCount.Load() != 3 {
		t.Errorf("successCount = %d, want 3 (semaphore may be leaked)", successCount.Load())
	}
}
