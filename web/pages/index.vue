<script setup lang="ts">
import type { CrashStats, CrashGroup } from '~/types'

definePageMeta({
  title: 'Dashboard',
})

const api = useApi()

const selectedAppId = ref<string>('')
const stats = ref<CrashStats | null>(null)
const recentGroups = ref<CrashGroup[]>([])
const loading = ref(true)
const error = ref<string | null>(null)

const apps = ref<{ id: string; name: string }[]>([])

// Login state
const password = ref('')
const loginError = ref<string | null>(null)
const loginLoading = ref(false)

// Password change state
const showPasswordChange = ref(false)
const oldPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const passwordChangeError = ref<string | null>(null)
const passwordChangeLoading = ref(false)

const handleLogin = async () => {
  if (!password.value) return

  loginLoading.value = true
  loginError.value = null

  try {
    await api.login(password.value)
    password.value = ''

    // Check if password change is needed
    if (api.needsPasswordChange.value) {
      showPasswordChange.value = true
    } else {
      loadData()
    }
  } catch (e: any) {
    loginError.value = e.data?.error || 'Invalid password'
  } finally {
    loginLoading.value = false
  }
}

const handlePasswordChange = async () => {
  if (!oldPassword.value || !newPassword.value) return
  if (newPassword.value !== confirmPassword.value) {
    passwordChangeError.value = 'Passwords do not match'
    return
  }
  if (newPassword.value.length < 4) {
    passwordChangeError.value = 'Password must be at least 4 characters'
    return
  }

  passwordChangeLoading.value = true
  passwordChangeError.value = null

  try {
    await api.changePassword(oldPassword.value, newPassword.value)
    showPasswordChange.value = false
    oldPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
    loadData()
  } catch (e: any) {
    passwordChangeError.value = e.data?.error || 'Failed to change password'
  } finally {
    passwordChangeLoading.value = false
  }
}

const loadData = async () => {
  if (!api.isAuthenticated.value) {
    loading.value = false
    return
  }

  try {
    loading.value = true
    error.value = null

    // Load apps first
    const appsResponse = await api.getApps()
    apps.value = appsResponse.data || []

    if (apps.value.length > 0 && !selectedAppId.value) {
      selectedAppId.value = apps.value[0].id
    }

    if (selectedAppId.value) {
      // Load stats
      stats.value = await api.getAppStats(selectedAppId.value)

      // Load recent groups
      const groupsResponse = await api.getGroups({
        app_id: selectedAppId.value,
        limit: 5,
        sort_by: 'last_seen',
        sort_order: 'desc',
      })
      recentGroups.value = groupsResponse.data || []
    }
  } catch (e: any) {
    error.value = e.message || 'Failed to load data'
  } finally {
    loading.value = false
  }
}

watch(selectedAppId, () => loadData())

onMounted(() => {
  // Check if password change needed after token load
  if (api.needsPasswordChange.value) {
    showPasswordChange.value = true
  }
  loadData()
})

const formatNumber = (num: number) => {
  return new Intl.NumberFormat().format(num)
}

