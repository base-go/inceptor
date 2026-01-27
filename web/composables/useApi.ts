import type { Crash, CrashGroup, App, Alert, CrashStats, PaginatedResponse } from '~/types'

interface AuthStatus {
  needs_password_change: boolean
}

interface LoginResponse {
  token: string
  expires_at: string
  needs_password_change: boolean
}

export const useApi = () => {
  const token = useState<string>('authToken', () => '')
  const needsPasswordChange = useState<boolean>('needsPasswordChange', () => false)

  // Always use relative path - this works for embedded dashboard
  const baseUrl = '/api/v1'

  const headers = computed(() => {
    const h: Record<string, string> = {
      'Content-Type': 'application/json',
    }
    if (token.value) {
      h['Authorization'] = `Bearer ${token.value}`
    }
    return h
  })

  const isAuthenticated = computed(() => !!token.value)

  // Helper to handle 401 errors - clears token to show login
  const handleUnauthorized = () => {
    token.value = ''
    needsPasswordChange.value = false
    if (process.client) {
      localStorage.removeItem('inceptor_token')
      // Navigate to home which shows login when not authenticated
      navigateTo('/')
    }
  }

  // Wrapper for authenticated API calls that handles 401
  const authFetch = async <T>(url: string, options: RequestInit = {}): Promise<T> => {
    try {
      return await $fetch<T>(url, {
        ...options,
        headers: { ...headers.value, ...(options.headers as Record<string, string> || {}) },
      } as any)
    } catch (error: any) {
      // Check for 401 in various error formats
      const status = error?.response?.status || error?.status || error?.statusCode || error?.data?.statusCode
      if (status === 401) {
        handleUnauthorized()
      }
      throw error
    }
  }

  // Auth methods
  const checkAuthStatus = async (): Promise<AuthStatus> => {
    return await $fetch<AuthStatus>(`${baseUrl}/auth/status`)
  }

  const login = async (password: string): Promise<LoginResponse> => {
    const response = await $fetch<LoginResponse>(`${baseUrl}/auth/login`, {
      method: 'POST',
      body: { password },
    })
    token.value = response.token
    needsPasswordChange.value = response.needs_password_change
    if (process.client) {
      localStorage.setItem('inceptor_token', response.token)
    }
    return response
  }

  const logout = async (): Promise<void> => {
    try {
      await $fetch(`${baseUrl}/auth/logout`, {
        method: 'POST',
        headers: headers.value,
      })
    } catch {}
    token.value = ''
    needsPasswordChange.value = false
    if (process.client) {
      localStorage.removeItem('inceptor_token')
    }
  }

  const changePassword = async (oldPassword: string, newPassword: string): Promise<void> => {
    await $fetch(`${baseUrl}/auth/change-password`, {
      method: 'POST',
      headers: headers.value,
      body: { old_password: oldPassword, new_password: newPassword },
    })
    needsPasswordChange.value = false
  }

  const loadToken = () => {
    if (process.client) {
      const savedToken = localStorage.getItem('inceptor_token')
      if (savedToken) {
        token.value = savedToken
      }
    }
  }

  // Crashes
  const getCrashes = async (params?: {
    app_id?: string
    group_id?: string
    platform?: string
    environment?: string
    search?: string
    limit?: number
    offset?: number
  }): Promise<PaginatedResponse<Crash>> => {
    const query = new URLSearchParams()
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined) query.set(key, String(value))
      })
    }
    return await authFetch<PaginatedResponse<Crash>>(`${baseUrl}/crashes?${query}`)
  }

  const getCrash = async (id: string): Promise<Crash> => {
    return await authFetch<Crash>(`${baseUrl}/crashes/${id}`)
  }

  const deleteCrash = async (id: string): Promise<void> => {
    await authFetch(`${baseUrl}/crashes/${id}`, { method: 'DELETE' })
  }

  // Groups
  const getGroups = async (params?: {
    app_id?: string
    status?: string
    search?: string
    sort_by?: string
    sort_order?: string
    limit?: number
    offset?: number
  }): Promise<PaginatedResponse<CrashGroup>> => {
    const query = new URLSearchParams()
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined) query.set(key, String(value))
      })
    }
    return await authFetch<PaginatedResponse<CrashGroup>>(`${baseUrl}/groups?${query}`)
  }

  const getGroup = async (id: string): Promise<CrashGroup> => {
    return await authFetch<CrashGroup>(`${baseUrl}/groups/${id}`)
  }

  const updateGroup = async (id: string, data: Partial<CrashGroup>): Promise<CrashGroup> => {
    return await authFetch<CrashGroup>(`${baseUrl}/groups/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    })
  }

  // Apps
  const getApps = async (): Promise<{ data: App[] }> => {
    return await authFetch<{ data: App[] }>(`${baseUrl}/apps`)
  }

  const getApp = async (id: string): Promise<App> => {
    return await authFetch<App>(`${baseUrl}/apps/${id}`)
  }

  const createApp = async (name: string, retention_days?: number): Promise<App> => {
    return await authFetch<App>(`${baseUrl}/apps`, {
      method: 'POST',
      body: JSON.stringify({ name, retention_days }),
    })
  }

  const getAppStats = async (id: string): Promise<CrashStats> => {
    return await authFetch<CrashStats>(`${baseUrl}/apps/${id}/stats`)
  }

  const regenerateAppKey = async (id: string): Promise<{ id: string; name: string; api_key: string }> => {
    return await authFetch<{ id: string; name: string; api_key: string }>(`${baseUrl}/apps/${id}/regenerate-key`, {
      method: 'POST',
    })
  }

  // Alerts
  const getAlerts = async (app_id?: string): Promise<{ data: Alert[] }> => {
    const query = app_id ? `?app_id=${app_id}` : ''
    return await authFetch<{ data: Alert[] }>(`${baseUrl}/alerts${query}`)
  }

  const createAlert = async (data: Omit<Alert, 'id' | 'created_at'>): Promise<Alert> => {
    return await authFetch<Alert>(`${baseUrl}/alerts`, {
      method: 'POST',
      body: JSON.stringify(data),
    })
  }

  const deleteAlert = async (id: string): Promise<void> => {
    await authFetch(`${baseUrl}/alerts/${id}`, { method: 'DELETE' })
  }

  return {
    token,
    isAuthenticated,
    needsPasswordChange,
    loadToken,
    checkAuthStatus,
    login,
    logout,
    changePassword,
    getCrashes,
    getCrash,
    deleteCrash,
    getGroups,
    getGroup,
    updateGroup,
    getApps,
    getApp,
    createApp,
    getAppStats,
    regenerateAppKey,
    getAlerts,
    createAlert,
    deleteAlert,
  }
}
