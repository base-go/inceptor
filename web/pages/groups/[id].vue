<script setup lang="ts">
import type { CrashGroup, Crash, PaginatedResponse } from '~/types'

definePageMeta({
  title: 'Group Details',
})

const route = useRoute()
const api = useApi()

const group = ref<CrashGroup | null>(null)
const crashes = ref<Crash[]>([])
const loading = ref(true)
const error = ref<string | null>(null)
const updating = ref(false)

const loadGroup = async () => {
  try {
    loading.value = true
    error.value = null

    const [groupData, crashesData] = await Promise.all([
      api.getGroup(route.params.id as string),
      api.getCrashes({ group_id: route.params.id as string, limit: 50 }),
    ])

    group.value = groupData
    crashes.value = crashesData.data
  } catch (e: any) {
    error.value = e.message || 'Failed to load group'
  } finally {
    loading.value = false
  }
}

onMounted(loadGroup)

const formatDate = (date: string) => {
  return new Date(date).toLocaleString('en-US', {
    weekday: 'short',
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const formatRelativeTime = (date: string) => {
  const now = new Date()
  const then = new Date(date)
  const diff = now.getTime() - then.getTime()

  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (minutes < 1) return 'just now'
  if (minutes < 60) return `${minutes}m ago`
  if (hours < 24) return `${hours}h ago`
  if (days < 7) return `${days}d ago`
  return formatDate(date)
}

const statusOptions = [
  { label: 'Open', value: 'open', color: 'yellow' },
  { label: 'Resolved', value: 'resolved', color: 'green' },
  { label: 'Ignored', value: 'ignored', color: 'gray' },
]

const updateStatus = async (status: string) => {
  if (!group.value || updating.value) return

  try {
    updating.value = true
    group.value = await api.updateGroup(group.value.id, { status: status as CrashGroup['status'] })
  } catch (e: any) {
    error.value = e.message || 'Failed to update status'
  } finally {
    updating.value = false
  }
}

const getStatusColor = (status: string) => {
  switch (status) {
    case 'open': return 'yellow'
    case 'resolved': return 'green'
    case 'ignored': return 'gray'
    default: return 'gray'
  }
}
</script>

<template>
  <div class="space-y-6">
    <!-- Back button -->
    <div>
      <UButton
        icon="i-heroicons-arrow-left"
        variant="ghost"
        @click="$router.back()"
      >
        Back
      </UButton>
    </div>

    <!-- Loading -->
    <div v-if="loading" class="flex items-center justify-center py-12">
      <UIcon name="i-heroicons-arrow-path" class="w-8 h-8 animate-spin text-primary-500" />
    </div>

    <!-- Error -->
    <UAlert v-else-if="error" color="red" :title="error" />

    <!-- Group Details -->
    <template v-else-if="group">
      <!-- Header -->
      <UCard>
        <div class="flex items-start justify-between">
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-3">
              <h2 class="text-xl font-bold text-white">{{ group.error_type }}</h2>
              <UBadge :color="getStatusColor(group.status)" size="lg">
                {{ group.status }}
              </UBadge>
            </div>
            <p class="mt-2 text-gray-400 break-words">{{ group.error_message }}</p>
          </div>
          <div class="flex gap-2 ml-4">
            <UDropdown
              :items="[[
                ...statusOptions.map(opt => ({
                  label: opt.label,
                  icon: group.status === opt.value ? 'i-heroicons-check' : undefined,
                  click: () => updateStatus(opt.value),
                }))
              ]]"
            >
              <UButton
                variant="outline"
                icon="i-heroicons-pencil-square"
                :loading="updating"
              >
                Change Status
              </UButton>
            </UDropdown>
          </div>
        </div>
      </UCard>

      <!-- Stats Grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <UCard>
          <div class="text-sm text-gray-400">Occurrences</div>
          <div class="mt-1 text-2xl font-bold text-primary-500">{{ group.occurrence_count }}</div>
        </UCard>
        <UCard>
          <div class="text-sm text-gray-400">First Seen</div>
          <div class="mt-1 font-medium">{{ formatDate(group.first_seen) }}</div>
        </UCard>
        <UCard>
          <div class="text-sm text-gray-400">Last Seen</div>
          <div class="mt-1 font-medium">{{ formatRelativeTime(group.last_seen) }}</div>
        </UCard>
        <UCard>
          <div class="text-sm text-gray-400">Fingerprint</div>
          <div class="mt-1 font-mono text-sm truncate">{{ group.fingerprint }}</div>
        </UCard>
      </div>

      <!-- Crashes List -->
      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold">Crash Instances ({{ crashes.length }})</h3>
          </div>
        </template>

        <div v-if="crashes.length === 0" class="text-center py-8 text-gray-500">
          No crash instances found
        </div>

        <div v-else class="divide-y divide-gray-800">
          <NuxtLink
            v-for="crash in crashes"
            :key="crash.id"
            :to="`/crashes/${crash.id}`"
            class="block p-4 hover:bg-gray-800/50 transition-colors"
          >
            <div class="flex items-center justify-between">
              <div class="flex-1 min-w-0">
                <div class="flex items-center gap-2">
                  <span class="font-medium">{{ crash.platform }}</span>
                  <span class="text-gray-500">·</span>
                  <span class="text-gray-400">v{{ crash.app_version }}</span>
                  <UBadge
                    v-if="crash.environment !== 'production'"
                    :color="crash.environment === 'staging' ? 'yellow' : 'gray'"
                    size="xs"
                  >
                    {{ crash.environment }}
                  </UBadge>
                </div>
                <div class="mt-1 text-sm text-gray-500">
                  <span v-if="crash.user_id">User: {{ crash.user_id }} · </span>
                  <span v-if="crash.device_model">{{ crash.device_model }} · </span>
                  {{ crash.os_version || 'Unknown OS' }}
                </div>
              </div>
              <div class="text-sm text-gray-500 ml-4">
                {{ formatRelativeTime(crash.created_at) }}
              </div>
            </div>
          </NuxtLink>
        </div>
      </UCard>
    </template>
  </div>
</template>
