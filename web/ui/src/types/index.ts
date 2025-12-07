// User types
export interface User {
  id: string;
  email: string;
  name: string;
  avatar?: string;
  isRoot: boolean;
  isBlocked: boolean;
  permissions: UserPermissions;
  createdAt: string;
  updatedAt: string;
}

export interface UserPermissions {
  createProjects: boolean;
  manageUsers: boolean;
  projectAccess: string[];
}

export interface CreateUserRequest {
  email: string;
  password: string;
  name: string;
  permissions: UserPermissions;
}

export interface UpdateUserRequest {
  name?: string;
  permissions?: UserPermissions;
}

export interface UpdateMeRequest {
  name?: string;
}

export interface ChangePasswordRequest {
  currentPassword: string;
  newPassword: string;
}

// Project types
export interface Project {
  id: string;
  name: string;
  slug: string;
  color?: string;
  status: ProjectStatus;
  apiKey: string;
  ownerID: string;
  members: string[];
  activeRelease?: string;
  runningSource?: string; // "release:<version>" or "debug:<branch>"
  createdAt: string;
  updatedAt: string;
}

export type ProjectStatus = 'running' | 'stopped';

export interface CreateProjectRequest {
  name: string;
  slug: string;
  color?: string;
}

export interface UpdateProjectRequest {
  name?: string;
  slug?: string;
  color?: string;
}

// Pipeline types
export interface Branch {
  id: string;
  projectID: string;
  name: string;
  code: string;
  parentBranch?: string;
  createdFromRelease?: string;
  createdAt: string;
  updatedAt: string;
}

