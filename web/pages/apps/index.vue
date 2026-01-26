<script setup lang="ts">
import type { App } from '~/types'

definePageMeta({
  title: 'Apps',
})

const api = useApi()

const apps = ref<App[]>([])
const loading = ref(true)
const showCreateModal = ref(false)
const newAppName = ref('')
const newAppRetention = ref(30)
const createdApp = ref<App | null>(null)

const loadApps = async () => {
  if (!api.isAuthenticated.value) {
    loading.value = false
    return
  }

  try {
    loading.value = true
    const response = await api.getApps()
    apps.value = response.data || []
  } catch (e) {
    console.error('Failed to load apps:', e)
  } finally {
    loading.value = false
  }
}

onMounted(loadApps)

const formatDate = (date: string) => {
  return new Date(date).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

const createApp = async () => {
  if (!newAppName.value.trim()) return

  try {
    createdApp.value = await api.createApp(newAppName.value.trim(), newAppRetention.value)
    showCreateModal.value = false
    newAppName.value = ''
    newAppRetention.value = 30
    loadApps()
  } catch (e) {
    console.error('Failed to create app:', e)
  }
}

const copyApiKey = () => {
  if (createdApp.value?.api_key) {
    navigator.clipboard.writeText(createdApp.value.api_key)
  }
}
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <h2 class="text-xl font-semibold">Registered Apps</h2>
      <UButton icon="i-heroicons-plus" @click="showCreateModal = true">
        Create App
      </UButton>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center py-12">
      <UIcon name="i-heroicons-arrow-path" class="w-8 h-8 animate-spin text-primary-500" />
    </div>

    <!-- Apps Grid -->
    <div v-else-if="apps.length" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <UCard v-for="app in apps" :key="app.id">
        <div class="flex items-start justify-between">
          <div>
            <h3 class="text-lg font-semibold text-white">{{ app.name }}</h3>
            <p class="text-sm text-gray-500 mt-1">Created {{ formatDate(app.created_at) }}</p>
          </div>
          <UIcon name="i-heroicons-squares-2x2" class="w-8 h-8 text-primary-500" />
        </div>

        <div class="mt-4 space-y-2">
          <div class="flex items-center justify-between text-sm">
            <span class="text-gray-400">App ID</span>
            <code class="text-gray-300 font-mono text-xs">{{ app.id.slice(0, 8) }}...</code>
          </div>
          <div class="flex items-center justify-between text-sm">
            <span class="text-gray-400">Retention</span>
            <span class="text-gray-300">{{ app.retention_days }} days</span>
          </div>
        </div>

        <div class="mt-4 pt-4 border-t border-gray-700">
          <NuxtLink :to="`/?app_id=${app.id}`">
            <UButton variant="ghost" size="sm" class="w-full">
              View Dashboard
            </UButton>
          </NuxtLink>
        </div>
      </UCard>
    </div>

    <!-- Empty State -->
    <UCard v-else class="text-center py-12">
      <UIcon name="i-heroicons-squares-2x2" class="w-12 h-12 mx-auto text-gray-600" />
      <h3 class="mt-4 text-lg font-medium text-white">No apps registered</h3>
      <p class="mt-2 text-gray-400">Create your first app to start tracking crashes.</p>
      <UButton class="mt-4" @click="showCreateModal = true">
        Create App
      </UButton>
    </UCard>

    <!-- Create Modal -->
    <UModal v-model="showCreateModal">
      <UCard>
        <template #header>
          <h3 class="text-lg font-semibold">Create New App</h3>
        </template>

        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-300 mb-1">App Name</label>
            <UInput v-model="newAppName" placeholder="My Flutter App" />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-300 mb-1">Retention Days</label>
            <p class="text-xs text-gray-500 mb-1">How long to keep crash data</p>
            <UInput v-model="newAppRetention" type="number" min="1" max="365" />
          </div>
        </div>

        <template #footer>
          <div class="flex justify-end gap-2">
            <UButton variant="ghost" @click="showCreateModal = false">Cancel</UButton>
            <UButton :disabled="!newAppName.trim()" @click="createApp">Create</UButton>
          </div>
        </template>
      </UCard>
    </UModal>

    <!-- API Key Modal -->
    <UModal :model-value="!!createdApp" @update:model-value="createdApp = null">
      <UCard>
        <template #header>
          <div class="flex items-center gap-2 text-green-500">
            <UIcon name="i-heroicons-check-circle" class="w-6 h-6" />
            <h3 class="text-lg font-semibold">App Created Successfully</h3>
          </div>
        </template>

        <UAlert color="yellow" icon="i-heroicons-exclamation-triangle" class="mb-4">
          <template #title>Save your API key now!</template>
          <template #description>
            This is the only time you'll see this API key. Store it securely.
          </template>
        </UAlert>

        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-300 mb-1">App Name</label>
            <UInput :model-value="createdApp?.name" readonly />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-300 mb-1">API Key</label>
            <div class="flex gap-2">
              <UInput :model-value="createdApp?.api_key" readonly class="flex-1 font-mono text-sm" />
              <UButton icon="i-heroicons-clipboard-document" variant="outline" @click="copyApiKey">
                Copy
              </UButton>
            </div>
          </div>
        </div>

        <template #footer>
          <div class="flex justify-end">
            <UButton @click="createdApp = null">Done</UButton>
          </div>
        </template>
      </UCard>
    </UModal>
  </div>
</template>
