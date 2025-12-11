package modules

import (
	"sync"

	"github.com/dop251/goja"
	"github.com/google/uuid"
	"github.com/levskiy0/m3m/pkg/schema"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UIBroadcaster interface for sending UI requests to specific sessions
type UIBroadcaster interface {
	SendUIRequest(projectID, sessionID string, data interface{})
}

// UIRequest represents a pending UI request waiting for response
type UIRequest struct {
	callback  goja.Callable
	sessionID string
}

// UIDialogType represents the type of UI dialog
type UIDialogType string

const (
	UIDialogAlert   UIDialogType = "alert"
	UIDialogConfirm UIDialogType = "confirm"
	UIDialogPrompt  UIDialogType = "prompt"
	UIDialogForm    UIDialogType = "form"
	UIDialogToast   UIDialogType = "toast"
)

// UIRequestData is sent to the frontend via WebSocket
type UIRequestData struct {
	RequestID  string       `json:"requestId"`
	DialogType UIDialogType `json:"dialogType"`
	Options    interface{}  `json:"options"`
}

// UIModule provides interactive UI dialogs via WebSocket
type UIModule struct {
	vm             *goja.Runtime
	projectID      primitive.ObjectID
	broadcaster    UIBroadcaster
	pendingReqs    map[string]*UIRequest
	mu             sync.Mutex
	currentSession string // set during action execution (WebSocket session ID)
}

// NewUIModule creates a new UIModule
func NewUIModule(vm *goja.Runtime, projectID primitive.ObjectID, broadcaster UIBroadcaster) *UIModule {
	return &UIModule{
		vm:          vm,
		projectID:   projectID,
		broadcaster: broadcaster,
		pendingReqs: make(map[string]*UIRequest),
	}
}

// Name returns the module name for JavaScript
func (u *UIModule) Name() string {
	return "$ui"
}

// Register registers the module into the JavaScript VM
func (u *UIModule) Register(vm interface{}) {
	vm.(*goja.Runtime).Set(u.Name(), map[string]interface{}{
		"alert":   u.Alert,
		"confirm": u.Confirm,
		"prompt":  u.Prompt,
		"form":    u.Form,
		"toast":   u.Toast,
	})
}

// SetCurrentSession sets the current session ID for UI requests
func (u *UIModule) SetCurrentSession(sessionID string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.currentSession = sessionID
}

// ClearCurrentSession clears the current session ID
func (u *UIModule) ClearCurrentSession() {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.currentSession = ""
}

// HandleResponse handles a UI response from the frontend
func (u *UIModule) HandleResponse(requestID string, data interface{}) {
	u.mu.Lock()
	req, ok := u.pendingReqs[requestID]
	if ok {
		delete(u.pendingReqs, requestID)
	}
	u.mu.Unlock()

	if !ok {
		return
	}

	if req.callback != nil {
		// Restore session context for the callback and async code that may run later
		// Don't clear it - let it persist for async operations like $schedule.delay
		u.SetCurrentSession(req.sessionID)

		// Convert data to goja.Value and call callback
		jsData := u.vm.ToValue(data)
		req.callback(goja.Undefined(), jsData)
	}
}

// getSessionID extracts sessionId from options or falls back to currentSession
func (u *UIModule) getSessionID(options interface{}) string {
	// Try to get sessionId from options (for async callbacks)
	if opts, ok := options.(map[string]interface{}); ok {
		if sid, ok := opts["sessionId"].(string); ok && sid != "" {
			return sid
		}
	}

	// Fall back to currentSession
	u.mu.Lock()
	currentSession := u.currentSession
	u.mu.Unlock()
	return currentSession
}

// Alert shows an alert notification (fire-and-forget)
func (u *UIModule) Alert(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(u.vm.NewTypeError("$ui.alert requires options argument"))
	}

	options := call.Arguments[0].Export()
	sessionID := u.getSessionID(options)

	if sessionID == "" || u.broadcaster == nil {
		// No session context or no broadcaster - silently ignore
		return goja.Undefined()
	}

	// Send alert to session (fire-and-forget, no callback)
	u.broadcaster.SendUIRequest(
		u.projectID.Hex(),
		sessionID,
		UIRequestData{
			RequestID:  uuid.New().String(),
			DialogType: UIDialogAlert,
			Options:    options,
		},
	)

	return goja.Undefined()
}

// Toast shows a toast notification (fire-and-forget, auto-dismisses)
func (u *UIModule) Toast(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 1 {
		panic(u.vm.NewTypeError("$ui.toast requires options argument"))
	}

	options := call.Arguments[0].Export()
	sessionID := u.getSessionID(options)

	if sessionID == "" || u.broadcaster == nil {
		// No session context or no broadcaster - silently ignore
		return goja.Undefined()
	}

	// Send toast to session (fire-and-forget, no callback)
	u.broadcaster.SendUIRequest(
		u.projectID.Hex(),
		sessionID,
		UIRequestData{
			RequestID:  uuid.New().String(),
			DialogType: UIDialogToast,
			Options:    options,
		},
	)

	return goja.Undefined()
}

