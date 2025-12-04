# UI Implementation Plan - M3M

## Overview

Реализация фронтенда для M3M - платформы управления JavaScript мини-сервисами.

## Stack

- React 19 + TypeScript
- Vite 7
- Tailwind CSS v4
- shadcn/ui (new-york style)
- React Router v7
- TanStack Query v5 (data fetching)
- Zustand (state management)
- React Hook Form + Zod (forms)
- Monaco Editor (code editing)
- TipTap (rich text)

---

## Phase 1: Foundation & Infrastructure

### 1.1 Dependencies Installation
```bash
npm install react-router-dom @tanstack/react-query zustand react-hook-form @hookform/resolvers zod @monaco-editor/react @tiptap/react @tiptap/starter-kit @tiptap/extension-placeholder date-fns
```

### 1.2 Core Structure
```
src/
├── api/                    # API client & hooks
│   ├── client.ts           # Fetch wrapper with JWT
│   ├── auth.ts             # Auth API
│   ├── projects.ts         # Projects API
│   ├── pipeline.ts         # Pipeline API
│   ├── goals.ts            # Goals API
│   ├── storage.ts          # Storage API
│   ├── models.ts           # Models API
│   ├── environment.ts      # Environment API
│   ├── runtime.ts          # Runtime API
│   └── users.ts            # Users API
├── components/
│   ├── ui/                 # shadcn components (existing)
│   ├── layout/             # Layout components
│   │   ├── app-sidebar.tsx
│   │   ├── app-header.tsx
│   │   └── app-layout.tsx
│   ├── forms/              # Reusable form components
│   │   ├── form-field.tsx
│   │   └── color-picker.tsx
│   └── shared/             # Shared components
│       ├── data-table.tsx
│       ├── confirm-dialog.tsx
│       ├── empty-state.tsx
│       └── loading-state.tsx
├── features/               # Feature modules
│   ├── auth/
│   ├── projects/
│   ├── pipeline/
│   ├── goals/
│   ├── storage/
│   ├── models/
│   ├── environment/
│   └── users/
├── hooks/                  # Custom hooks
│   ├── use-mobile.ts       # (existing)
│   ├── use-auth.ts
│   └── use-toast.ts
├── lib/
│   ├── config.ts           # (existing)
│   ├── utils.ts            # (existing)
│   └── constants.ts        # Colors, field types, etc.
├── providers/
│   ├── auth-provider.tsx
│   ├── query-provider.tsx
│   └── theme-provider.tsx
├── routes/
│   └── index.tsx           # Route definitions
├── types/
│   └── index.ts            # TypeScript types
├── App.tsx
├── main.tsx
└── index.css
```

### 1.3 API Client
- Базовый fetch wrapper с JWT токенами
- Automatic token refresh
- Error handling
- Request/response interceptors

### 1.4 Auth Provider
- JWT token storage (localStorage)
- Auto-login on app load
- Protected routes
- User context

---

## Phase 2: Authentication & Layout

### 2.1 Login Page (`/login`)
- Email/password form
- Validation with Zod
- Error handling
- Redirect to dashboard on success

### 2.2 App Layout
- Sidebar navigation (collapsible)
- Header with breadcrumbs
- User menu (profile, logout)
- Responsive design

### 2.3 Sidebar Structure
```
[Logo - M3M]
────────────
Dashboard
Projects (expandable)
  └── [project list...]
Goals
────────────
Settings
  ├── Users (admin)
  └── Profile
────────────
[User Menu]
```

---

## Phase 3: Projects Module

### 3.1 Projects List (`/projects`)
- Grid/List view toggle
- Project cards with:
  - Name, slug, color indicator
  - Status badge (running/stopped)
  - Quick actions (start/stop, settings)
- Create project button
- Search/filter

### 3.2 Create Project Modal
- Name input
- Slug input (auto-generate from name)
- Color picker (16 preset colors)
- Form validation

### 3.3 Project Dashboard (`/projects/:id`)
- Overview statistics
- Quick actions (start/stop/restart)
- Status indicator
- Recent logs preview
- Navigation tabs:
  - Pipeline (code)
  - Storage
  - Models
  - Goals
  - Environment
  - Settings

### 3.4 Project Settings (`/projects/:id/settings`)
- General settings form
- Members management
- API key display/regenerate
- Danger zone (delete project)

---

## Phase 4: Pipeline (Code Editor)

### 4.1 Pipeline Page (`/projects/:id/pipeline`)
- Monaco Editor integration
- Branch selector dropdown
- Release selector
- Save button
- Publish button

### 4.2 Branches Panel
- List branches
- Create branch (from release/branch)
- Delete branch
- Reset branch to version

### 4.3 Releases Panel
- List releases with version (X.X)
- Active release indicator
- Create release dialog:
  - Source branch
  - Version bump (minor/major)
  - Comment
  - Tag selection (stable, not-fix, night-build, develop)
- Activate release
- Delete release

### 4.4 Monaco Configuration
- TypeScript support
- Runtime API type definitions (from `/runtime/types`)
- Syntax highlighting
- Auto-completion
- Theme matching app theme

