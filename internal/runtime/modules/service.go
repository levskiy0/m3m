package modules

import (
	"sync"
	"time"

	"github.com/dop251/goja"
)

// ServiceModule manages the lifecycle of a service
type ServiceModule struct {
	bootCallbacks     []goja.Callable
	startCallbacks    []goja.Callable
	shutdownCallbacks []goja.Callable
	mu                sync.Mutex
	vm                *goja.Runtime
	booted            bool
	started           bool
	shutdownTimeout   time.Duration
}

func NewServiceModule(vm *goja.Runtime, shutdownTimeout time.Duration) *ServiceModule {
	if shutdownTimeout == 0 {
		shutdownTimeout = 30 * time.Second
	}
	return &ServiceModule{
		bootCallbacks:     []goja.Callable{},
		startCallbacks:    []goja.Callable{},
		shutdownCallbacks: []goja.Callable{},
		vm:                vm,
		shutdownTimeout:   shutdownTimeout,
	}
}

// Boot registers a callback to be called during service initialization
func (s *ServiceModule) Boot(callback goja.Callable) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bootCallbacks = append(s.bootCallbacks, callback)
}

// Start registers a callback to be called when service is ready
func (s *ServiceModule) Start(callback goja.Callable) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.startCallbacks = append(s.startCallbacks, callback)
}

// Shutdown registers a callback to be called when service is stopping
func (s *ServiceModule) Shutdown(callback goja.Callable) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.shutdownCallbacks = append(s.shutdownCallbacks, callback)
}

// ExecuteBoot executes all boot callbacks
func (s *ServiceModule) ExecuteBoot() error {
	s.mu.Lock()
	callbacks := make([]goja.Callable, len(s.bootCallbacks))
	copy(callbacks, s.bootCallbacks)
	s.mu.Unlock()

	for _, cb := range callbacks {
		if _, err := cb(goja.Undefined()); err != nil {
			return err
		}
	}

	s.mu.Lock()
	s.booted = true
	s.mu.Unlock()

	return nil
}

// ExecuteStart executes all start callbacks
func (s *ServiceModule) ExecuteStart() error {
	s.mu.Lock()
	callbacks := make([]goja.Callable, len(s.startCallbacks))
	copy(callbacks, s.startCallbacks)
	s.mu.Unlock()

	for _, cb := range callbacks {
		if _, err := cb(goja.Undefined()); err != nil {
			return err
		}
	}

	s.mu.Lock()
	s.started = true
	s.mu.Unlock()

	return nil
}

// ExecuteShutdown executes all shutdown callbacks with timeout
func (s *ServiceModule) ExecuteShutdown() error {
	s.mu.Lock()
	callbacks := make([]goja.Callable, len(s.shutdownCallbacks))
	copy(callbacks, s.shutdownCallbacks)
	s.mu.Unlock()

	done := make(chan error, 1)

	go func() {
		for _, cb := range callbacks {
			if _, err := cb(goja.Undefined()); err != nil {
				done <- err
				return
			}
		}
		done <- nil
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(s.shutdownTimeout):
		return nil // Timeout, continue with shutdown
	}
}

// IsBooted returns whether boot phase has completed
func (s *ServiceModule) IsBooted() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.booted
}

// IsStarted returns whether start phase has completed
func (s *ServiceModule) IsStarted() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.started
}

// GetJSObject returns the JavaScript object for this module
func (s *ServiceModule) GetJSObject() map[string]interface{} {
	return map[string]interface{}{
		"boot":     s.Boot,
		"start":    s.Start,
		"shutdown": s.Shutdown,
	}
}
