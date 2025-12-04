# M3M Updates Plan

## Implementation Status

| Task | Status |
|------|--------|
| Add delayed_test.go | DONE |
| Add logger_test.go | DONE |
| Add env_test.go | DONE |
| Add goals_test.go | DONE |
| Extend env module (Has, Keys, GetString, GetInt, GetFloat, GetBool, GetAll) | DONE |
| Extend goals module (GetValue, GetStats, List, Get) | DONE |
| Implement draw module | DONE |
| Fix goals module nil check bug | DONE |

## Analysis Summary

After analyzing the codebase against `issue.md` requirements, the following gaps and improvements have been identified.

---

## 1. Missing Runtime Modules

### 1.1 Draw Module (HIGH Priority)
The `issue.md` specifies a `draw` module for canvas/drawing operations, but only `image` module exists (image processing only).

**Required functionality:**
- `draw.createCanvas(width, height)` - Create a new canvas
- `draw.loadImage(path)` - Load image onto canvas
- `draw.rect(x, y, width, height, color)` - Draw rectangle
- `draw.circle(x, y, radius, color)` - Draw circle
- `draw.line(x1, y1, x2, y2, color, width)` - Draw line
- `draw.text(text, x, y, options)` - Draw text with font/size/color options
- `draw.save(canvas, path)` - Save canvas to file
- `draw.toBase64(canvas)` - Export canvas as base64

**Implementation:** Use `golang.org/x/image/draw` and `github.com/fogleman/gg` for drawing operations.

---

## 2. Missing Tests (HIGH Priority)

The following runtime modules lack test coverage:

### 2.1 delayed_test.go
**Test cases:**
- Basic task execution
- Worker pool semaphore limits (concurrent task limiting)
- Panic recovery in handler
- Zero/negative pool size handling
- Multiple concurrent tasks
- Task completion verification

### 2.2 logger_test.go
**Test cases:**
- Log file creation
- Fallback to stdout on file error
- Log format verification (timestamp, level, message)
- All log levels (debug, info, warn, error)
- Concurrent logging thread safety
- File close behavior
- Large message handling

### 2.3 env_test.go
**Test cases:**
- Get existing key
- Get non-existing key (returns nil)
- Nil vars map handling
- Various value types (string, int, bool, json)
- Empty string key

### 2.4 goals_test.go
**Test cases:**
- Increment with default value (1)
- Increment with custom value
- Increment non-existing goal (error handling)
- Multiple increments
- Zero/negative increment values

---

## 3. Runtime Module Enhancements

### 3.1 Delayed Module Enhancements
Current implementation is minimal. Add:
```go
// New methods
func (d *DelayedModule) RunWithTimeout(handler, timeout) bool
func (d *DelayedModule) GetPoolSize() int
func (d *DelayedModule) GetActiveCount() int
```

### 3.2 Logger Module Enhancements
Add ability to format log output and structured logging:
```go
// Additional log methods
func (l *LoggerModule) Printf(format string, args ...interface{})
func (l *LoggerModule) WithFields(fields map[string]interface{}) *LoggerModule
```

### 3.3 Env Module Enhancements
Add helper methods:
```go
func (e *EnvModule) Has(key string) bool
func (e *EnvModule) Keys() []string
func (e *EnvModule) GetString(key string, defaultValue string) string
func (e *EnvModule) GetInt(key string, defaultValue int) int
func (e *EnvModule) GetBool(key string, defaultValue bool) bool
```

### 3.4 Goals Module Enhancements
Add read operations:
```go
func (g *GoalsModule) GetValue(slug string) int64
func (g *GoalsModule) GetStats(slug string, days int) []GoalStat
func (g *GoalsModule) List() []Goal
```

### 3.5 Storage Module Enhancements
Add missing operations from `issue.md`:
```go
func (s *StorageModule) Copy(src, dst string) bool
func (s *StorageModule) Move(src, dst string) bool
func (s *StorageModule) GetSize(path string) int64
func (s *StorageModule) GetMimeType(path string) string
func (s *StorageModule) ReadBinary(path string) []byte
func (s *StorageModule) WriteBinary(path string, data []byte) bool
```

### 3.6 HTTP Module Enhancements
Add flexible request method:
```go
func (h *HTTPModule) Request(method, url string, options HTTPOptions) HTTPResponse
func (h *HTTPModule) Head(url string, options HTTPOptions) HTTPResponse
func (h *HTTPModule) Patch(url string, body interface{}, options HTTPOptions) HTTPResponse
```