---

## Phase 5: Storage (File Manager)

### 5.1 Storage Page (`/projects/:id/storage`)
- File browser layout:
  - Breadcrumb navigation
  - Grid/List view toggle
  - File/folder list

### 5.2 File Operations
- Create folder
- Upload file (drag & drop)
- Create new file (text/json/yaml)
- Download file
- Rename file/folder
- Delete file/folder
- Generate public link

### 5.3 File Preview/Edit
- Image preview with thumbnails
- Text file editor (Monaco for code files)
- JSON viewer/editor

---

## Phase 6: Models (Database)

### 6.1 Models List (`/projects/:id/models`)
- List of models
- Create model button
- Model cards with field count

### 6.2 Model Schema Editor (`/projects/:id/models/:id/schema`)
- Fields list with drag reorder
- Add field form:
  - Key
  - Type (String, Text, Number, Float, Bool, Document, File, Ref, Date, DateTime)
  - Required toggle
  - Default value (type-dependent)
- Table view configuration
- Form view configuration

### 6.3 Model Data View (`/projects/:id/models/:id/data`)
- Data table with:
  - Configurable columns
  - Sorting
  - Filtering
  - Pagination
- CRUD operations:
  - Create record (form based on schema)
  - View record
  - Edit record
  - Delete record

### 6.4 Field Type Components
- String: Input
- Text: Textarea or TipTap
- Number/Float: Number input
- Bool: Checkbox/Switch
- Document: JSON editor
- File: File picker (from storage)
- Ref: Select with search
- Date: Date picker
- DateTime: DateTime picker

---

## Phase 7: Goals & Analytics

### 7.1 Goals Page (`/projects/:id/goals`)
- Goals list
- Create goal form:
  - Name
  - Slug
  - Color
  - Type (counter / daily_counter)
  - Description

### 7.2 Global Goals (`/goals`)
- Global goals list (admin)
- Link to projects

### 7.3 Goal Statistics
- Statistics page with charts
- Date range picker
- Goal selector
- Counter display
- Daily breakdown (for daily_counter)

---

## Phase 8: Environment

### 8.1 Environment Page (`/projects/:id/environment`)
- Key-value list
- Add variable form:
  - Key
  - Type (String, Text, Json, Integer, Float, Boolean)
  - Value (type-dependent input)
- Edit inline
- Delete with confirmation

---

## Phase 9: Runtime & Logs

### 9.1 Runtime Controls
- Start/Stop/Restart buttons
- Status indicator
- Release selector for start

### 9.2 Logs Viewer (`/projects/:id/logs`)
- Log display area (terminal-like)
- Auto-scroll toggle
- Download logs button
- Clear button
- Severity filter

### 9.3 Monitor Dashboard
- Runtime statistics
- Resource usage
- Request metrics

---

## Phase 10: User Management

### 10.1 Profile Page (`/settings/profile`)
- Update name
- Update avatar
- Change password form

### 10.2 Users Page (`/settings/users`) - Admin
- Users list table
- Create user form:
  - Email
  - Password
  - Name
  - Permissions:
    - Create projects
    - Manage users
    - Project access list
- Edit user
- Block/Unblock user
- Delete user

---

## Phase 11: Additional Components

### 11.1 shadcn Components to Install
```bash
npx shadcn@latest add table dialog alert-dialog select textarea switch checkbox tabs badge form label toast sonner popover command calendar scroll-area context-menu alert progress slider radio-group toggle toggle-group aspect-ratio resizable
```

### 11.2 Custom Components
- ColorPicker (16 preset colors)
- StatusBadge (running/stopped/error)
- LogViewer (terminal-like)
- FileIcon (by mime type)
- CodeEditor (Monaco wrapper)
- RichTextEditor (TipTap wrapper)
- DataTable (generic with sorting/filtering/pagination)
- ConfirmDialog (delete confirmations)
- EmptyState (no data placeholders)

---

## Implementation Order

1. **Phase 1**: Foundation (deps, structure, API client)
2. **Phase 2**: Auth & Layout
3. **Phase 3**: Projects (list, create, dashboard)
4. **Phase 4**: Pipeline (editor, branches, releases)
5. **Phase 5**: Storage (file manager)
6. **Phase 6**: Models (schema, data CRUD)
7. **Phase 7**: Goals
8. **Phase 8**: Environment
9. **Phase 9**: Runtime & Logs
10. **Phase 10**: User Management
11. **Phase 11**: Polish & Testing

---

## UI/UX Guidelines

### Colors (16 preset)
```typescript
const COLORS = [
  { name: 'slate', value: '#64748b' },
  { name: 'gray', value: '#6b7280' },
  { name: 'zinc', value: '#71717a' },
  { name: 'red', value: '#ef4444' },
  { name: 'orange', value: '#f97316' },
  { name: 'amber', value: '#f59e0b' },
  { name: 'yellow', value: '#eab308' },
  { name: 'lime', value: '#84cc16' },
  { name: 'green', value: '#22c55e' },
  { name: 'emerald', value: '#10b981' },
  { name: 'teal', value: '#14b8a6' },
  { name: 'cyan', value: '#06b6d4' },
  { name: 'sky', value: '#0ea5e9' },
  { name: 'blue', value: '#3b82f6' },
  { name: 'indigo', value: '#6366f1' },
  { name: 'violet', value: '#8b5cf6' },
  { name: 'purple', value: '#a855f7' },
  { name: 'fuchsia', value: '#d946ef' },
  { name: 'pink', value: '#ec4899' },
  { name: 'rose', value: '#f43f5e' },
]
```

