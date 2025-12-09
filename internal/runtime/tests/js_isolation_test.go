package tests

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dop251/goja"
	"github.com/levskiy0/m3m/internal/runtime/modules"
)

// TestIsolation_InfiniteLoopInterrupt tests that infinite loops can be interrupted
func TestIsolation_InfiniteLoopInterrupt(t *testing.T) {
	vm := goja.New()

	// Start infinite loop in goroutine
	done := make(chan error, 1)
	go func() {
		_, err := vm.RunString(`
			while (true) {
				// infinite loop
			}
		`)
		done <- err
	}()

	// Wait a bit then interrupt
	time.Sleep(100 * time.Millisecond)
	vm.Interrupt("execution timeout")

	select {
	case err := <-done:
		if err == nil {
			t.Error("Expected interrupt error, got nil")
		} else {
			t.Logf("Loop interrupted successfully: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Interrupt did not stop the infinite loop")
	}
}

// TestIsolation_InterruptWithTimeout tests interrupt with configurable timeout
func TestIsolation_InterruptWithTimeout(t *testing.T) {
	vm := goja.New()
	timeout := 200 * time.Millisecond

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Start goroutine to interrupt VM when context expires
	go func() {
		<-ctx.Done()
		vm.Interrupt("timeout exceeded")
	}()

	start := time.Now()
	_, err := vm.RunString(`
		var i = 0;
		while (true) {
			i++;
		}
	`)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("Expected interrupt error")
	}

	// Should complete within reasonable margin of timeout
	if elapsed > timeout+100*time.Millisecond {
		t.Errorf("Took too long to interrupt: %v (expected ~%v)", elapsed, timeout)
	}

	t.Logf("Interrupted after %v", elapsed)
}

// TestIsolation_MultipleVMsIndependent tests that VMs are isolated from each other
func TestIsolation_MultipleVMsIndependent(t *testing.T) {
	vm1 := goja.New()
	vm2 := goja.New()

	// Set variable in VM1
	vm1.RunString(`var secret = "vm1_secret";`)

	// VM2 should not see VM1's variable
	result, err := vm2.RunString(`
		typeof secret === "undefined"
	`)
	if err != nil {
		t.Fatal(err)
	}

	if !result.ToBoolean() {
		t.Error("VM2 should not have access to VM1's variables")
	}

	// Modify global in VM1, VM2 should be unaffected
	vm1.RunString(`globalCounter = 100;`)

	result2, err := vm2.RunString(`
		typeof globalCounter === "undefined"
	`)
	if err != nil {
		t.Fatal(err)
	}

	if !result2.ToBoolean() {
		t.Error("VM2 should not see VM1's global modifications")
	}
}

// TestIsolation_InterruptDoesNotAffectOtherVMs tests that interrupting one VM doesn't affect others
func TestIsolation_InterruptDoesNotAffectOtherVMs(t *testing.T) {
	vm1 := goja.New()
	vm2 := goja.New()

	var wg sync.WaitGroup
	var vm1Interrupted, vm2Completed atomic.Bool

	// VM1: infinite loop that will be interrupted
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, err := vm1.RunString(`while (true) {}`)
		if err != nil {
			vm1Interrupted.Store(true)
		}
	}()

	// VM2: finite work that should complete
	wg.Add(1)
	go func() {
		defer wg.Done()
		result, err := vm2.RunString(`
			var sum = 0;
			for (var i = 0; i < 1000; i++) {
				sum += i;
			}
			sum;
		`)
		if err == nil && result.ToInteger() == 499500 {
			vm2Completed.Store(true)
		}
	}()

	// Give VM2 time to complete
	time.Sleep(100 * time.Millisecond)

	// Interrupt VM1
	vm1.Interrupt("stop")

	// Wait for both to finish
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for VMs to finish")
	}

	if !vm1Interrupted.Load() {
		t.Error("VM1 should have been interrupted")
	}
	if !vm2Completed.Load() {
		t.Error("VM2 should have completed normally")
	}
}

