<script setup lang="ts">
import type { Crash } from '~/types'

definePageMeta({
  title: 'Crash Details',
})

const route = useRoute()
const api = useApi()

const crash = ref<Crash | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)

const loadCrash = async () => {
  try {
    loading.value = true
    error.value = null
    crash.value = await api.getCrash(route.params.id as string)
  } catch (e: any) {
    error.value = e.message || 'Failed to load crash'
  } finally {
    loading.value = false
  }
}

onMounted(loadCrash)

const formatDate = (date: string) => {
  return new Date(date).toLocaleString('en-US', {
    weekday: 'short',
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

const copyToClipboard = (text: string) => {
  navigator.clipboard.writeText(text)
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

    <!-- Crash Details -->
    <template v-else-if="crash">
      <!-- Header -->
      <UCard>
        <div class="flex items-start justify-between">
          <div>
            <h2 class="text-xl font-bold text-white">{{ crash.error_type }}</h2>
            <p class="mt-1 text-gray-400">{{ crash.error_message }}</p>
          </div>
          <div class="flex gap-2">
            <NuxtLink :to="`/groups/${crash.group_id}`">
              <UButton variant="outline" icon="i-heroicons-rectangle-stack">
                View Group
              </UButton>
            </NuxtLink>
          </div>
        </div>
      </UCard>

      <!-- Info Grid -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <UCard>
          <div class="text-sm text-gray-400">Platform</div>
          <div class="mt-1 font-medium capitalize">{{ crash.platform }}</div>
        </UCard>
        <UCard>
          <div class="text-sm text-gray-400">App Version</div>
          <div class="mt-1 font-medium">{{ crash.app_version }}</div>
        </UCard>
        <UCard>
          <div class="text-sm text-gray-400">OS Version</div>
          <div class="mt-1 font-medium">{{ crash.os_version || 'N/A' }}</div>
        </UCard>
        <UCard>
          <div class="text-sm text-gray-400">Device</div>
          <div class="mt-1 font-medium">{{ crash.device_model || 'N/A' }}</div>
        </UCard>
        <UCard>
          <div class="text-sm text-gray-400">Environment</div>
          <div class="mt-1">
            <UBadge
              :color="crash.environment === 'production' ? 'red' : crash.environment === 'staging' ? 'yellow' : 'gray'"
            >
              {{ crash.environment }}
            </UBadge>
          </div>
        </UCard>
        <UCard>
          <div class="text-sm text-gray-400">User ID</div>
          <div class="mt-1 font-medium">{{ crash.user_id || 'Anonymous' }}</div>
        </UCard>
        <UCard>
          <div class="text-sm text-gray-400">Fingerprint</div>
          <div class="mt-1 font-mono text-sm">{{ crash.fingerprint }}</div>
        </UCard>
        <UCard>
          <div class="text-sm text-gray-400">Occurred At</div>
          <div class="mt-1 font-medium">{{ formatDate(crash.created_at) }}</div>
        </UCard>
      </div>

      <!-- Stack Trace -->
      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold">Stack Trace</h3>
            <UButton
              icon="i-heroicons-clipboard-document"
              variant="ghost"
              size="sm"
              @click="copyToClipboard(JSON.stringify(crash.stack_trace, null, 2))"
            >
              Copy
            </UButton>
          </div>
        </template>

        <div class="space-y-1 font-mono text-sm">
          <div
            v-for="(frame, index) in crash.stack_trace"
            :key="index"
            :class="[
              'p-3 rounded-lg',
              frame.native ? 'bg-gray-800/30 text-gray-500' : 'bg-gray-800/50 text-gray-300',
            ]"
          >
            <div class="flex items-start">
              <span class="text-gray-500 mr-3 select-none">#{{ index }}</span>
              <div class="flex-1 min-w-0">
                <span v-if="frame.class_name" class="text-blue-400">{{ frame.class_name }}.</span>
                <span class="text-yellow-400">{{ frame.method_name }}</span>
                <div class="mt-1 text-gray-500 truncate">
                  {{ frame.file_name }}
                  <span v-if="frame.line_number">:{{ frame.line_number }}</span>
                  <span v-if="frame.column_number">:{{ frame.column_number }}</span>
                </div>
              </div>
              <UBadge v-if="frame.native" color="gray" size="xs">native</UBadge>
            </div>
          </div>
        </div>
      </UCard>

      <!-- Breadcrumbs -->
      <UCard v-if="crash.breadcrumbs?.length">
        <template #header>
          <h3 class="text-lg font-semibold">Breadcrumbs</h3>
        </template>

        <div class="space-y-2">
          <div
            v-for="(breadcrumb, index) in crash.breadcrumbs"
            :key="index"
            class="flex items-start p-3 rounded-lg bg-gray-800/50"
          >
            <div class="mr-3">
              <UBadge
                :color="breadcrumb.level === 'error' ? 'red' : breadcrumb.level === 'warning' ? 'yellow' : 'gray'"
                size="xs"
              >
                {{ breadcrumb.type }}
              </UBadge>
            </div>
            <div class="flex-1">
              <div class="text-sm font-medium">{{ breadcrumb.message }}</div>
              <div class="text-xs text-gray-500 mt-1">
                {{ breadcrumb.category }} · {{ formatDate(breadcrumb.timestamp) }}
              </div>
            </div>
          </div>
        </div>
      </UCard>

      <!-- Metadata -->
      <UCard v-if="crash.metadata && Object.keys(crash.metadata).length">
        <template #header>
          <h3 class="text-lg font-semibold">Metadata</h3>
        </template>

        <pre class="p-4 rounded-lg bg-gray-800/50 overflow-auto text-sm text-gray-300">{{ JSON.stringify(crash.metadata, null, 2) }}</pre>
      </UCard>
    </template>
  </div>
</template>