// Lightweight version without code for list operations
export interface BranchSummary {
  id: string;
  projectID: string;
  name: string;
  parentBranch?: string;
  createdFromRelease?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateBranchRequest {
  name: string;
  sourceType: 'branch' | 'release';
  sourceName: string;
}

export interface UpdateBranchRequest {
  code: string;
}

export interface ResetBranchRequest {
  target_version: string;
}

export interface Release {
  id: string;
  projectID: string;
  version: string;
  code: string;
  comment?: string;
  tag?: ReleaseTag;
  isActive: boolean;
  createdAt: string;
}

// Lightweight version without code for list operations
export interface ReleaseSummary {
  id: string;
  projectID: string;
  version: string;
  comment?: string;
  tag?: ReleaseTag;
  isActive: boolean;
  createdAt: string;
}

export type ReleaseTag = 'stable' | 'hot-fix' | 'night-build' | 'develop';

export interface CreateReleaseRequest {
  branch_name: string;
  bump_type: 'minor' | 'major';
  comment?: string;
  tag: ReleaseTag;
}

// Goal types
export interface Goal {
  id: string;
  name: string;
  slug: string;
  color?: string;
  type: GoalType;
  description?: string;
  projectID?: string;
  projectAccess: string[];
  createdAt: string;
  updatedAt: string;
}

export type GoalType = 'counter' | 'daily_counter';

// Widget types
export interface Widget {
  id: string;
  projectId: string;
  goalId: string;
  variant: WidgetVariant;
  order: number;
  createdAt: string;
  updatedAt: string;
}

export type WidgetVariant = 'mini' | 'detailed' | 'simple';

export interface CreateWidgetRequest {
  goalId: string;
  variant: WidgetVariant;
}

export interface UpdateWidgetRequest {
  variant?: WidgetVariant;
  order?: number;
}

export interface ReorderWidgetsRequest {
  widgetIds: string[];
}

export interface CreateGoalRequest {
  name: string;
  slug: string;
  color?: string;
  type: GoalType;
  description?: string;
  projectAccess?: string[];
}

export interface UpdateGoalRequest {
  name?: string;
  color?: string;
  description?: string;
  projectAccess?: string[];
}

export interface GoalStats {
  goalID: string;
  value: number;
  dailyStats?: DailyGoalStat[];
}

export interface DailyGoalStat {
  date: string;
  value: number;
}

// Environment types
export interface Environment {
  id: string;
  projectID: string;
  key: string;
  type: EnvType;
  value: string;
}

export type EnvType = 'string' | 'text' | 'json' | 'integer' | 'float' | 'boolean';

export interface CreateEnvRequest {
  key: string;
  type: EnvType;
  value: string;
}

export interface UpdateEnvRequest {
  type?: EnvType;
  value?: string;
}

// Model types
export interface Model {
  id: string;
  projectID: string;
  name: string;
  slug: string;
  fields: ModelField[];
  table_config?: TableConfig;
  form_config?: FormConfig;
  createdAt: string;
  updatedAt: string;
}

export interface ModelField {
  key: string;
  type: FieldType;
  required: boolean;
  default_value?: unknown;
  ref_model?: string;
  options?: string[];
}

export interface FieldOptions {
  min?: number;
  max?: number;
  pattern?: string;
  enum?: string[];
}

export type FieldType =
  | 'string'
  | 'text'
  | 'number'
  | 'float'
  | 'bool'
  | 'document'
  | 'file'
  | 'ref'
  | 'date'
  | 'datetime'
  | 'select'
  | 'multiselect';

export interface TableConfig {
  columns: string[];
  filters: string[];
  sort_columns: string[];
  searchable?: string[];
}

export interface FormConfig {
  field_order: string[];
  hidden_fields: string[];
  field_views: Record<string, FieldView>;
}

export type FieldView =
  | 'input'
  | 'select'
  | 'multiselect'
  | 'textarea'
  | 'tiptap'
  | 'markdown'
  | 'slider'
  | 'checkbox'
  | 'switch'
  | 'datepicker'
  | 'datetimepicker'
  | 'file'
  | 'image'
  | 'combobox'
  | 'json';

export interface CreateModelRequest {
  name: string;
  slug: string;
  fields: ModelField[];
  tableConfig?: TableConfig;
  formConfig?: FormConfig;
}

export interface UpdateModelRequest {
  name?: string;
  fields?: ModelField[];
  table_config?: TableConfig;
  form_config?: FormConfig;
}

export interface ModelData {
  _id: string;
  [key: string]: unknown;
  createdAt: string;
  updatedAt: string;
}

export interface QueryDataRequest {
  page?: number;
  limit?: number;
  sort?: string;
  order?: string;
  filters?: FilterCondition[];
  search?: string;
  searchIn?: string[];
}

export interface FilterCondition {
  field: string;
  operator: 'eq' | 'ne' | 'gt' | 'gte' | 'lt' | 'lte' | 'contains' | 'startsWith' | 'endsWith' | 'in' | 'notIn';
  value: unknown;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
  totalPages: number;
}

// Storage types
export interface StorageItem {
  name: string;
  path: string;
  is_dir: boolean;
  size: number;
  mime_type?: string;
  updated_at: string;
  url?: string;
  download_url?: string;
  thumb_url?: string;
}

export interface CreateDirRequest {
  path: string;
  name: string;
}

export interface RenameRequest {
  path: string;
  newName: string;
}

export interface CreateFileRequest {
  path: string;
  name: string;
  content: string;
}

// Runtime types
export interface RuntimeStatus {
  status: ProjectStatus;
  uptime?: number;
  startedAt?: string;
  activeRelease?: string;
}

export interface SparklineData {
  memory: number[];   // MB values
  requests: number[]; // Request counts per interval
  jobs?: number[];    // Scheduled jobs execution counts
  cpu?: number[];     // CPU percent values
}

export interface RuntimeStats {
  project_id: string;
  status: string;
  started_at: string;
  uptime_seconds: number;
  uptime_formatted: string;
  routes_count: number;
  routes_by_method: Record<string, number>;
  scheduled_jobs: number;
  scheduler_active: boolean;
  memory: {
    alloc: number;
    total_alloc: number;
    sys: number;
    num_gc: number;
  };
  total_requests: number;
  hits_by_path: Record<string, number>;
  history?: SparklineData;
  // Extended stats (may not be available on all backends)
  storage_bytes?: number;
  database_bytes?: number;
  cpu_percent?: number;
}

export interface LogEntry {
  timestamp: string;
  level: 'debug' | 'info' | 'warn' | 'error';
  message: string;
}

// Auth types
export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  user: User;
}

// API response types
export interface ApiError {
  error: string;
  message?: string;
  code?: string;
}