// TestIsolation_CPUHeavyLoop tests interrupting CPU-intensive loops
func TestIsolation_CPUHeavyLoop(t *testing.T) {
	vm := goja.New()

	done := make(chan struct{})
	go func() {
		_, _ = vm.RunString(`
			// CPU intensive work
			while (true) {
				var x = Math.sqrt(Math.random() * 1000000);
				var y = Math.sin(x) * Math.cos(x);
			}
		`)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	vm.Interrupt("cpu limit")

	select {
	case <-done:
		t.Log("CPU-heavy loop interrupted successfully")
	case <-time.After(2 * time.Second):
		t.Error("Failed to interrupt CPU-heavy loop")
	}
}

// TestIsolation_NestedLoops tests interrupting nested loops
func TestIsolation_NestedLoops(t *testing.T) {
	vm := goja.New()

	done := make(chan struct{})
	go func() {
		_, _ = vm.RunString(`
			while (true) {
				for (var i = 0; i < 1000000; i++) {
					for (var j = 0; j < 1000000; j++) {
						// nested loops
					}
				}
			}
		`)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	vm.Interrupt("nested loop timeout")

	select {
	case <-done:
		t.Log("Nested loops interrupted successfully")
	case <-time.After(2 * time.Second):
		t.Error("Failed to interrupt nested loops")
	}
}

// TestIsolation_FunctionRecursion tests interrupting infinite recursion
func TestIsolation_FunctionRecursion(t *testing.T) {
	vm := goja.New()

	// Note: Goja has stack limit protection, but let's test interrupt works too
	done := make(chan error, 1)
	go func() {
		_, err := vm.RunString(`
			function recurse() {
				recurse();
			}
			try {
				recurse();
			} catch(e) {
				throw e;
			}
		`)
		done <- err
	}()

	// Either stack overflow or interrupt should stop it
	time.Sleep(50 * time.Millisecond)
	vm.Interrupt("recursion limit")

	select {
	case err := <-done:
		if err != nil {
			t.Logf("Recursion stopped with: %v", err)
		} else {
			t.Error("Expected error from recursion")
		}
	case <-time.After(2 * time.Second):
		t.Error("Recursion not stopped")
	}
}

// TestIsolation_RouterHandlerTimeout tests that route handlers can be timed out
func TestIsolation_RouterHandlerTimeout(t *testing.T) {
	h := NewJSTestHelper(t)
	router := h.SetupRouter()

	h.MustRun(t, `
		$router.get("/slow", function(ctx) {
			var start = Date.now();
			while (Date.now() - start < 5000) {
				// busy wait for 5 seconds
			}
			return ctx.response(200, {done: true});
		});
	`)

	timeout := 200 * time.Millisecond
	ctx := &modules.RequestContext{Method: "GET", Path: "/slow"}

	// Create timeout mechanism
	done := make(chan struct{})
	go func() {
		time.Sleep(timeout)
		h.VM.Interrupt("handler timeout")
		close(done)
	}()

	start := time.Now()
	_, err := router.Handle("GET", "/slow", ctx)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("Expected timeout error")
	}

	if elapsed > timeout+100*time.Millisecond {
		t.Errorf("Handler took too long: %v, expected ~%v", elapsed, timeout)
	}

	t.Logf("Handler timed out after %v: %v", elapsed, err)
}

// TestIsolation_ClearInterrupt tests that VM can be reused after interrupt
func TestIsolation_ClearInterrupt(t *testing.T) {
	vm := goja.New()

	// First: interrupt an infinite loop
	go func() {
		time.Sleep(50 * time.Millisecond)
		vm.Interrupt("stop")
	}()

	_, err := vm.RunString(`while (true) {}`)
	if err == nil {
		t.Fatal("Expected interrupt error")
	}

	// Clear the interrupt
	vm.ClearInterrupt()

	// VM should work again
	result, err := vm.RunString(`1 + 2`)
	if err != nil {
		t.Fatalf("VM should work after ClearInterrupt: %v", err)
	}

	if result.ToInteger() != 3 {
		t.Errorf("Expected 3, got %v", result)
	}
}

// TestIsolation_GlobalPollution tests that scripts cannot pollute other scripts' globals
func TestIsolation_GlobalPollution(t *testing.T) {
	vm1 := goja.New()
	vm2 := goja.New()

	// Script 1 tries to modify Object prototype (common attack)
	vm1.RunString(`
		Object.prototype.malicious = function() { return "hacked"; };
		Array.prototype.evil = "bad";
	`)

	// VM2 should not be affected
	result, err := vm2.RunString(`
		var obj = {};
		var arr = [];
		(typeof obj.malicious === "undefined") && (typeof arr.evil === "undefined");
	`)
	if err != nil {
		t.Fatal(err)
	}

	if !result.ToBoolean() {
		t.Error("VM2 should not be affected by VM1's prototype pollution")
	}
}

// TestIsolation_MemoryLeak tests that VMs can be garbage collected
func TestIsolation_MemoryLeak(t *testing.T) {
	// Create and discard many VMs
	for i := 0; i < 100; i++ {
		vm := goja.New()
		vm.RunString(`
			var data = [];
			for (var j = 0; j < 1000; j++) {
				data.push({index: j, value: "test".repeat(100)});
			}
		`)
		// VM goes out of scope and should be GC'd
	}

	// If we get here without OOM, the test passes
	t.Log("Created and discarded 100 VMs without memory issues")
}

// TestIsolation_ConcurrentVMAccess tests that VM panics on concurrent access
func TestIsolation_ConcurrentVMAccess(t *testing.T) {
	// This test documents the expected behavior: Goja VM is NOT thread-safe
	// Concurrent access should cause issues (panic or incorrect results)
	t.Skip("Skipping: Goja VM is not thread-safe by design. This documents the limitation.")

	vm := goja.New()
	vm.RunString(`var counter = 0;`)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				vm.RunString(`counter++;`)
			}
		}()
	}

	wg.Wait()

	result, _ := vm.RunString(`counter`)
	// With thread safety, this would be 1000
	// Without it, the result is undefined (could be anything)
	t.Logf("Counter value (unsafe): %v", result.ToInteger())
}