const formatDate = (date: string) => {
  return new Date(date).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

const statusColor = (status: string) => {
  switch (status) {
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
</script>

<template>
  <div class="space-y-6">
    <!-- Password Change Modal -->
    <UModal v-model="showPasswordChange" :prevent-close="api.needsPasswordChange.value">
      <UCard>
        <template #header>
          <h3 class="text-lg font-semibold">Change Password</h3>
          <p v-if="api.needsPasswordChange.value" class="text-sm text-gray-400 mt-1">
            Please change your default password to continue.
          </p>
        </template>

        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-300 mb-1">Current Password</label>
            <UInput
              v-model="oldPassword"
              type="password"
              placeholder="Enter current password"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-300 mb-1">New Password</label>
            <UInput
              v-model="newPassword"
              type="password"
              placeholder="Enter new password (min 4 chars)"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-gray-300 mb-1">Confirm Password</label>
            <UInput
              v-model="confirmPassword"
              type="password"
              placeholder="Confirm new password"
              @keyup.enter="handlePasswordChange"
            />
          </div>
          <UAlert v-if="passwordChangeError" color="red" :title="passwordChangeError" />
        </div>

        <template #footer>
          <div class="flex justify-end gap-2">
            <UButton
              v-if="!api.needsPasswordChange.value"
              color="gray"
              @click="showPasswordChange = false"
            >
              Cancel
            </UButton>
            <UButton
              :loading="passwordChangeLoading"
              @click="handlePasswordChange"
            >
              Change Password
            </UButton>
          </div>
        </template>
      </UCard>
    </UModal>

    <!-- Login Form if not authenticated -->
    <UCard v-if="!api.isAuthenticated.value" class="max-w-md">
      <template #header>
        <h3 class="text-lg font-semibold">Login</h3>
      </template>
      <div class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-300 mb-1">Password</label>
          <UInput
            v-model="password"
            type="password"
            placeholder="Enter password"
            @keyup.enter="handleLogin"
          />
        </div>
        <UAlert v-if="loginError" color="red" :title="loginError" />
        <UButton
          :loading="loginLoading"
          block
          @click="handleLogin"
        >
          Login
        </UButton>
      </div>
      <template #footer>
        <p class="text-sm text-gray-500">
          Default password: inceptor
        </p>
      </template>
    </UCard>

    <template v-else>
      <!-- App Selector -->
      <div class="flex items-center gap-4">
        <USelectMenu
          v-model="selectedAppId"
          :options="apps"
          option-attribute="name"
          value-attribute="id"
          placeholder="Select an app"
          class="w-64"
        />
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="flex items-center justify-center py-12">
        <UIcon name="i-heroicons-arrow-path" class="w-8 h-8 animate-spin text-primary-500" />
      </div>

      <!-- Error State -->
      <UAlert v-else-if="error" color="red" :title="error" />

      <!-- Stats Cards -->
      <template v-else-if="stats">
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
          <UCard>
            <div class="flex items-center">
              <div class="p-3 rounded-lg bg-red-500/10">
                <UIcon name="i-heroicons-bug-ant" class="w-6 h-6 text-red-500" />
              </div>
              <div class="ml-4">
                <p class="text-sm text-gray-400">Total Crashes</p>
                <p class="text-2xl font-bold text-white">{{ formatNumber(stats.total_crashes) }}</p>
              </div>
            </div>
          </UCard>

          <UCard>
            <div class="flex items-center">
              <div class="p-3 rounded-lg bg-yellow-500/10">
                <UIcon name="i-heroicons-exclamation-triangle" class="w-6 h-6 text-yellow-500" />
              </div>
              <div class="ml-4">
                <p class="text-sm text-gray-400">Open Issues</p>
                <p class="text-2xl font-bold text-white">{{ formatNumber(stats.open_groups) }}</p>
              </div>
            </div>
          </UCard>

          <UCard>
            <div class="flex items-center">
              <div class="p-3 rounded-lg bg-blue-500/10">
                <UIcon name="i-heroicons-clock" class="w-6 h-6 text-blue-500" />
              </div>
              <div class="ml-4">
                <p class="text-sm text-gray-400">Last 24h</p>
                <p class="text-2xl font-bold text-white">{{ formatNumber(stats.crashes_last_24h) }}</p>
              </div>
            </div>
          </UCard>

          <UCard>
            <div class="flex items-center">
              <div class="p-3 rounded-lg bg-green-500/10">
                <UIcon name="i-heroicons-chart-bar" class="w-6 h-6 text-green-500" />
              </div>
              <div class="ml-4">
                <p class="text-sm text-gray-400">Last 7 Days</p>
                <p class="text-2xl font-bold text-white">{{ formatNumber(stats.crashes_last_7d) }}</p>
              </div>
            </div>
          </UCard>
        </div>

        <!-- Recent Issues -->
        <UCard>
          <template #header>
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-semibold">Recent Issues</h3>
              <NuxtLink to="/groups">
                <UButton variant="ghost" size="sm">View All</UButton>
              </NuxtLink>
            </div>
          </template>

          <div v-if="recentGroups.length === 0" class="text-center py-8 text-gray-500">
            No crash groups yet
          </div>

          <div v-else class="space-y-3">
            <NuxtLink
              v-for="group in recentGroups"
              :key="group.id"
              :to="`/groups/${group.id}`"
              class="block p-4 rounded-lg bg-gray-800/50 hover:bg-gray-800 transition-colors"
            >
              <div class="flex items-start justify-between">
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-2">
                    <UBadge :color="statusColor(group.status)" size="xs">
                      {{ group.status }}
                    </UBadge>
                    <span class="text-sm font-medium text-white truncate">
                      {{ group.error_type }}
                    </span>
                  </div>
                  <p class="mt-1 text-sm text-gray-400 truncate">
                    {{ group.error_message }}
                  </p>
                </div>
                <div class="text-right ml-4">
                  <p class="text-sm font-medium text-white">
                    {{ formatNumber(group.occurrence_count) }}
                  </p>
                  <p class="text-xs text-gray-500">
                    {{ formatDate(group.last_seen) }}
                  </p>
                </div>
              </div>
            </NuxtLink>
          </div>
        </UCard>

        <!-- Top Errors -->
        <UCard v-if="stats.top_errors?.length">
          <template #header>
            <h3 class="text-lg font-semibold">Top Errors</h3>
          </template>

          <UTable
            :rows="stats.top_errors"
            :columns="[
              { key: 'error_type', label: 'Error Type' },
              { key: 'error_message', label: 'Message' },
              { key: 'count', label: 'Count' },
            ]"
          >
            <template #error_message-data="{ row }">
              <span class="text-gray-400 truncate max-w-xs block">
                {{ row.error_message }}
              </span>
            </template>
            <template #count-data="{ row }">
              <UBadge color="red">{{ formatNumber(row.count) }}</UBadge>
            </template>
          </UTable>
        </UCard>
      </template>
    </template>
  </div>
</template>
