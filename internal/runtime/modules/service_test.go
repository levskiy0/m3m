package modules

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dop251/goja"
)

func TestServiceModule_NewServiceModule(t *testing.T) {
	vm := goja.New()
	service := NewServiceModule(vm, 30*time.Second)

	if service == nil {
		t.Fatal("NewServiceModule() returned nil")
	}

	if service.booted {
		t.Error("service should not be booted initially")
	}

	if service.started {
		t.Error("service should not be started initially")
	}
}

func TestServiceModule_Boot(t *testing.T) {
	vm := goja.New()
	service := NewServiceModule(vm, 30*time.Second)

	var executed bool
	handler := vm.ToValue(func(call goja.FunctionCall) goja.Value {
		executed = true
		return goja.Undefined()
	})

	callable, _ := goja.AssertFunction(handler)
	service.Boot(callable)

	err := service.ExecuteBoot()
	if err != nil {
		t.Errorf("ExecuteBoot() error: %v", err)
	}

	if !executed {
		t.Error("boot callback was not executed")
	}

	if !service.IsBooted() {
		t.Error("IsBooted() should return true after ExecuteBoot()")
	}
}

func TestServiceModule_Start(t *testing.T) {
	vm := goja.New()
	service := NewServiceModule(vm, 30*time.Second)

	var executed bool
	handler := vm.ToValue(func(call goja.FunctionCall) goja.Value {
		executed = true
		return goja.Undefined()
	})

	callable, _ := goja.AssertFunction(handler)
	service.Start(callable)

	err := service.ExecuteStart()
	if err != nil {
		t.Errorf("ExecuteStart() error: %v", err)
	}

	if !executed {
		t.Error("start callback was not executed")
	}

	if !service.IsStarted() {
		t.Error("IsStarted() should return true after ExecuteStart()")
	}
}

func TestServiceModule_Shutdown(t *testing.T) {
	vm := goja.New()
	service := NewServiceModule(vm, 30*time.Second)

	var executed bool
	handler := vm.ToValue(func(call goja.FunctionCall) goja.Value {
		executed = true
		return goja.Undefined()
	})

	callable, _ := goja.AssertFunction(handler)
	service.Shutdown(callable)

	err := service.ExecuteShutdown()
	if err != nil {
		t.Errorf("ExecuteShutdown() error: %v", err)
	}

	if !executed {
		t.Error("shutdown callback was not executed")
	}
}

func TestServiceModule_MultipleCallbacks(t *testing.T) {
	vm := goja.New()
	service := NewServiceModule(vm, 30*time.Second)

	var order []int
	var mu sync.Mutex

	addCallback := func(n int) goja.Callable {
		handler := vm.ToValue(func(call goja.FunctionCall) goja.Value {
			mu.Lock()
			order = append(order, n)
			mu.Unlock()
			return goja.Undefined()
		})
		callable, _ := goja.AssertFunction(handler)
		return callable
	}

	service.Boot(addCallback(1))
	service.Boot(addCallback(2))
	service.Boot(addCallback(3))

	err := service.ExecuteBoot()
	if err != nil {
		t.Errorf("ExecuteBoot() error: %v", err)
	}

	if len(order) != 3 {
		t.Errorf("expected 3 callbacks, got %d", len(order))
	}

	// Check callbacks were executed in order
	for i, n := range order {
		if n != i+1 {
			t.Errorf("callback order[%d] = %d, want %d", i, n, i+1)
		}
	}
}

func TestServiceModule_LifecycleOrder(t *testing.T) {
	vm := goja.New()
	service := NewServiceModule(vm, 30*time.Second)

	var order []string
	var mu sync.Mutex

	createCallback := func(phase string) goja.Callable {
		handler := vm.ToValue(func(call goja.FunctionCall) goja.Value {
			mu.Lock()
			order = append(order, phase)
			mu.Unlock()
			return goja.Undefined()
		})
		callable, _ := goja.AssertFunction(handler)
		return callable
	}

	service.Boot(createCallback("boot"))
	service.Start(createCallback("start"))
	service.Shutdown(createCallback("shutdown"))

	// Execute in order
	service.ExecuteBoot()
	service.ExecuteStart()
	service.ExecuteShutdown()

	expected := []string{"boot", "start", "shutdown"}
	if len(order) != len(expected) {
		t.Errorf("expected %d phases, got %d", len(expected), len(order))
		return
	}

	for i, phase := range order {
		if phase != expected[i] {
			t.Errorf("phase[%d] = %q, want %q", i, phase, expected[i])
		}
	}
}

