<script setup lang="ts">
import type { Alert } from '~/types'

definePageMeta({
  title: 'Settings',
})

const api = useApi()

// Version state
const versionInfo = ref<{ current: string; latest: string; updateAvailable: boolean } | null>(null)
const loadingVersion = ref(true)
const updating = ref(false)
const updateError = ref<string | null>(null)
const updateSuccess = ref(false)

// Alerts state
const alerts = ref<Alert[]>([])
const loading = ref(true)
const showCreateAlert = ref(false)

const newAlert = ref({
  app_id: '',
  type: 'webhook' as const,
  config: {
    url: '',
    conditions: {
      on_new_group: true,
    },
  },
  enabled: true,
})

const apps = ref<{ id: string; name: string }[]>([])

const alertTypes = [
  { label: 'Webhook', value: 'webhook', icon: 'i-heroicons-globe-alt' },
  { label: 'Email', value: 'email', icon: 'i-heroicons-envelope' },
  { label: 'Slack', value: 'slack', icon: 'i-simple-icons-slack' },
]

const loadVersion = async () => {
  try {
    loadingVersion.value = true
    versionInfo.value = await api.getVersion()
  } catch (e) {
    console.error('Failed to load version:', e)
  } finally {
    loadingVersion.value = false
  }
}

const performUpdate = async () => {
  if (!api.isAuthenticated.value) {
    updateError.value = 'You must be logged in to update'
    return
  }

  try {
    updating.value = true
    updateError.value = null
    updateSuccess.value = false
    await api.triggerUpdate()
    updateSuccess.value = true
    // Reload version after a delay (server restarts)
    setTimeout(() => {
      loadVersion()
    }, 5000)
  } catch (e: any) {
    updateError.value = e.message || 'Failed to update'
  } finally {
    updating.value = false
  }
}

const loadData = async () => {
  if (!api.isAuthenticated.value) {
    loading.value = false
    return
  }

  try {
    loading.value = true
    const [alertsResponse, appsResponse] = await Promise.all([
      api.getAlerts(),
      api.getApps(),
    ])
    alerts.value = alertsResponse.data || []
    apps.value = appsResponse.data || []

    if (apps.value.length && !newAlert.value.app_id) {
      newAlert.value.app_id = apps.value[0].id
    }
  } catch (e) {
    console.error('Failed to load data:', e)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  loadVersion()
  loadData()
})

const createAlert = async () => {
  try {
    await api.createAlert(newAlert.value)
    showCreateAlert.value = false
    newAlert.value = {
      app_id: apps.value[0]?.id || '',
      type: 'webhook',
      config: {
        url: '',
        conditions: {
          on_new_group: true,
        },
      },
      enabled: true,
    }
    loadData()
  } catch (e) {
    console.error('Failed to create alert:', e)
  }
}

const deleteAlert = async (id: string) => {
  if (!confirm('Are you sure you want to delete this alert?')) return

  try {
    await api.deleteAlert(id)
    loadData()
  } catch (e) {
    console.error('Failed to delete alert:', e)
  }
}

const typeIcon = (type: string) => {
  return alertTypes.find((t) => t.value === type)?.icon || 'i-heroicons-bell'
}
</script>