### Form Patterns
- Use shadcn Field components (FieldGroup, Field, FieldLabel, FieldDescription)
- Zod validation
- Error messages below inputs
- Loading states on submit

### Data Tables
- Sortable columns
- Filterable
- Pagination (10/25/50/100)
- Bulk actions where applicable

### Confirmations
- AlertDialog for destructive actions
- Toast notifications for success/error

### Responsive
- Mobile-first approach
- Collapsible sidebar
- Stacked layouts on mobile

---

## API Types

```typescript
// User
interface User {
  id: string;
  email: string;
  name: string;
  avatar?: string;
  isRoot: boolean;
  isBlocked: boolean;
  permissions: {
    createProjects: boolean;
    manageUsers: boolean;
    projectAccess: string[]; // project IDs
  };
  createdAt: string;
  updatedAt: string;
}

// Project
interface Project {
  id: string;
  name: string;
  slug: string;
  color?: string;
  status: 'running' | 'stopped';
  apiKey: string;
  ownerID: string;
  members: string[];
  activeRelease?: string;
  createdAt: string;
  updatedAt: string;
}

// Branch
interface Branch {
  id: string;
  projectID: string;
  name: string;
  code: string;
  createdAt: string;
  updatedAt: string;
}

// Release
interface Release {
  id: string;
  projectID: string;
  version: string;
  code: string;
  comment?: string;
  tag?: string;
  isActive: boolean;
  createdAt: string;
}

// Goal
interface Goal {
  id: string;
  name: string;
  slug: string;
  color?: string;
  type: 'counter' | 'daily_counter';
  description?: string;
  projectID?: string; // null = global
  projectAccess: string[];
  createdAt: string;
  updatedAt: string;
}

// Environment
interface Environment {
  id: string;
  projectID: string;
  key: string;
  type: 'string' | 'text' | 'json' | 'integer' | 'float' | 'boolean';
  value: string;
}

// Model
interface Model {
  id: string;
  projectID: string;
  name: string;
  slug: string;
  fields: ModelField[];
  tableConfig?: TableConfig;
  formConfig?: FormConfig;
  createdAt: string;
  updatedAt: string;
}

interface ModelField {
  key: string;
  type: FieldType;
  required: boolean;
  defaultValue?: any;
}

type FieldType = 'string' | 'text' | 'number' | 'float' | 'bool' | 'document' | 'file' | 'ref' | 'date' | 'datetime';
```

---

## File List (Implementation)

### Core Files
- [ ] `src/api/client.ts`
- [ ] `src/api/auth.ts`
- [ ] `src/api/projects.ts`
- [ ] `src/api/pipeline.ts`
- [ ] `src/api/goals.ts`
- [ ] `src/api/storage.ts`
- [ ] `src/api/models.ts`
- [ ] `src/api/environment.ts`
- [ ] `src/api/runtime.ts`
- [ ] `src/api/users.ts`
- [ ] `src/providers/auth-provider.tsx`
- [ ] `src/providers/query-provider.tsx`
- [ ] `src/routes/index.tsx`
- [ ] `src/types/index.ts`
- [ ] `src/lib/constants.ts`

### Layout
- [ ] `src/components/layout/app-layout.tsx`
- [ ] `src/components/layout/app-sidebar.tsx`
- [ ] `src/components/layout/app-header.tsx`

### Shared Components
- [ ] `src/components/shared/data-table.tsx`
- [ ] `src/components/shared/confirm-dialog.tsx`
- [ ] `src/components/shared/empty-state.tsx`
- [ ] `src/components/shared/loading-state.tsx`
- [ ] `src/components/shared/color-picker.tsx`
- [ ] `src/components/shared/status-badge.tsx`
- [ ] `src/components/shared/code-editor.tsx`
- [ ] `src/components/shared/log-viewer.tsx`

### Feature Pages
- [ ] `src/features/auth/login-page.tsx`
- [ ] `src/features/projects/projects-page.tsx`
- [ ] `src/features/projects/project-dashboard.tsx`
- [ ] `src/features/projects/project-settings.tsx`
- [ ] `src/features/pipeline/pipeline-page.tsx`
- [ ] `src/features/storage/storage-page.tsx`
- [ ] `src/features/models/models-page.tsx`
- [ ] `src/features/models/model-schema-page.tsx`
- [ ] `src/features/models/model-data-page.tsx`
- [ ] `src/features/goals/goals-page.tsx`
- [ ] `src/features/environment/environment-page.tsx`
- [ ] `src/features/users/users-page.tsx`
- [ ] `src/features/users/profile-page.tsx`
