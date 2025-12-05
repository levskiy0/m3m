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
  tableConfig?: TableConfig;
  formConfig?: FormConfig;
  createdAt: string;
  updatedAt: string;
}

export interface ModelField {
  key: string;
  type: FieldType;
  required: boolean;
  defaultValue?: unknown;
  refModel?: string;
  options?: FieldOptions;
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
  | 'datetime';

export interface TableConfig {
  columns: string[];
  sortableColumns: string[];
  filterableColumns: string[];
  defaultSort?: {
    column: string;
    direction: 'asc' | 'desc';
  };
}

export interface FormConfig {
  fields: FormFieldConfig[];
}

export interface FormFieldConfig {
  key: string;
  hidden?: boolean;
  order?: number;
  widget?: 'input' | 'textarea' | 'richtext' | 'select' | 'checkbox' | 'datepicker' | 'filepicker';
}

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
  tableConfig?: TableConfig;
  formConfig?: FormConfig;
}

export interface ModelData {
  _id: string;
  [key: string]: unknown;
  createdAt: string;
  updatedAt: string;
}

export interface QueryDataRequest {
  filter?: Record<string, unknown>;
  sort?: Record<string, 1 | -1>;
  page?: number;
  limit?: number;
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
  isDir: boolean;
  size: number;
  mimeType?: string;
  modTime: string;
  thumbnail?: string;
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

export interface GenerateLinkRequest {
  path: string;
  expiresIn?: number;
}

export interface PublicLink {
  url: string;
  expiresAt: string;
}

// Runtime types
export interface RuntimeStatus {
  status: ProjectStatus;
  uptime?: number;
  startedAt?: string;
  activeRelease?: string;
}

export interface RuntimeStats {
  requests: number;
  errors: number;
  avgResponseTime: number;
  memoryUsage: number;
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