<template>
  <div class="space-y-6 max-w-4xl">
    <!-- System Section -->
    <UCard>
      <template #header>
        <div class="flex items-center gap-2">
          <UIcon name="i-heroicons-server" class="w-5 h-5 text-primary-500" />
          <h3 class="text-lg font-semibold">System</h3>
        </div>
      </template>

      <div class="space-y-4">
        <!-- Version Info -->
        <div class="flex items-center justify-between p-4 rounded-lg bg-gray-800/50">
          <div class="flex items-center gap-4">
            <div class="p-3 rounded-lg bg-primary-500/10">
              <UIcon name="i-heroicons-cube" class="w-6 h-6 text-primary-500" />
            </div>
            <div>
              <div class="text-sm text-gray-400">Inceptor Version</div>
              <div v-if="loadingVersion" class="flex items-center gap-2 mt-1">
                <UIcon name="i-heroicons-arrow-path" class="w-4 h-4 animate-spin" />
                <span class="text-gray-500">Loading...</span>
              </div>
              <div v-else-if="versionInfo" class="flex items-center gap-3 mt-1">
                <span class="text-xl font-bold">v{{ versionInfo.current }}</span>
                <UBadge v-if="versionInfo.updateAvailable" color="yellow" size="sm">
                  Update available
                </UBadge>
                <UBadge v-else color="green" size="sm">
                  Up to date
                </UBadge>
              </div>
            </div>
          </div>
          <div v-if="versionInfo?.updateAvailable" class="flex flex-col items-end gap-2">
            <UButton
              icon="i-heroicons-arrow-down-tray"
              :loading="updating"
              @click="performUpdate"
            >
              Update to v{{ versionInfo.latest }}
            </UButton>
            <span class="text-xs text-gray-500">Server will restart</span>
          </div>
        </div>

        <!-- Update Status -->
        <UAlert
          v-if="updateSuccess"
          color="green"
          icon="i-heroicons-check-circle"
          title="Update initiated"
          description="The server is updating and will restart. This page will refresh automatically."
        />
        <UAlert
          v-if="updateError"
          color="red"
          icon="i-heroicons-exclamation-triangle"
          :title="updateError"
        />

        <!-- System Info -->
        <div class="grid grid-cols-2 gap-4">
          <div class="p-4 rounded-lg bg-gray-800/30">
            <div class="text-sm text-gray-400">Status</div>
            <div class="flex items-center gap-2 mt-1">
              <span class="w-2 h-2 rounded-full bg-green-500"></span>
              <span class="font-medium">Healthy</span>
            </div>
          </div>
          <div class="p-4 rounded-lg bg-gray-800/30">
            <div class="text-sm text-gray-400">GitHub</div>
            <a
              href="https://github.com/base-go/inceptor"
              target="_blank"
              class="flex items-center gap-2 mt-1 text-primary-500 hover:underline"
            >
              <span class="font-medium">base-go/inceptor</span>
              <UIcon name="i-heroicons-arrow-top-right-on-square" class="w-4 h-4" />
            </a>
          </div>
        </div>
      </div>
    </UCard>

    <!-- Alerts Section -->
    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-2">
            <UIcon name="i-heroicons-bell" class="w-5 h-5 text-primary-500" />
            <h3 class="text-lg font-semibold">Alert Rules</h3>
          </div>
          <UButton icon="i-heroicons-plus" size="sm" @click="showCreateAlert = true">
            Add Alert
          </UButton>
        </div>
      </template>

      <!-- Loading -->
      <div v-if="loading" class="flex items-center justify-center py-8">
        <UIcon name="i-heroicons-arrow-path" class="w-6 h-6 animate-spin text-primary-500" />
      </div>

      <!-- Alerts List -->
      <div v-else-if="alerts.length" class="space-y-3">
        <div
          v-for="alert in alerts"
          :key="alert.id"
          class="flex items-center justify-between p-4 rounded-lg bg-gray-800/50"
        >
          <div class="flex items-center gap-4">
            <div class="p-2 rounded-lg bg-gray-700">
              <UIcon :name="typeIcon(alert.type)" class="w-5 h-5" />
            </div>
            <div>
              <div class="flex items-center gap-2">
                <span class="font-medium capitalize">{{ alert.type }}</span>
                <UBadge :color="alert.enabled ? 'green' : 'gray'" size="xs">
                  {{ alert.enabled ? 'Active' : 'Disabled' }}
                </UBadge>
              </div>
              <div class="text-sm text-gray-400 mt-1">
                <span v-if="alert.type === 'webhook'">{{ (alert.config as any).url }}</span>
                <span v-else-if="alert.type === 'email'">{{ (alert.config as any).to }}</span>
                <span v-else>Configured</span>
              </div>
            </div>
          </div>
          <UButton
            icon="i-heroicons-trash"
            color="red"
            variant="ghost"
            @click="deleteAlert(alert.id)"
          />
        </div>
      </div>

      <!-- Empty State -->
      <div v-else class="text-center py-8 text-gray-500">
        <UIcon name="i-heroicons-bell-slash" class="w-12 h-12 mx-auto mb-4" />
        <p>No alerts configured</p>
        <p class="text-sm mt-1">Create an alert to get notified of crashes</p>
      </div>
    </UCard>

    <!-- Documentation Section -->
    <UCard>
      <template #header>
        <div class="flex items-center gap-2">
          <UIcon name="i-heroicons-book-open" class="w-5 h-5 text-primary-500" />
          <h3 class="text-lg font-semibold">Documentation</h3>
        </div>
      </template>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <a
          href="https://github.com/base-go/inceptor#readme"
          target="_blank"
          class="flex items-center gap-4 p-4 rounded-lg bg-gray-800/50 hover:bg-gray-800 transition-colors"
        >
          <UIcon name="i-heroicons-document-text" class="w-8 h-8 text-gray-400" />
          <div>
            <div class="font-medium">Getting Started</div>
            <div class="text-sm text-gray-500">Setup and configuration guide</div>
          </div>
        </a>
        <a
          href="https://pub.dev/packages/inceptor_flutter"
          target="_blank"
          class="flex items-center gap-4 p-4 rounded-lg bg-gray-800/50 hover:bg-gray-800 transition-colors"
        >
          <UIcon name="i-simple-icons-flutter" class="w-8 h-8 text-gray-400" />
          <div>
            <div class="font-medium">Flutter SDK</div>
            <div class="text-sm text-gray-500">Integrate with your Flutter app</div>
          </div>
        </a>
        <a
          href="https://github.com/base-go/inceptor/blob/main/docs/api-reference.md"
          target="_blank"
          class="flex items-center gap-4 p-4 rounded-lg bg-gray-800/50 hover:bg-gray-800 transition-colors"
        >
          <UIcon name="i-heroicons-code-bracket" class="w-8 h-8 text-gray-400" />
          <div>
            <div class="font-medium">API Reference</div>
            <div class="text-sm text-gray-500">REST API documentation</div>
          </div>
        </a>
        <a
          href="https://github.com/base-go/inceptor/issues"
          target="_blank"
          class="flex items-center gap-4 p-4 rounded-lg bg-gray-800/50 hover:bg-gray-800 transition-colors"
        >
          <UIcon name="i-heroicons-bug-ant" class="w-8 h-8 text-gray-400" />
          <div>
            <div class="font-medium">Report Issue</div>
            <div class="text-sm text-gray-500">Bug reports and feature requests</div>
          </div>
        </a>
      </div>
    </UCard>

    <!-- Create Alert Modal -->
    <UModal v-model="showCreateAlert">
      <UCard>
        <template #header>
          <h3 class="text-lg font-semibold">Create Alert</h3>
        </template>

        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-300 mb-1">App</label>
            <USelectMenu
              v-model="newAlert.app_id"
              :options="apps"
              option-attribute="name"
              value-attribute="id"
              placeholder="Select an app"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-300 mb-1">Alert Type</label>
            <USelectMenu
              v-model="newAlert.type"
              :options="alertTypes"
              option-attribute="label"
              value-attribute="value"
            />
          </div>

          <!-- Webhook Config -->
          <template v-if="newAlert.type === 'webhook'">
            <div>
              <label class="block text-sm font-medium text-gray-300 mb-1">Webhook URL</label>
              <UInput v-model="(newAlert.config as any).url" placeholder="https://example.com/webhook" />
            </div>
          </template>

          <!-- Email Config -->
          <template v-if="newAlert.type === 'email'">
            <div>
              <label class="block text-sm font-medium text-gray-300 mb-1">Email Address</label>
              <UInput v-model="(newAlert.config as any).to" type="email" placeholder="alerts@example.com" />
            </div>
          </template>

          <!-- Slack Config -->
          <template v-if="newAlert.type === 'slack'">
            <div>
              <label class="block text-sm font-medium text-gray-300 mb-1">Slack Webhook URL</label>
              <UInput v-model="(newAlert.config as any).webhook_url" placeholder="https://hooks.slack.com/..." />
            </div>
          </template>

          <div>
            <label class="block text-sm font-medium text-gray-300 mb-1">Conditions</label>
            <div class="space-y-2">
              <UCheckbox
                v-model="(newAlert.config as any).conditions.on_new_group"
                label="Alert on new crash group"
              />
              <UCheckbox
                v-model="(newAlert.config as any).conditions.on_every_crash"
                label="Alert on every crash"
              />
            </div>
          </div>

          <div>
            <UCheckbox v-model="newAlert.enabled" label="Enable this alert" />
          </div>
        </div>

        <template #footer>
          <div class="flex justify-end gap-2">
            <UButton variant="ghost" @click="showCreateAlert = false">Cancel</UButton>
            <UButton @click="createAlert">Create</UButton>
          </div>
        </template>
      </UCard>
    </UModal>
  </div>
</template>
