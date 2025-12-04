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

// Name returns the module name for JavaScript
func (s *ServiceModule) Name() string {
	return "$service"
}

// Register registers the module into the JavaScript VM
func (s *ServiceModule) Register(vm interface{}) {
	vm.(*goja.Runtime).Set(s.Name(), s.GetJSObject())
}

// GetSchema implements JSSchemaProvider
func (s *ServiceModule) GetSchema() JSModuleSchema {
	return JSModuleSchema{
		Name:        "$service",
		Description: "Service lifecycle management for boot, start, and shutdown phases",
		Methods: []JSMethodSchema{
			{
				Name:        "boot",
				Description: "Register a callback to run during service initialization",
				Params:      []JSParamSchema{{Name: "callback", Type: "() => void", Description: "Function to execute during boot phase"}},
			},
			{
				Name:        "start",
				Description: "Register a callback to run when service is ready",
				Params:      []JSParamSchema{{Name: "callback", Type: "() => void", Description: "Function to execute when service starts"}},
			},
			{
				Name:        "shutdown",
				Description: "Register a callback to run when service is stopping",
				Params:      []JSParamSchema{{Name: "callback", Type: "() => void", Description: "Function to execute during shutdown"}},
			},
		},
	}
}

// GetServiceSchema returns the service schema (static version)
func GetServiceSchema() JSModuleSchema {
	return (&ServiceModule{}).GetSchema()
}
