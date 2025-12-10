package modules

import (
	"fmt"
	"sync"

	"github.com/dop251/goja"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/levskiy0/m3m/internal/domain"
	"github.com/levskiy0/m3m/pkg/schema"
)

// ActionBroadcaster interface for broadcasting action state changes
type ActionBroadcaster interface {
	BroadcastActionStates(projectID string, states []domain.ActionRuntimeState)
}

// actionHandler stores the handler and action info
type actionHandler struct {
	name    string
	slug    string
	handler goja.Callable
}

// ActionsModule manages action handlers and states
type ActionsModule struct {
	handlers    map[string]*actionHandler     // slug -> handler
	states      map[string]domain.ActionState // slug -> state
	mu          sync.RWMutex
	vm          *goja.Runtime
	projectID   primitive.ObjectID
	broadcaster ActionBroadcaster
}

// NewActionsModule creates a new ActionsModule
func NewActionsModule(vm *goja.Runtime, projectID primitive.ObjectID, broadcaster ActionBroadcaster) *ActionsModule {
	return &ActionsModule{
		handlers:    make(map[string]*actionHandler),
		states:      make(map[string]domain.ActionState),
		vm:          vm,
		projectID:   projectID,
		broadcaster: broadcaster,
	}
}

// Name returns the module name for JavaScript
func (m *ActionsModule) Name() string {
	return "$actions"
}

// Register registers the module into the JavaScript VM
func (m *ActionsModule) Register(vm interface{}) {
	vm.(*goja.Runtime).Set(m.Name(), map[string]interface{}{
		"on": m.On,
	})
}

// On registers an action handler
// Usage: $actions.on('action-slug', (a) => { a.loader(); ... a.enable(); })
func (m *ActionsModule) On(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(m.vm.NewTypeError("$actions.on requires slug and handler arguments"))
	}

	slug := call.Arguments[0].String()
	handlerArg := call.Arguments[1]

	handler, ok := goja.AssertFunction(handlerArg)
	if !ok {
		panic(m.vm.NewTypeError("second argument must be a function"))
	}

	// Optional name from third argument
	name := slug
	if len(call.Arguments) >= 3 {
		name = call.Arguments[2].String()
	}

	m.mu.Lock()
	m.handlers[slug] = &actionHandler{
		name:    name,
		slug:    slug,
		handler: handler,
	}
	// Initialize state as enabled
	m.states[slug] = domain.ActionStateEnabled
	m.mu.Unlock()

	return goja.Undefined()
}

// Trigger executes an action handler
func (m *ActionsModule) Trigger(slug string) error {
	m.mu.RLock()
	h, ok := m.handlers[slug]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("action handler not registered: %s", slug)
	}

	// Create action context with methods
	ctx := m.createActionContext(h.name, slug)

	// Call the handler with the context
	_, err := h.handler(goja.Undefined(), m.vm.ToValue(ctx))
	return err
}

// createActionContext creates a JS object with action info and control methods
func (m *ActionsModule) createActionContext(name, slug string) map[string]interface{} {
	return map[string]interface{}{
		"name": name,
		"slug": slug,
		"disable": func() {
			m.setState(slug, domain.ActionStateDisabled)
		},
		"enable": func() {
			m.setState(slug, domain.ActionStateEnabled)
		},
		"loader": func() {
			m.setState(slug, domain.ActionStateLoading)
		},
	}
}

// setState updates the state of an action and broadcasts the change
func (m *ActionsModule) setState(slug string, state domain.ActionState) {
	m.mu.Lock()
	m.states[slug] = state
	states := m.getStatesLocked()
	m.mu.Unlock()

	// Broadcast the change
	if m.broadcaster != nil {
		m.broadcaster.BroadcastActionStates(m.projectID.Hex(), states)
	}
}

// GetStates returns all action states
func (m *ActionsModule) GetStates() []domain.ActionRuntimeState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.getStatesLocked()
}

// getStatesLocked returns states (caller must hold lock)
func (m *ActionsModule) getStatesLocked() []domain.ActionRuntimeState {
	states := make([]domain.ActionRuntimeState, 0, len(m.states))
	for slug, state := range m.states {
		states = append(states, domain.ActionRuntimeState{
			Slug:  slug,
			State: state,
		})
	}
	return states
}

// ResetStates resets all action states to enabled
func (m *ActionsModule) ResetStates() {
	m.mu.Lock()
	for slug := range m.states {
		m.states[slug] = domain.ActionStateEnabled
	}
	states := m.getStatesLocked()
	m.mu.Unlock()

	// Broadcast the reset
	if m.broadcaster != nil && len(states) > 0 {
		m.broadcaster.BroadcastActionStates(m.projectID.Hex(), states)
	}
}

// HasHandler checks if a handler is registered for the given slug
func (m *ActionsModule) HasHandler(slug string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.handlers[slug]
	return ok
}

// GetSchema implements JSSchemaProvider
func (m *ActionsModule) GetSchema() schema.ModuleSchema {
	return schema.ModuleSchema{
		Name:        "$actions",
		Description: "Action handlers for interactive UI buttons",
		Types: []schema.TypeSchema{
			{
				Name:        "ActionContext",
				Description: "Context object passed to action handlers",
				Fields: []schema.ParamSchema{
					{Name: "name", Type: "string", Description: "Action name"},
					{Name: "slug", Type: "string", Description: "Action slug identifier"},
					{Name: "disable", Type: "() => void", Description: "Disable the action button"},
					{Name: "enable", Type: "() => void", Description: "Enable the action button"},
					{Name: "loader", Type: "() => void", Description: "Show loading spinner on the button"},
				},
			},
		},
		Methods: []schema.MethodSchema{
			{
				Name:        "on",
				Description: "Register a handler for an action",
				Params: []schema.ParamSchema{
					{Name: "slug", Type: "string", Description: "Action slug identifier"},
					{Name: "handler", Type: "(ctx: ActionContext) => void", Description: "Handler function called when action is triggered"},
				},
			},
		},
	}
}

// GetActionsSchema returns the actions schema (static version)
func GetActionsSchema() schema.ModuleSchema {
	return (&ActionsModule{}).GetSchema()
}