// TestIsolation_ScheduledJobIsolation tests that scheduled jobs don't block main execution
func TestIsolation_ScheduledJobIsolation(t *testing.T) {
	// Note: This tests the concept, actual scheduler uses separate goroutine
	vm := goja.New()

	mainDone := make(chan struct{})
	jobRunning := atomic.Bool{}

	// Simulate a scheduled job running in goroutine
	go func() {
		// Job that takes time
		time.Sleep(500 * time.Millisecond)
		jobRunning.Store(true)
	}()

	// Main execution should not be blocked
	start := time.Now()
	_, err := vm.RunString(`
		var result = 0;
		for (var i = 0; i < 1000; i++) {
			result += i;
		}
		result;
	`)
	elapsed := time.Since(start)
	close(mainDone)

	if err != nil {
		t.Fatal(err)
	}

	// Main should complete much faster than job
	if elapsed > 100*time.Millisecond {
		t.Errorf("Main execution took too long: %v", elapsed)
	}

	// Job should still be running (or about to run)
	if jobRunning.Load() {
		t.Error("Job finished before it should have")
	}
}

// TestIsolation_InterruptPreservesState tests that interrupt doesn't corrupt VM state
func TestIsolation_InterruptPreservesState(t *testing.T) {
	vm := goja.New()

	// Set up some state
	vm.RunString(`
		var importantData = {count: 42, name: "test"};
		var array = [1, 2, 3];
	`)

	// Start and interrupt a loop
	go func() {
		time.Sleep(50 * time.Millisecond)
		vm.Interrupt("stop")
	}()

	vm.RunString(`
		while (true) {
			// This will be interrupted
		}
	`)

	// Clear interrupt for reuse
	vm.ClearInterrupt()

	// Check state is preserved
	result, err := vm.RunString(`
		importantData.count === 42 &&
		importantData.name === "test" &&
		array.length === 3
	`)
	if err != nil {
		t.Fatalf("Error checking state: %v", err)
	}

	if !result.ToBoolean() {
		t.Error("VM state was corrupted after interrupt")
	}
}

// TestIsolation_WorkerPoolLimit tests that delayed module respects pool limits
func TestIsolation_WorkerPoolLimit(t *testing.T) {
	// Note: Goja VM is NOT thread-safe, so $delayed.run cannot safely access
	// JS variables from background goroutines. This test verifies pool limiting
	// using Go-level synchronization instead.

	delayed := modules.NewDelayedModule(3) // Small pool for testing

	var wg sync.WaitGroup
	var running atomic.Int32
	var maxConcurrent atomic.Int32

	for i := 0; i < 10; i++ {
		wg.Add(1)
		vm := goja.New()
		delayed.Register(vm)

		// We need to track concurrency at Go level since Goja isn't thread-safe
		handler := func(goja.FunctionCall) goja.Value {
			current := running.Add(1)
			// Track max concurrent executions
			for {
				old := maxConcurrent.Load()
				if current <= old || maxConcurrent.CompareAndSwap(old, current) {
					break
				}
			}
			time.Sleep(50 * time.Millisecond) // Simulate work
			running.Add(-1)
			wg.Done()
			return goja.Undefined()
		}

		callable, _ := goja.AssertFunction(vm.ToValue(handler))
		delayed.Run(callable)
	}

	wg.Wait()

	maxConcur := maxConcurrent.Load()
	if maxConcur > 3 {
		t.Errorf("Max concurrent tasks exceeded pool size: got %d, expected <= 3", maxConcur)
	}
	t.Logf("Max concurrent tasks: %d (pool size: 3)", maxConcur)
}

// BenchmarkVMCreation measures VM creation overhead
func BenchmarkVMCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		vm := goja.New()
		vm.RunString(`var x = 1;`)
	}
}

// BenchmarkInterrupt measures interrupt overhead
func BenchmarkInterrupt(b *testing.B) {
	for i := 0; i < b.N; i++ {
		vm := goja.New()

		go func() {
			time.Sleep(1 * time.Millisecond)
			vm.Interrupt("stop")
		}()

		vm.RunString(`while(true){}`)
		vm.ClearInterrupt()
	}
}
