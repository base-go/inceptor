<script setup lang="ts">
import type { Crash } from '~/types'

definePageMeta({
  title: 'Crashes',
})

const api = useApi()

const crashes = ref<Crash[]>([])
const total = ref(0)
const loading = ref(true)
const page = ref(1)
const limit = 20

const search = ref('')
const platform = ref('')
const environment = ref('')

const platformOptions = [
  { label: 'All Platforms', value: '' },
  { label: 'iOS', value: 'ios' },
  { label: 'Android', value: 'android' },
  { label: 'Web', value: 'web' },
  { label: 'Flutter', value: 'flutter' },
]

const environmentOptions = [
  { label: 'All Environments', value: '' },
  { label: 'Production', value: 'production' },
  { label: 'Staging', value: 'staging' },
  { label: 'Development', value: 'development' },
]

const columns = [
  { key: 'error_type', label: 'Error Type' },
  { key: 'error_message', label: 'Message' },
  { key: 'platform', label: 'Platform' },
  { key: 'app_version', label: 'Version' },
  { key: 'environment', label: 'Env' },
  { key: 'created_at', label: 'Time' },
  { key: 'actions', label: '' },
]

const loadCrashes = async () => {
  if (!api.apiKey.value) {
    loading.value = false
    return
  }

  try {
    loading.value = true
    const response = await api.getCrashes({
      search: search.value || undefined,
      platform: platform.value || undefined,
      environment: environment.value || undefined,
      limit,
      offset: (page.value - 1) * limit,
    })
    crashes.value = response.data || []
    total.value = response.total
  } catch (e) {
    console.error('Failed to load crashes:', e)
  } finally {
    loading.value = false
  }
}

watch([search, platform, environment], () => {
  page.value = 1
  loadCrashes()
})

watch(page, loadCrashes)

onMounted(loadCrashes)

const formatDate = (date: string) => {
  return new Date(date).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const platformIcon = (p: string) => {
  switch (p) {
    case 'ios':
      return 'i-simple-icons-apple'
    case 'android':
      return 'i-simple-icons-android'
    case 'web':
      return 'i-heroicons-globe-alt'
    case 'flutter':
      return 'i-simple-icons-flutter'
    default:
      return 'i-heroicons-device-phone-mobile'
  }
}

const handleDelete = async (crash: Crash) => {
  if (!confirm('Are you sure you want to delete this crash?')) return

  try {
    await api.deleteCrash(crash.id)
    loadCrashes()
  } catch (e) {
    console.error('Failed to delete crash:', e)
  }
}
</script>

<template>
  <div class="space-y-6">
    <!-- Filters -->
    <UCard>
      <div class="flex flex-wrap gap-4">
        <UInput
          v-model="search"
          placeholder="Search crashes..."
          icon="i-heroicons-magnifying-glass"
          class="w-64"
        />
        <USelectMenu
          v-model="platform"
          :options="platformOptions"
          option-attribute="label"
          value-attribute="value"
          class="w-40"
        />
        <USelectMenu
          v-model="environment"
          :options="environmentOptions"
          option-attribute="label"
          value-attribute="value"
          class="w-40"
        />
      </div>
    </UCard>

    <!-- Table -->
    <UCard>
      <UTable
        :rows="crashes"
        :columns="columns"
        :loading="loading"
        :empty-state="{ icon: 'i-heroicons-bug-ant', label: 'No crashes found' }"
      >
        <template #error_type-data="{ row }">
          <NuxtLink :to="`/crashes/${row.id}`" class="text-primary-500 hover:underline font-medium">
            {{ row.error_type }}
          </NuxtLink>
        </template>

        <template #error_message-data="{ row }">
          <span class="text-gray-400 truncate max-w-xs block">
            {{ row.error_message }}
          </span>
        </template>

        <template #platform-data="{ row }">
          <div class="flex items-center gap-2">
            <UIcon :name="platformIcon(row.platform)" class="w-4 h-4" />
            <span class="capitalize">{{ row.platform }}</span>
          </div>
        </template>

        <template #environment-data="{ row }">
          <UBadge
            :color="row.environment === 'production' ? 'red' : row.environment === 'staging' ? 'yellow' : 'gray'"
            size="xs"
          >
            {{ row.environment }}
          </UBadge>
        </template>

        <template #created_at-data="{ row }">
          <span class="text-gray-400 text-sm">
            {{ formatDate(row.created_at) }}
          </span>
        </template>

        <template #actions-data="{ row }">
          <UDropdown
            :items="[[
              { label: 'View Details', icon: 'i-heroicons-eye', click: () => navigateTo(`/crashes/${row.id}`) },
              { label: 'View Group', icon: 'i-heroicons-rectangle-stack', click: () => navigateTo(`/groups/${row.group_id}`) },
              { label: 'Delete', icon: 'i-heroicons-trash', click: () => handleDelete(row) },
            ]]"
          >
            <UButton icon="i-heroicons-ellipsis-vertical" variant="ghost" color="gray" />
          </UDropdown>
        </template>
      </UTable>

      <!-- Pagination -->
      <div v-if="total > limit" class="flex justify-center mt-4">
        <UPagination
          v-model="page"
          :total="total"
          :page-count="limit"
        />
      </div>
    </UCard>
  </div>
</template>
