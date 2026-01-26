export interface Crash {
  id: string
  app_id: string
  app_version: string
  platform: string
  os_version: string
  device_model: string
  error_type: string
  error_message: string
  stack_trace: StackFrame[]
  fingerprint: string
  group_id: string
  user_id?: string
  environment: string
  created_at: string
  log_file_path?: string
  metadata?: Record<string, any>
  breadcrumbs?: Breadcrumb[]
}

export interface StackFrame {
  file_name: string
  line_number: number
  column_number?: number
  method_name: string
  class_name?: string
  native?: boolean
}

export interface Breadcrumb {
  timestamp: string
  type: string
  category: string
  message: string
  data?: Record<string, any>
  level: string
}

export interface CrashGroup {
  id: string
  app_id: string
  fingerprint: string
  error_type: string
  error_message: string
  first_seen: string
  last_seen: string
  occurrence_count: number
  status: 'open' | 'resolved' | 'ignored'
  assigned_to?: string
  notes?: string
}

export interface App {
  id: string
  name: string
  api_key?: string
  created_at: string
  retention_days: number
}

export interface Alert {
  id: string
  app_id: string
  type: 'webhook' | 'email' | 'slack'
  config: Record<string, any>
  enabled: boolean
  created_at: string
}

export interface CrashStats {
  app_id: string
  total_crashes: number
  total_groups: number
  open_groups: number
  crashes_last_24h: number
  crashes_last_7d: number
  crashes_last_30d: number
  top_errors: ErrorSummary[]
  crash_trend: TrendPoint[]
}

export interface ErrorSummary {
  group_id: string
  error_type: string
  error_message: string
  count: number
}

export interface TrendPoint {
  date: string
  count: number
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  limit: number
  offset: number
}