### 3.7 Crypto Module Enhancements
Add HMAC and additional algorithms:
```go
func (c *CryptoModule) HmacSHA256(data, key string) string
func (c *CryptoModule) HmacMD5(data, key string) string
func (c *CryptoModule) SHA512(data string) string
```

### 3.8 Database Module Enhancements
Add convenience methods:
```go
func (c *CollectionModule) FindById(id string) map[string]interface{}
func (c *CollectionModule) Upsert(filter, data map[string]interface{}) bool
func (c *CollectionModule) UpdateMany(filter, data map[string]interface{}) int
func (c *CollectionModule) DeleteMany(filter map[string]interface{}) int
```

---

## 4. Plugin System Fixes

### 4.1 Plugin Config Loading (TODO in code)
Location: `internal/plugin/loader.go:122`

Currently plugins are loaded with empty config. Need to:
1. Read plugin configs from `config.yaml`
2. Pass config to plugin `Initialize()` method

**Config structure:**
```yaml
plugins:
  path: "./plugins"
  config:
    telegram:
      bot_token: "..."
    custom_plugin:
      setting1: "value"
```

---

## 5. Type Definitions Updates

Update `internal/runtime/modules/types.go` to include:
- Draw module TypeScript definitions
- New methods for enhanced modules

---

## Implementation Order

### Phase 1: Tests for Existing Modules (Priority: HIGH)
1. `delayed_test.go` - Test worker pool behavior
2. `logger_test.go` - Test logging functionality
3. `env_test.go` - Test environment access
4. `goals_test.go` - Test metrics tracking

### Phase 2: Module Enhancements (Priority: MEDIUM)
1. Env module enhancements (Has, Keys, GetString, GetInt, GetBool)
2. Goals module enhancements (GetValue, GetStats, List)
3. Delayed module enhancements (timeout, active count)
4. Storage module enhancements (Copy, Move, GetSize)
5. HTTP module enhancements (Request, Patch, Head)
6. Crypto module enhancements (HMAC)
7. Database module enhancements (FindById, Upsert, bulk operations)

### Phase 3: New Modules (Priority: MEDIUM)
1. Draw module implementation

### Phase 4: Infrastructure Fixes (Priority: LOW)
1. Plugin config loading fix

---

## Test Strategy

### Runtime Module Tests
Each module test should cover:
1. **Happy path** - Normal operation with valid inputs
2. **Edge cases** - Empty inputs, nil values, boundary values
3. **Error handling** - Invalid inputs, failures, exceptions
4. **Concurrency** - Thread safety where applicable
5. **Resource cleanup** - Proper cleanup of resources

### Test Pattern (from existing tests):
```go
func TestModule_Method(t *testing.T) {
    module := NewModule(...)

    tests := []struct {
        name     string
        input    InputType
        expected OutputType
        wantErr  bool
    }{
        {"happy path", validInput, expectedOutput, false},
        {"empty input", "", "", false},
        {"error case", invalidInput, nil, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := module.Method(tt.input)
            if got != tt.expected {
                t.Errorf("Method(%v) = %v, want %v", tt.input, got, tt.expected)
            }
        })
    }
}
```

---

## Files to Create/Modify

### New Files:
- `internal/runtime/modules/delayed_test.go`
- `internal/runtime/modules/logger_test.go`
- `internal/runtime/modules/env_test.go`
- `internal/runtime/modules/goals_test.go`
- `internal/runtime/modules/draw.go`
- `internal/runtime/modules/draw_test.go`

### Files to Modify:
- `internal/runtime/modules/delayed.go` - Add enhancements
- `internal/runtime/modules/env.go` - Add helper methods
- `internal/runtime/modules/goals.go` - Add read operations
- `internal/runtime/modules/storage.go` - Add file operations
- `internal/runtime/modules/http.go` - Add methods
- `internal/runtime/modules/crypto.go` - Add HMAC
- `internal/runtime/modules/database.go` - Add convenience methods
- `internal/runtime/modules/types.go` - Update TypeScript definitions
- `internal/runtime/runtime.go` - Register new draw module
- `internal/plugin/loader.go` - Fix config loading
- `internal/config/config.go` - Add plugin config structure
