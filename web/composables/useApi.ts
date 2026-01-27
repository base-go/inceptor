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

  // Use relative URL in browser, fallback to config for SSR
  const baseUrl = computed(() => {
    if (process.client) {
      return '/api/v1'
    }
    return useRuntimeConfig().public.apiBase
  })

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

  // Auth methods
  const checkAuthStatus = async (): Promise<AuthStatus> => {
    return await $fetch<AuthStatus>(`${baseUrl.value}/auth/status`)
  }

  const login = async (password: string): Promise<LoginResponse> => {
    const response = await $fetch<LoginResponse>(`${baseUrl.value}/auth/login`, {
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
      await $fetch(`${baseUrl.value}/auth/logout`, {
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
    await $fetch(`${baseUrl.value}/auth/change-password`, {
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

    return await $fetch<PaginatedResponse<Crash>>(`${baseUrl.value}/crashes?${query}`, {
      headers: headers.value,
    })
  }

  const getCrash = async (id: string): Promise<Crash> => {
    return await $fetch<Crash>(`${baseUrl.value}/crashes/${id}`, {
      headers: headers.value,
    })
  }

  const deleteCrash = async (id: string): Promise<void> => {
    await $fetch(`${baseUrl.value}/crashes/${id}`, {
      method: 'DELETE',
      headers: headers.value,
    })
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

    return await $fetch<PaginatedResponse<CrashGroup>>(`${baseUrl.value}/groups?${query}`, {
      headers: headers.value,
    })
  }

  const getGroup = async (id: string): Promise<CrashGroup> => {
    return await $fetch<CrashGroup>(`${baseUrl.value}/groups/${id}`, {
      headers: headers.value,
    })
  }

  const updateGroup = async (id: string, data: Partial<CrashGroup>): Promise<CrashGroup> => {
    return await $fetch<CrashGroup>(`${baseUrl.value}/groups/${id}`, {
      method: 'PATCH',
      headers: headers.value,
      body: data,
    })
  }

  // Apps
  const getApps = async (): Promise<{ data: App[] }> => {
    return await $fetch<{ data: App[] }>(`${baseUrl.value}/apps`, {
      headers: headers.value,
    })
  }

  const getApp = async (id: string): Promise<App> => {
    return await $fetch<App>(`${baseUrl.value}/apps/${id}`, {
      headers: headers.value,
    })
  }

  const createApp = async (name: string, retention_days?: number): Promise<App> => {
    return await $fetch<App>(`${baseUrl.value}/apps`, {
      method: 'POST',
      headers: headers.value,
      body: { name, retention_days },
    })
  }

  const getAppStats = async (id: string): Promise<CrashStats> => {
    return await $fetch<CrashStats>(`${baseUrl.value}/apps/${id}/stats`, {
      headers: headers.value,
    })
  }

  // Alerts
  const getAlerts = async (app_id?: string): Promise<{ data: Alert[] }> => {
    const query = app_id ? `?app_id=${app_id}` : ''
    return await $fetch<{ data: Alert[] }>(`${baseUrl.value}/alerts${query}`, {
      headers: headers.value,
    })
  }

  const createAlert = async (data: Omit<Alert, 'id' | 'created_at'>): Promise<Alert> => {
    return await $fetch<Alert>(`${baseUrl.value}/alerts`, {
      method: 'POST',
      headers: headers.value,
      body: data,
    })
  }

  const deleteAlert = async (id: string): Promise<void> => {
    await $fetch(`${baseUrl.value}/alerts/${id}`, {
      method: 'DELETE',
      headers: headers.value,
    })
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
    getAlerts,
    createAlert,
    deleteAlert,
  }
}
