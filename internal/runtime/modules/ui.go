package modules

import (
	"fmt"
	"sync"

	"github.com/dop251/goja"
	"github.com/google/uuid"
	"github.com/levskiy0/m3m/pkg/schema"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UIBroadcaster interface for sending UI requests to specific users
type UIBroadcaster interface {
	SendUIRequest(projectID, userID string, data interface{})
}

// UIRequest represents a pending UI request waiting for response
type UIRequest struct {
	callback goja.Callable
	userID   string
}

// UIDialogType represents the type of UI dialog
type UIDialogType string

const (
	UIDialogAlert   UIDialogType = "alert"
	UIDialogConfirm UIDialogType = "confirm"
	UIDialogPrompt  UIDialogType = "prompt"
	UIDialogForm    UIDialogType = "form"
)

// UIRequestData is sent to the frontend via WebSocket
type UIRequestData struct {
	RequestID  string       `json:"requestId"`
	DialogType UIDialogType `json:"dialogType"`
	Options    interface{}  `json:"options"`
}

// UIModule provides interactive UI dialogs via WebSocket
type UIModule struct {
	vm          *goja.Runtime
	projectID   primitive.ObjectID
	broadcaster UIBroadcaster
	pendingReqs map[string]*UIRequest
	mu          sync.Mutex
	currentUser string // set during action execution
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
	})
}

// SetCurrentUser sets the current user ID for UI requests
func (u *UIModule) SetCurrentUser(userID string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.currentUser = userID
}

// ClearCurrentUser clears the current user ID
func (u *UIModule) ClearCurrentUser() {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.currentUser = ""
}

// HandleResponse handles a UI response from the frontend
func (u *UIModule) HandleResponse(requestID string, data interface{}) {
	fmt.Printf("[UI] HandleResponse called: requestID=%s, data=%v\n", requestID, data)

	u.mu.Lock()
	req, ok := u.pendingReqs[requestID]
	if ok {
		delete(u.pendingReqs, requestID)
	}
	u.mu.Unlock()

	if !ok {
		fmt.Printf("[UI] HandleResponse: request not found\n")
		return
	}

	fmt.Printf("[UI] HandleResponse: found request, userID=%s\n", req.userID)

	if req.callback != nil {
		// Restore user context for the callback and async code that may run later
		// Don't clear it - let it persist for async operations like $schedule.delay
		u.SetCurrentUser(req.userID)

		fmt.Printf("[UI] HandleResponse: calling callback, currentUser set to %s\n", req.userID)

		// Convert data to goja.Value and call callback
		jsData := u.vm.ToValue(data)
		req.callback(goja.Undefined(), jsData)

		fmt.Printf("[UI] HandleResponse: callback finished\n")
	}
}

// getUserID extracts userId from options or falls back to currentUser
func (u *UIModule) getUserID(options interface{}) string {
	// Try to get userId from options
	if opts, ok := options.(map[string]interface{}); ok {
		if uid, ok := opts["userId"].(string); ok && uid != "" {
			fmt.Printf("[UI] getUserID: found userId in options: %s\n", uid)
			return uid
		}
	}

	// Fall back to currentUser
	u.mu.Lock()
	currentUser := u.currentUser
	u.mu.Unlock()
	fmt.Printf("[UI] getUserID: falling back to currentUser: %s\n", currentUser)
	return currentUser
}