func TestServiceModule_CallbackError(t *testing.T) {
	vm := goja.New()
	service := NewServiceModule(vm, 30*time.Second)

	// Register JS code that throws an error
	_, err := vm.RunString(`
		var errorHandler = function() {
			throw new Error("callback error");
		};
	`)
	if err != nil {
		t.Fatalf("Failed to setup error handler: %v", err)
	}

	val := vm.Get("errorHandler")
	callable, _ := goja.AssertFunction(val)
	service.Boot(callable)

	err = service.ExecuteBoot()
	if err == nil {
		t.Error("ExecuteBoot() should return error for throwing callback")
	}
}

func TestServiceModule_ShutdownTimeout(t *testing.T) {
	vm := goja.New()
	service := NewServiceModule(vm, 100*time.Millisecond)

	var completed atomic.Bool
	handler := vm.ToValue(func(call goja.FunctionCall) goja.Value {
		time.Sleep(500 * time.Millisecond)
		completed.Store(true)
		return goja.Undefined()
	})

	callable, _ := goja.AssertFunction(handler)
	service.Shutdown(callable)

	start := time.Now()
	err := service.ExecuteShutdown()
	elapsed := time.Since(start)

	// Should timeout before callback completes
	if elapsed > 200*time.Millisecond {
		t.Errorf("shutdown should timeout, took %v", elapsed)
	}

	// Error should be nil (timeout is silent)
	if err != nil {
		t.Errorf("ExecuteShutdown() should not return error on timeout, got: %v", err)
	}
}

func TestServiceModule_GetJSObject(t *testing.T) {
	vm := goja.New()
	service := NewServiceModule(vm, 30*time.Second)

	obj := service.GetJSObject()

	if obj["boot"] == nil {
		t.Error("GetJSObject() should have 'boot' function")
	}
	if obj["start"] == nil {
		t.Error("GetJSObject() should have 'start' function")
	}
	if obj["shutdown"] == nil {
		t.Error("GetJSObject() should have 'shutdown' function")
	}
}

func TestServiceModule_ConcurrentCallbackRegistration(t *testing.T) {
	vm := goja.New()
	service := NewServiceModule(vm, 30*time.Second)

	var wg sync.WaitGroup
	handler := vm.ToValue(func(call goja.FunctionCall) goja.Value {
		return goja.Undefined()
	})
	callable, _ := goja.AssertFunction(handler)

	// Register callbacks concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			service.Boot(callable)
		}()
	}
	wg.Wait()

	// Should have 10 boot callbacks
	if len(service.bootCallbacks) != 10 {
		t.Errorf("expected 10 boot callbacks, got %d", len(service.bootCallbacks))
	}
}

func TestServiceModule_EmptyCallbacks(t *testing.T) {
	vm := goja.New()
	service := NewServiceModule(vm, 30*time.Second)

	// Execute without registering any callbacks
	if err := service.ExecuteBoot(); err != nil {
		t.Errorf("ExecuteBoot() with no callbacks should not error: %v", err)
	}

	if err := service.ExecuteStart(); err != nil {
		t.Errorf("ExecuteStart() with no callbacks should not error: %v", err)
	}

	if err := service.ExecuteShutdown(); err != nil {
		t.Errorf("ExecuteShutdown() with no callbacks should not error: %v", err)
	}

	// Status should still be updated
	if !service.IsBooted() {
		t.Error("IsBooted() should be true even with no callbacks")
	}
	if !service.IsStarted() {
		t.Error("IsStarted() should be true even with no callbacks")
	}
}

func TestServiceModule_DefaultTimeout(t *testing.T) {
	vm := goja.New()
	service := NewServiceModule(vm, 0) // 0 timeout

	// Default should be 30 seconds
	if service.shutdownTimeout != 30*time.Second {
		t.Errorf("default shutdownTimeout = %v, want 30s", service.shutdownTimeout)
	}
}
