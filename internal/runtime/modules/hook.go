package modules

import (
	"fmt"
	"sync"
	"time"

	"github.com/dop251/goja"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/levskiy0/m3m/internal/domain"
	"github.com/levskiy0/m3m/pkg/schema"
)

// HookBroadcaster interface for broadcasting action state changes
type HookBroadcaster interface {
	BroadcastActionStates(projectID string, states []domain.ActionRuntimeState)
}

// ModelHookType represents the type of model hook
type ModelHookType string

const (
	ModelHookInsert ModelHookType = "insert"
	ModelHookUpdate ModelHookType = "update"
	ModelHookDelete ModelHookType = "delete"
)

// actionHandler stores the handler and action info
type actionHandler struct {
	name    string
	slug    string
	handler goja.Callable
}

// modelHandler stores the handler for model hooks
type modelHandler struct {
	modelName string
	hookType  ModelHookType
	handler   goja.Callable
}

// HookModule manages action and model handlers
type HookModule struct {
	actionHandlers   map[string]*actionHandler                    // slug -> handler
	modelHandlers    map[string]map[ModelHookType][]*modelHandler // modelName -> hookType -> handlers
	actionStates     map[string]domain.ActionState                // slug -> state
	mu               sync.RWMutex
	vm               *goja.Runtime
	projectID        primitive.ObjectID
	broadcaster      HookBroadcaster
	currentUserID    string // current user ID for action context
	currentSessionID string // current WebSocket session ID for UI targeting
}

// NewHookModule creates a new HookModule
func NewHookModule(vm *goja.Runtime, projectID primitive.ObjectID, broadcaster HookBroadcaster) *HookModule {
	return &HookModule{
		actionHandlers: make(map[string]*actionHandler),
		modelHandlers:  make(map[string]map[ModelHookType][]*modelHandler),
		actionStates:   make(map[string]domain.ActionState),
		vm:             vm,
		projectID:      projectID,
		broadcaster:    broadcaster,
	}
}

// Name returns the module name for JavaScript
func (m *HookModule) Name() string {
	return "$hook"
}

// Register registers the module into the JavaScript VM
func (m *HookModule) Register(vm interface{}) {
	vm.(*goja.Runtime).Set(m.Name(), map[string]interface{}{
		"onAction":      m.OnAction,
		"onModelInsert": m.OnModelInsert,
		"onModelUpdate": m.OnModelUpdate,
		"onModelDelete": m.OnModelDelete,
	})
}

// OnAction registers an action handler
// Usage: $hook.onAction('action-slug', (a) => { a.loader(); ... a.enable(); })
func (m *HookModule) OnAction(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(m.vm.NewTypeError("$hook.onAction requires slug and handler arguments"))
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
	m.actionHandlers[slug] = &actionHandler{
		name:    name,
		slug:    slug,
		handler: handler,
	}
	// Initialize state as enabled
	m.actionStates[slug] = domain.ActionStateEnabled
	states := m.getActionStatesLocked()
	m.mu.Unlock()

	// Broadcast the initial state so frontend gets updated on restart
	if m.broadcaster != nil {
		m.broadcaster.BroadcastActionStates(m.projectID.Hex(), states)
	}

	return goja.Undefined()
}

// OnModelInsert registers a handler for model insert events
// Usage: $hook.onModelInsert('modelname', (m) => { ... })
func (m *HookModule) OnModelInsert(call goja.FunctionCall) goja.Value {
	return m.registerModelHandler(call, ModelHookInsert)
}

// OnModelUpdate registers a handler for model update events
// Usage: $hook.onModelUpdate('modelname', (m) => { ... })
func (m *HookModule) OnModelUpdate(call goja.FunctionCall) goja.Value {
	return m.registerModelHandler(call, ModelHookUpdate)
}