// Alert shows an alert notification (fire-and-forget)
func (u *UIModule) Alert(call goja.FunctionCall) goja.Value {
	fmt.Printf("[UI] Alert called\n")

	if len(call.Arguments) < 1 {
		panic(u.vm.NewTypeError("$ui.alert requires options argument"))
	}

	options := call.Arguments[0].Export()
	userID := u.getUserID(options)

	fmt.Printf("[UI] Alert: userID=%s, broadcaster=%v\n", userID, u.broadcaster != nil)

	if userID == "" || u.broadcaster == nil {
		// No user context or no broadcaster - silently ignore
		fmt.Printf("[UI] Alert: no user context or broadcaster, ignoring\n")
		return goja.Undefined()
	}

	fmt.Printf("[UI] Alert: sending request\n")

	// Send alert to user (fire-and-forget, no callback)
	u.broadcaster.SendUIRequest(
		u.projectID.Hex(),
		userID,
		UIRequestData{
			RequestID:  uuid.New().String(),
			DialogType: UIDialogAlert,
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

	userID := u.getUserID(options)

	if userID == "" || u.broadcaster == nil {
		// No user context - call callback with null
		callback(goja.Undefined(), goja.Null())
		return goja.Undefined()
	}

	requestID := uuid.New().String()

	u.mu.Lock()
	u.pendingReqs[requestID] = &UIRequest{
		callback: callback,
		userID:   userID,
	}
	u.mu.Unlock()

	u.broadcaster.SendUIRequest(
		u.projectID.Hex(),
		userID,
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

	userID := u.getUserID(options)

	if userID == "" || u.broadcaster == nil {
		// No user context - call callback with null
		callback(goja.Undefined(), goja.Null())
		return goja.Undefined()
	}

	requestID := uuid.New().String()

	u.mu.Lock()
	u.pendingReqs[requestID] = &UIRequest{
		callback: callback,
		userID:   userID,
	}
	u.mu.Unlock()

	u.broadcaster.SendUIRequest(
		u.projectID.Hex(),
		userID,
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

	userID := u.getUserID(options)

	if userID == "" || u.broadcaster == nil {
		// No user context - call callback with null
		callback(goja.Undefined(), goja.Null())
		return goja.Undefined()
	}

	requestID := uuid.New().String()

	u.mu.Lock()
	u.pendingReqs[requestID] = &UIRequest{
		callback: callback,
		userID:   userID,
	}
	u.mu.Unlock()

	u.broadcaster.SendUIRequest(
		u.projectID.Hex(),
		userID,
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
	u.currentUser = ""
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
					{Name: "title", Type: "string", Description: "Dialog title"},
					{Name: "text", Type: "string", Description: "Dialog message text"},
					{Name: "severity", Type: "'info' | 'success' | 'warning' | 'error'", Description: "Alert severity (default: 'info')"},
				},
			},
			{
				Name:        "ConfirmOptions",
				Description: "Options for confirm dialog",
				Fields: []schema.ParamSchema{
					{Name: "title", Type: "string", Description: "Dialog title"},
					{Name: "text", Type: "string", Description: "Dialog message text"},
					{Name: "yes", Type: "string", Description: "Yes button label (default: 'Yes')"},
					{Name: "no", Type: "string", Description: "No button label (default: 'No')"},
				},
			},
			{
				Name:        "PromptOptions",
				Description: "Options for prompt dialog",
				Fields: []schema.ParamSchema{
					{Name: "title", Type: "string", Description: "Dialog title"},
					{Name: "text", Type: "string", Description: "Dialog message text"},
					{Name: "placeholder", Type: "string", Description: "Input placeholder text"},
					{Name: "defaultValue", Type: "string", Description: "Default input value"},
				},
			},
			{
				Name:        "FormField",
				Description: "Form field definition",
				Fields: []schema.ParamSchema{
					{Name: "name", Type: "string", Description: "Field name (used as key in result)"},
					{Name: "type", Type: "'input' | 'textarea' | 'checkbox' | 'select' | 'combobox' | 'radiogroup' | 'date' | 'datetime'", Description: "Field type"},
					{Name: "label", Type: "string", Description: "Field label"},
					{Name: "hint", Type: "string", Description: "Helper text below the field"},
					{Name: "colspan", Type: "number | 'full'", Description: "Column span in 6-column grid (1-6 or 'full')"},
					{Name: "required", Type: "boolean", Description: "Whether field is required"},
					{Name: "placeholder", Type: "string", Description: "Placeholder text"},
					{Name: "defaultValue", Type: "any", Description: "Default value"},
					{Name: "options", Type: "string[] | {label: string, value: string}[]", Description: "Options for select/combobox/radiogroup"},
				},
			},
			{
				Name:        "FormAction",
				Description: "Form action button",
				Fields: []schema.ParamSchema{
					{Name: "label", Type: "string", Description: "Button label"},
					{Name: "variant", Type: "'default' | 'outline' | 'destructive'", Description: "Button variant"},
					{Name: "action", Type: "'submit' | 'cancel' | string", Description: "Action identifier"},
				},
			},
			{
				Name:        "FormOptions",
				Description: "Options for form dialog",
				Fields: []schema.ParamSchema{
					{Name: "title", Type: "string", Description: "Dialog title"},
					{Name: "text", Type: "string", Description: "Dialog description text"},
					{Name: "schema", Type: "FormField[]", Description: "Form fields schema"},
					{Name: "actions", Type: "FormAction[]", Description: "Form action buttons"},
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