// Confirm shows a confirmation dialog with yes/no buttons
func (u *UIModule) Confirm(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(u.vm.NewTypeError("$ui.confirm requires options and callback arguments"))
	}

	options := call.Arguments[0].Export()
	callback, ok := goja.AssertFunction(call.Arguments[1])
	if !ok {
		panic(u.vm.NewTypeError("second argument must be a function"))
	}

	sessionID := u.getSessionID(options)

	if sessionID == "" || u.broadcaster == nil {
		// No session context - call callback with null
		callback(goja.Undefined(), goja.Null())
		return goja.Undefined()
	}

	requestID := uuid.New().String()

	u.mu.Lock()
	u.pendingReqs[requestID] = &UIRequest{
		callback:  callback,
		sessionID: sessionID,
	}
	u.mu.Unlock()

	u.broadcaster.SendUIRequest(
		u.projectID.Hex(),
		sessionID,
		UIRequestData{
			RequestID:  requestID,
			DialogType: UIDialogConfirm,
			Options:    options,
		},
	)

	return goja.Undefined()
}

// Prompt shows a text input dialog
func (u *UIModule) Prompt(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(u.vm.NewTypeError("$ui.prompt requires options and callback arguments"))
	}

	options := call.Arguments[0].Export()
	callback, ok := goja.AssertFunction(call.Arguments[1])
	if !ok {
		panic(u.vm.NewTypeError("second argument must be a function"))
	}

	sessionID := u.getSessionID(options)

	if sessionID == "" || u.broadcaster == nil {
		// No session context - call callback with null
		callback(goja.Undefined(), goja.Null())
		return goja.Undefined()
	}

	requestID := uuid.New().String()

	u.mu.Lock()
	u.pendingReqs[requestID] = &UIRequest{
		callback:  callback,
		sessionID: sessionID,
	}
	u.mu.Unlock()

	u.broadcaster.SendUIRequest(
		u.projectID.Hex(),
		sessionID,
		UIRequestData{
			RequestID:  requestID,
			DialogType: UIDialogPrompt,
			Options:    options,
		},
	)

	return goja.Undefined()
}

// Form shows a form dialog with multiple fields
func (u *UIModule) Form(call goja.FunctionCall) goja.Value {
	if len(call.Arguments) < 2 {
		panic(u.vm.NewTypeError("$ui.form requires options and callback arguments"))
	}

	options := call.Arguments[0].Export()
	callback, ok := goja.AssertFunction(call.Arguments[1])
	if !ok {
		panic(u.vm.NewTypeError("second argument must be a function"))
	}

	sessionID := u.getSessionID(options)

	if sessionID == "" || u.broadcaster == nil {
		// No session context - call callback with null
		callback(goja.Undefined(), goja.Null())
		return goja.Undefined()
	}

	requestID := uuid.New().String()

	u.mu.Lock()
	u.pendingReqs[requestID] = &UIRequest{
		callback:  callback,
		sessionID: sessionID,
	}
	u.mu.Unlock()

	u.broadcaster.SendUIRequest(
		u.projectID.Hex(),
		sessionID,
		UIRequestData{
			RequestID:  requestID,
			DialogType: UIDialogForm,
			Options:    options,
		},
	)

	return goja.Undefined()
}

// Cleanup clears all pending requests (called on runtime shutdown)
func (u *UIModule) Cleanup() {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.pendingReqs = make(map[string]*UIRequest)
	u.currentSession = ""
}