// OnModelDelete registers a handler for model delete events
// Usage: $hook.onModelDelete('modelname', (m) => { ... })
func (m *HookModule) OnModelDelete(call goja.FunctionCall) goja.Value {
	return m.registerModelHandler(call, ModelHookDelete)
}

// registerModelHandler registers a model hook handler
func (m *HookModule) registerModelHandler(call goja.FunctionCall, hookType ModelHookType) goja.Value {
	if len(call.Arguments) < 2 {
		panic(m.vm.NewTypeError(fmt.Sprintf("$hook.onModel%s requires modelName and handler arguments", capitalizeFirst(string(hookType)))))
	}

	modelName := call.Arguments[0].String()
	handlerArg := call.Arguments[1]

	handler, ok := goja.AssertFunction(handlerArg)
	if !ok {
		panic(m.vm.NewTypeError("second argument must be a function"))
	}

	m.mu.Lock()
	if m.modelHandlers[modelName] == nil {
		m.modelHandlers[modelName] = make(map[ModelHookType][]*modelHandler)
	}
	m.modelHandlers[modelName][hookType] = append(m.modelHandlers[modelName][hookType], &modelHandler{
		modelName: modelName,
		hookType:  hookType,
		handler:   handler,
	})
	m.mu.Unlock()

	return goja.Undefined()
}

// TriggerAction executes an action handler
func (m *HookModule) TriggerAction(slug string) error {
	m.mu.RLock()
	h, ok := m.actionHandlers[slug]
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

// TriggerModelHook executes model hook handlers
func (m *HookModule) TriggerModelHook(modelSlug string, hookType ModelHookType, data map[string]interface{}) error {
	m.mu.RLock()
	handlers, ok := m.modelHandlers[modelSlug]
	if !ok {
		m.mu.RUnlock()
		return nil // No handlers registered for this model
	}
	hookHandlers := handlers[hookType]
	m.mu.RUnlock()

	if len(hookHandlers) == 0 {
		return nil
	}

	// Prepare clean data for hooks
	cleanData := make(map[string]interface{})
	for k, v := range data {
		// Skip internal _model_id field
		if k == "_model_id" {
			continue
		}
		// Convert ObjectID to hex string
		if id, ok := v.(primitive.ObjectID); ok {
			cleanData[k] = id.Hex()
			continue
		}
		// Convert time.Time to UTC
		if t, ok := v.(time.Time); ok {
			cleanData[k] = t.UTC()
			continue
		}
		// Convert primitive.DateTime to UTC time.Time
		if dt, ok := v.(primitive.DateTime); ok {
			cleanData[k] = dt.Time().UTC()
			continue
		}
		cleanData[k] = v
	}

	// Call all handlers
	for _, h := range hookHandlers {
		_, err := h.handler(goja.Undefined(), m.vm.ToValue(cleanData))
		if err != nil {
			return err
		}
	}

	return nil
}

// SetCurrentUser sets the current user ID for action context
func (m *HookModule) SetCurrentUser(userID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentUserID = userID
}

// ClearCurrentUser clears the current user ID
func (m *HookModule) ClearCurrentUser() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentUserID = ""
}

// SetCurrentSession sets the current session ID for UI targeting
func (m *HookModule) SetCurrentSession(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentSessionID = sessionID
}

// ClearCurrentSession clears the current session ID
func (m *HookModule) ClearCurrentSession() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentSessionID = ""
}

// createActionContext creates a JS object with action info and control methods
func (m *HookModule) createActionContext(name, slug string) map[string]interface{} {
	// Capture current user and session IDs at the time the action is triggered
	m.mu.RLock()
	userID := m.currentUserID
	sessionID := m.currentSessionID
	m.mu.RUnlock()

	return map[string]interface{}{
		"name":      name,
		"slug":      slug,
		"userId":    userID,    // Include userId for informational purposes
		"sessionId": sessionID, // Include sessionId for async $ui calls
		"disable": func() {
			m.setActionState(slug, domain.ActionStateDisabled)
		},
		"enable": func() {
			m.setActionState(slug, domain.ActionStateEnabled)
		},
		"loader": func() {
			m.setActionState(slug, domain.ActionStateLoading)
		},
	}
}

