# M3M Updates Plan

## Analysis Summary

Analyzed `issue.md` requirements against current implementation. Core functionality is mostly in place, but several features are missing and comprehensive testing is needed.

---

## 1. Testing Coverage (HIGH PRIORITY)

### 1.1 Runtime Modules Tests
- [x] `utils` module tests (utils_test.go)
  - Random functions, slugify, truncate, capitalize
  - Regex operations, date formatting/parsing
  - Edge cases: empty strings, unicode, invalid patterns
- [x] `crypto` module tests (crypto_test.go)
  - MD5, SHA256 hashing
  - RandomBytes generation
  - Determinism, output lengths
- [x] `encoding` module tests (encoding_test.go)
  - Base64 encode/decode
  - JSON parse/stringify
  - URL encode/decode
  - Roundtrip tests, invalid input handling
- [x] `http` module tests (http_test.go)
  - GET, POST, PUT, DELETE methods
  - Headers, timeouts, status codes
  - Error handling (invalid URLs, connection refused)
- [x] `storage` module tests (storage_test.go)
  - Read, write, exists, delete
  - Directory operations
  - Edge cases: empty, large, binary content
- [x] `router` module tests (router_test.go)
  - Route registration and matching
  - Path params, query params, headers
  - Multiple routes, method handling
- [x] `schedule` module tests (schedule_test.go)
  - Daily, hourly, cron job registration
  - Start/stop lifecycle
  - Invalid expressions, error handling
- [x] `service` lifecycle tests (service_test.go)
  - boot, start, shutdown callbacks
  - Execution order
  - Error handling, timeout
- [x] `database` module tests (database_test.go)
  - Basic structure tests
  - MongoDB integration (skipped without DB)
- [x] `image` module tests (image_test.go)
  - Info, resize, crop, thumbnail operations
  - Non-existent files, invalid images
  - Format conversion, base64 encoding
- [x] `smtp` module tests (smtp_test.go)
  - Message building, headers
  - Error handling (missing config)
  - Options handling (CC, BCC, HTML)

### 1.2 Service Layer Tests
- [ ] user_service_test.go
- [ ] project_service_test.go
- [ ] goal_service_test.go
- [ ] pipeline_service_test.go
- [ ] environment_service_test.go

---

## 2. Missing Runtime Modules

### 2.1 Image Module (from issue.md) - COMPLETED
Location: `internal/runtime/modules/image.go`
Functions:
- [x] `resize(path, width, height)` - Resize image
- [x] `resizeKeepRatio(path, maxWidth, maxHeight)` - Resize keeping aspect ratio
- [x] `crop(path, x, y, width, height)` - Crop image
- [x] `thumbnail(path, size)` - Generate square thumbnail
- [x] `info(path)` - Get image dimensions and format
- [x] `readAsBase64(path)` - Read image as data URI

### 2.2 SMTP Module (from issue.md) - COMPLETED
Location: `internal/runtime/modules/smtp.go`
Functions:
- [x] `send(to, subject, body, options)` - Send email
- [x] `sendHTML(to, subject, body)` - Send HTML email
- [x] Configuration via environment variables (SMTP_HOST, SMTP_PORT, SMTP_USER, SMTP_PASS, SMTP_FROM)

### 2.3 Draw Module (from issue.md)
Location: `internal/runtime/modules/draw.go`
Functions:
- [ ] Canvas operations for image drawing/generation

---

## 3. Missing Features

### 3.1 User Management - COMPLETED
- [x] Add `IsBlocked` field to User domain
- [x] Block/unblock user service methods
- [x] Block/unblock user endpoints (POST /users/:id/block, /users/:id/unblock)
- [x] Blocked user cannot login (403 Forbidden)

### 3.2 TypeScript Definitions - COMPLETED
- [x] Add `service` module to types.go
- [x] Add `image` module to types.go
- [x] Add `smtp` module to types.go

### 3.3 Monaco Type Definitions for Models
- [ ] Generate TypeScript definitions from project models
- [ ] Include in runtime type definitions endpoint

---

## 4. Implementation Order

### Phase 1: Complete Testing - COMPLETED
1. [x] Runtime modules tests
2. [x] Service lifecycle tests
3. [x] Database module tests (basic)
4. [x] Image module tests
5. [x] SMTP module tests

### Phase 2: Missing Modules - COMPLETED
1. [x] Image module
2. [x] SMTP module
3. [ ] Draw module (lower priority)

### Phase 3: User Features - COMPLETED
1. [x] User block/unblock
2. [x] TypeScript definitions update

### Phase 4: Frontend
- [ ] Monaco Editor integration
- [ ] TipTap rich text editor
- [ ] UI components per ui-style.md

---

## Test Execution

```bash
# Run all runtime module tests
go test ./internal/runtime/modules/... -v

# Run specific test
go test ./internal/runtime/modules/... -run TestRouterModule -v

# Run all tests
make test
```

---

## Summary of Completed Work

### Backend Implementation
1. **Runtime Module Tests**: Comprehensive tests for all modules (utils, crypto, encoding, http, storage, router, schedule, service, database, image, smtp)
2. **Image Module**: Full implementation with resize, crop, thumbnail, info, and base64 encoding
3. **SMTP Module**: Email sending with HTML support, configurable via environment variables
4. **User Blocking**: Complete flow with domain, service, and handler layers

### Files Modified/Created
- `internal/runtime/modules/image.go` - New image processing module
- `internal/runtime/modules/image_test.go` - Image module tests
- `internal/runtime/modules/smtp.go` - New SMTP email module
- `internal/runtime/modules/smtp_test.go` - SMTP module tests
- `internal/runtime/modules/types.go` - Added TypeScript definitions for service, image, smtp
- `internal/runtime/runtime.go` - Registered image and smtp modules
- `internal/domain/user.go` - Added IsBlocked field
- `internal/service/user_service.go` - Added Block, Unblock, IsBlocked methods
- `internal/service/auth_service.go` - Added ErrUserBlocked, blocked user login check
- `internal/handler/user_handler.go` - Added Block, Unblock endpoints
- `internal/handler/auth_handler.go` - Handle blocked user error on login

### Remaining Work
1. Service layer tests (user, project, goal, pipeline, environment)
2. Draw module (canvas operations)
3. Dynamic model TypeScript definitions
4. Frontend implementation (Monaco, TipTap, UI)
