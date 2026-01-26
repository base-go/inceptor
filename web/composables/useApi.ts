import type { Crash, CrashGroup, App, Alert, CrashStats, PaginatedResponse } from '~/types'

export const useApi = () => {
  const config = useRuntimeConfig()
  const apiKey = useState<string>('apiKey', () => '')

  const baseUrl = config.public.apiBase

  const headers = computed(() => ({
    'Content-Type': 'application/json',
    'X-API-Key': apiKey.value,
  }))

  const setApiKey = (key: string) => {
    apiKey.value = key
    if (process.client) {
      localStorage.setItem('inceptor_api_key', key)
    }
  }

  const loadApiKey = () => {
    if (process.client) {
      const savedKey = localStorage.getItem('inceptor_api_key')
      if (savedKey) {
        apiKey.value = savedKey
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

    const response = await $fetch<PaginatedResponse<Crash>>(`${baseUrl}/crashes?${query}`, {
      headers: headers.value,
    })
    return response
  }

  const getCrash = async (id: string): Promise<Crash> => {
    return await $fetch<Crash>(`${baseUrl}/crashes/${id}`, {
      headers: headers.value,
    })
  }

  const deleteCrash = async (id: string): Promise<void> => {
    await $fetch(`${baseUrl}/crashes/${id}`, {
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

    return await $fetch<PaginatedResponse<CrashGroup>>(`${baseUrl}/groups?${query}`, {
      headers: headers.value,
    })
  }

  const getGroup = async (id: string): Promise<CrashGroup> => {
    return await $fetch<CrashGroup>(`${baseUrl}/groups/${id}`, {
      headers: headers.value,
    })
  }

  const updateGroup = async (id: string, data: Partial<CrashGroup>): Promise<CrashGroup> => {
    return await $fetch<CrashGroup>(`${baseUrl}/groups/${id}`, {
      method: 'PATCH',
      headers: headers.value,
      body: data,
    })
  }

  // Apps
  const getApps = async (): Promise<{ data: App[] }> => {
    return await $fetch<{ data: App[] }>(`${baseUrl}/apps`, {
      headers: headers.value,
    })
  }

  const getApp = async (id: string): Promise<App> => {
    return await $fetch<App>(`${baseUrl}/apps/${id}`, {
      headers: headers.value,
    })
  }

  const createApp = async (name: string, retention_days?: number): Promise<App> => {
    return await $fetch<App>(`${baseUrl}/apps`, {
      method: 'POST',
      headers: headers.value,
      body: { name, retention_days },
    })
  }

  const getAppStats = async (id: string): Promise<CrashStats> => {
    return await $fetch<CrashStats>(`${baseUrl}/apps/${id}/stats`, {
      headers: headers.value,
    })
  }

  // Alerts
  const getAlerts = async (app_id?: string): Promise<{ data: Alert[] }> => {
    const query = app_id ? `?app_id=${app_id}` : ''
    return await $fetch<{ data: Alert[] }>(`${baseUrl}/alerts${query}`, {
      headers: headers.value,
    })
  }

  const createAlert = async (data: Omit<Alert, 'id' | 'created_at'>): Promise<Alert> => {
    return await $fetch<Alert>(`${baseUrl}/alerts`, {
      method: 'POST',
      headers: headers.value,
      body: data,
    })
  }

  const deleteAlert = async (id: string): Promise<void> => {
    await $fetch(`${baseUrl}/alerts/${id}`, {
      method: 'DELETE',
      headers: headers.value,
    })
  }

  return {
    apiKey,
    setApiKey,
    loadApiKey,
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