// setActionState updates the state of an action and broadcasts the change
func (m *HookModule) setActionState(slug string, state domain.ActionState) {
	m.mu.Lock()
	m.actionStates[slug] = state
	states := m.getActionStatesLocked()
	m.mu.Unlock()

	// Broadcast the change
	if m.broadcaster != nil {
		m.broadcaster.BroadcastActionStates(m.projectID.Hex(), states)
	}
}

// GetActionStates returns all action states
func (m *HookModule) GetActionStates() []domain.ActionRuntimeState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.getActionStatesLocked()
}

// getActionStatesLocked returns states (caller must hold lock)
func (m *HookModule) getActionStatesLocked() []domain.ActionRuntimeState {
	states := make([]domain.ActionRuntimeState, 0, len(m.actionStates))
	for slug, state := range m.actionStates {
		states = append(states, domain.ActionRuntimeState{
			Slug:  slug,
			State: state,
		})
	}
	return states
}

// ResetActionStates resets all action states to enabled
func (m *HookModule) ResetActionStates() {
	m.mu.Lock()
	for slug := range m.actionStates {
		m.actionStates[slug] = domain.ActionStateEnabled
	}
	states := m.getActionStatesLocked()
	m.mu.Unlock()

	// Broadcast the reset
	if m.broadcaster != nil && len(states) > 0 {
		m.broadcaster.BroadcastActionStates(m.projectID.Hex(), states)
	}
}

// HasActionHandler checks if a handler is registered for the given slug
func (m *HookModule) HasActionHandler(slug string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.actionHandlers[slug]
	return ok
}

// HasModelHandler checks if a handler is registered for the given model and hook type
func (m *HookModule) HasModelHandler(modelSlug string, hookType ModelHookType) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	handlers, ok := m.modelHandlers[modelSlug]
	if !ok {
		return false
	}
	return len(handlers[hookType]) > 0
}

// capitalizeFirst capitalizes the first letter of a string
func capitalizeFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]-32) + s[1:]
}

// GetSchema implements JSSchemaProvider
func (m *HookModule) GetSchema() schema.ModuleSchema {
	return schema.ModuleSchema{
		Name:        "$hook",
		Description: "Universal hook system for actions and model events",
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
				Name:        "onAction",
				Description: "Register a handler for an action button click",
				Params: []schema.ParamSchema{
					{Name: "slug", Type: "string", Description: "Action slug identifier"},
					{Name: "handler", Type: "(ctx: ActionContext) => void", Description: "Handler function called when action is triggered"},
				},
			},
			{
				Name:        "onModelInsert",
				Description: "Register a handler for model insert events (triggered only from frontend)",
				Params: []schema.ParamSchema{
					{Name: "modelName", Type: "string", Description: "Model slug name"},
					{Name: "handler", Type: "(data: object) => void", Description: "Handler function called when a record is inserted"},
				},
			},
			{
				Name:        "onModelUpdate",
				Description: "Register a handler for model update events (triggered only from frontend)",
				Params: []schema.ParamSchema{
					{Name: "modelName", Type: "string", Description: "Model slug name"},
					{Name: "handler", Type: "(data: object) => void", Description: "Handler function called when a record is updated"},
				},
			},
			{
				Name:        "onModelDelete",
				Description: "Register a handler for model delete events (triggered only from frontend)",
				Params: []schema.ParamSchema{
					{Name: "modelName", Type: "string", Description: "Model slug name"},
					{Name: "handler", Type: "(data: object) => void", Description: "Handler function called when a record is deleted"},
				},
			},
		},
	}
}

// GetHookSchema returns the hook schema (static version)
func GetHookSchema() schema.ModuleSchema {
	return (&HookModule{}).GetSchema()
}
