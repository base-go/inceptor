<script setup lang="ts">
import type { Alert } from '~/types'

definePageMeta({
  title: 'Settings',
})

const api = useApi()

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

onMounted(loadData)

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
    <!-- Alerts Section -->
    <UCard>
      <template #header>
        <div class="flex items-center justify-between">
          <h3 class="text-lg font-semibold">Alert Rules</h3>
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
