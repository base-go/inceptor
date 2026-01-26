<script setup lang="ts">
import type { CrashGroup } from '~/types'

definePageMeta({
  title: 'Crash Groups',
})

const api = useApi()

const groups = ref<CrashGroup[]>([])
const total = ref(0)
const loading = ref(true)
const page = ref(1)
const limit = 20

const search = ref('')
const status = ref('')
const sortBy = ref('last_seen')
const sortOrder = ref('desc')

const statusOptions = [
  { label: 'All Statuses', value: '' },
  { label: 'Open', value: 'open' },
  { label: 'Resolved', value: 'resolved' },
  { label: 'Ignored', value: 'ignored' },
]

const sortOptions = [
  { label: 'Last Seen', value: 'last_seen' },
  { label: 'First Seen', value: 'first_seen' },
  { label: 'Occurrences', value: 'occurrence_count' },
]

const columns = [
  { key: 'status', label: 'Status' },
  { key: 'error_type', label: 'Error Type' },
  { key: 'error_message', label: 'Message' },
  { key: 'occurrence_count', label: 'Count' },
  { key: 'first_seen', label: 'First Seen' },
  { key: 'last_seen', label: 'Last Seen' },
  { key: 'actions', label: '' },
]

const loadGroups = async () => {
  if (!api.isAuthenticated.value) {
    loading.value = false
    return
  }

  try {
    loading.value = true
    const response = await api.getGroups({
      search: search.value || undefined,
      status: status.value || undefined,
      sort_by: sortBy.value,
      sort_order: sortOrder.value,
      limit,
      offset: (page.value - 1) * limit,
    })
    groups.value = response.data || []
    total.value = response.total
  } catch (e) {
    console.error('Failed to load groups:', e)
  } finally {
    loading.value = false
  }
}

watch([search, status, sortBy, sortOrder], () => {
  page.value = 1
  loadGroups()
})

watch(page, loadGroups)

onMounted(loadGroups)

const formatDate = (date: string) => {
  return new Date(date).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const formatNumber = (num: number) => {
  return new Intl.NumberFormat().format(num)
}

const statusColor = (s: string) => {
  switch (s) {
    case 'open':
      return 'red'
    case 'resolved':
      return 'green'
    case 'ignored':
      return 'gray'
    default:
      return 'gray'
  }
}

const updateStatus = async (group: CrashGroup, newStatus: string) => {
  try {
    await api.updateGroup(group.id, { status: newStatus as any })
    group.status = newStatus as any
  } catch (e) {
    console.error('Failed to update status:', e)
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
          placeholder="Search errors..."
          icon="i-heroicons-magnifying-glass"
          class="w-64"
        />
        <USelectMenu
          v-model="status"
          :options="statusOptions"
          option-attribute="label"
          value-attribute="value"
          class="w-40"
        />
        <USelectMenu
          v-model="sortBy"
          :options="sortOptions"
          option-attribute="label"
          value-attribute="value"
          class="w-40"
        />
        <UButton
          :icon="sortOrder === 'desc' ? 'i-heroicons-bars-arrow-down' : 'i-heroicons-bars-arrow-up'"
          variant="ghost"
          @click="sortOrder = sortOrder === 'desc' ? 'asc' : 'desc'"
        />
      </div>
    </UCard>

    <!-- Table -->
    <UCard>
      <UTable
        :rows="groups"
        :columns="columns"
        :loading="loading"
        :empty-state="{ icon: 'i-heroicons-rectangle-stack', label: 'No crash groups found' }"
      >
        <template #status-data="{ row }">
          <UDropdown
            :items="[
              [
                { label: 'Open', click: () => updateStatus(row, 'open') },
                { label: 'Resolved', click: () => updateStatus(row, 'resolved') },
                { label: 'Ignored', click: () => updateStatus(row, 'ignored') },
              ],
            ]"
          >
            <UBadge :color="statusColor(row.status)" class="cursor-pointer">
              {{ row.status }}
              <UIcon name="i-heroicons-chevron-down" class="w-3 h-3 ml-1" />
            </UBadge>
          </UDropdown>
        </template>

        <template #error_type-data="{ row }">
          <NuxtLink :to="`/groups/${row.id}`" class="text-primary-500 hover:underline font-medium">
            {{ row.error_type }}
          </NuxtLink>
        </template>

        <template #error_message-data="{ row }">
          <span class="text-gray-400 truncate max-w-xs block">
            {{ row.error_message }}
          </span>
        </template>

        <template #occurrence_count-data="{ row }">
          <UBadge color="red">{{ formatNumber(row.occurrence_count) }}</UBadge>
        </template>

        <template #first_seen-data="{ row }">
          <span class="text-gray-400 text-sm">
            {{ formatDate(row.first_seen) }}
          </span>
        </template>

        <template #last_seen-data="{ row }">
          <span class="text-gray-400 text-sm">
            {{ formatDate(row.last_seen) }}
          </span>
        </template>

        <template #actions-data="{ row }">
          <UDropdown
            :items="[[
              { label: 'View Details', icon: 'i-heroicons-eye', click: () => navigateTo(`/groups/${row.id}`) },
              { label: 'View Crashes', icon: 'i-heroicons-bug-ant', click: () => navigateTo(`/crashes?group_id=${row.id}`) },
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