// GetSchema implements JSSchemaProvider
func (u *UIModule) GetSchema() schema.ModuleSchema {
	return schema.ModuleSchema{
		Name:        "$ui",
		Description: "Interactive UI dialogs shown to users via WebSocket",
		Types: []schema.TypeSchema{
			{
				Name:        "AlertOptions",
				Description: "Options for alert dialog",
				Fields: []schema.ParamSchema{
					{Name: "title", Type: "string", Description: "Dialog title", Optional: true},
					{Name: "text", Type: "string", Description: "Dialog message text", Optional: true},
					{Name: "severity", Type: "'info' | 'success' | 'warning' | 'error'", Description: "Alert severity (default: 'info')", Optional: true},
				},
			},
			{
				Name:        "ToastOptions",
				Description: "Options for toast notification",
				Fields: []schema.ParamSchema{
					{Name: "text", Type: "string", Description: "Toast message text"},
					{Name: "icon", Type: "string", Description: "Icon name", Optional: true},
					{Name: "severity", Type: "'info' | 'success' | 'warning' | 'error'", Description: "Toast severity (default: 'info')", Optional: true},
				},
			},
			{
				Name:        "ConfirmOptions",
				Description: "Options for confirm dialog",
				Fields: []schema.ParamSchema{
					{Name: "title", Type: "string", Description: "Dialog title", Optional: true},
					{Name: "text", Type: "string", Description: "Dialog message text", Optional: true},
					{Name: "yes", Type: "string", Description: "Yes button label (default: 'Yes')", Optional: true},
					{Name: "no", Type: "string", Description: "No button label (default: 'No')", Optional: true},
				},
			},
			{
				Name:        "PromptOptions",
				Description: "Options for prompt dialog",
				Fields: []schema.ParamSchema{
					{Name: "title", Type: "string", Description: "Dialog title", Optional: true},
					{Name: "text", Type: "string", Description: "Dialog message text", Optional: true},
					{Name: "placeholder", Type: "string", Description: "Input placeholder text", Optional: true},
					{Name: "defaultValue", Type: "string", Description: "Default input value", Optional: true},
				},
			},
			{
				Name:        "FormField",
				Description: "Form field definition",
				Fields: []schema.ParamSchema{
					{Name: "name", Type: "string", Description: "Field name (used as key in result)"},
					{Name: "type", Type: "'input' | 'textarea' | 'checkbox' | 'select' | 'combobox' | 'radiogroup' | 'date' | 'datetime'", Description: "Field type"},
					{Name: "label", Type: "string", Description: "Field label", Optional: true},
					{Name: "hint", Type: "string", Description: "Helper text below the field", Optional: true},
					{Name: "colspan", Type: "number | 'full'", Description: "Column span in 6-column grid (1-6 or 'full')", Optional: true},
					{Name: "required", Type: "boolean", Description: "Whether field is required", Optional: true},
					{Name: "placeholder", Type: "string", Description: "Placeholder text", Optional: true},
					{Name: "defaultValue", Type: "any", Description: "Default value", Optional: true},
					{Name: "options", Type: "string[] | {label: string, value: string}[]", Description: "Options for select/combobox/radiogroup", Optional: true},
				},
			},
			{
				Name:        "FormAction",
				Description: "Form action button",
				Fields: []schema.ParamSchema{
					{Name: "label", Type: "string", Description: "Button label"},
					{Name: "variant", Type: "'default' | 'outline' | 'destructive'", Description: "Button variant", Optional: true},
					{Name: "action", Type: "'submit' | 'cancel' | string", Description: "Action identifier"},
				},
			},
			{
				Name:        "FormOptions",
				Description: "Options for form dialog",
				Fields: []schema.ParamSchema{
					{Name: "title", Type: "string", Description: "Dialog title", Optional: true},
					{Name: "text", Type: "string", Description: "Dialog description text", Optional: true},
					{Name: "schema", Type: "FormField[]", Description: "Form fields schema"},
					{Name: "actions", Type: "FormAction[]", Description: "Form action buttons", Optional: true},
				},
			},
			{
				Name:        "FormResult",
				Description: "Form submission result",
				Fields: []schema.ParamSchema{
					{Name: "action", Type: "string", Description: "Action that was triggered"},
					{Name: "data", Type: "object", Description: "Form field values keyed by field name"},
				},
			},
		},
		Methods: []schema.MethodSchema{
			{
				Name:        "alert",
				Description: "Show an alert notification (fire-and-forget, no callback)",
				Params: []schema.ParamSchema{
					{Name: "options", Type: "AlertOptions", Description: "Alert options"},
				},
			},
			{
				Name:        "toast",
				Description: "Show a toast notification (fire-and-forget, auto-dismisses)",
				Params: []schema.ParamSchema{
					{Name: "options", Type: "ToastOptions", Description: "Toast options"},
				},
			},
			{
				Name:        "confirm",
				Description: "Show a confirmation dialog with Yes/No buttons",
				Params: []schema.ParamSchema{
					{Name: "options", Type: "ConfirmOptions", Description: "Confirm options"},
					{Name: "callback", Type: "(confirmed: boolean | null) => void", Description: "Called with true/false or null if no user context"},
				},
			},
			{
				Name:        "prompt",
				Description: "Show a text input dialog",
				Params: []schema.ParamSchema{
					{Name: "options", Type: "PromptOptions", Description: "Prompt options"},
					{Name: "callback", Type: "(value: string | null) => void", Description: "Called with input value or null if cancelled/no user"},
				},
			},
			{
				Name:        "form",
				Description: "Show a form dialog with multiple fields. Callback can return validation errors.",
				Params: []schema.ParamSchema{
					{Name: "options", Type: "FormOptions", Description: "Form options with schema and actions"},
					{Name: "callback", Type: "(result: FormResult | null) => void | {[field: string]: string}", Description: "Called with form result or null. Return object with field errors to show validation."},
				},
			},
		},
	}
}

// GetUISchema returns the UI schema (static version)
func GetUISchema() schema.ModuleSchema {
	return (&UIModule{}).GetSchema()
}
