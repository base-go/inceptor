<script setup lang="ts">
const route = useRoute()
const api = useApi()

const navigation = [
  { name: 'Dashboard', to: '/', icon: 'i-heroicons-home' },
  { name: 'Crashes', to: '/crashes', icon: 'i-heroicons-bug-ant' },
  { name: 'Groups', to: '/groups', icon: 'i-heroicons-rectangle-stack' },
  { name: 'Apps', to: '/apps', icon: 'i-heroicons-squares-2x2' },
  { name: 'Settings', to: '/settings', icon: 'i-heroicons-cog-6-tooth' },
]

const isActive = (path: string) => {
  if (path === '/') return route.path === '/'
  return route.path.startsWith(path)
}

// Load API key on mount
onMounted(() => {
  api.loadApiKey()
})
</script>

<template>
  <div class="min-h-screen bg-gray-900">
    <!-- Sidebar -->
    <div class="fixed inset-y-0 left-0 z-50 w-64 bg-gray-800 border-r border-gray-700">
      <!-- Logo -->
      <div class="flex items-center h-16 px-6 border-b border-gray-700">
        <span class="text-xl font-bold text-white">
          <span class="text-primary-500">In</span>ceptor
        </span>
      </div>

      <!-- Navigation -->
      <nav class="p-4 space-y-1">
        <NuxtLink
          v-for="item in navigation"
          :key="item.name"
          :to="item.to"
          :class="[
            'flex items-center px-4 py-2.5 text-sm font-medium rounded-lg transition-colors',
            isActive(item.to)
              ? 'bg-primary-500/10 text-primary-500'
              : 'text-gray-400 hover:bg-gray-700 hover:text-white',
          ]"
        >
          <UIcon :name="item.icon" class="w-5 h-5 mr-3" />
          {{ item.name }}
        </NuxtLink>
      </nav>

      <!-- API Key Status -->
      <div class="absolute bottom-0 left-0 right-0 p-4 border-t border-gray-700">
        <div class="flex items-center text-sm">
          <UIcon
            :name="api.apiKey.value ? 'i-heroicons-check-circle' : 'i-heroicons-exclamation-circle'"
            :class="[
              'w-5 h-5 mr-2',
              api.apiKey.value ? 'text-green-500' : 'text-yellow-500',
            ]"
          />
          <span class="text-gray-400">
            {{ api.apiKey.value ? 'API Connected' : 'No API Key' }}
          </span>
        </div>
      </div>
    </div>

    <!-- Main content -->
    <div class="pl-64">
      <!-- Header -->
      <header class="sticky top-0 z-40 flex items-center h-16 px-6 bg-gray-900/80 backdrop-blur border-b border-gray-800">
        <h1 class="text-lg font-semibold text-white">
          {{ route.meta.title || 'Dashboard' }}
        </h1>
        <div class="flex-1" />
        <UButton
          icon="i-heroicons-sun"
          color="gray"
          variant="ghost"
          class="hidden"
        />
      </header>

      <!-- Page content -->
      <main class="p-6">
        <slot />
      </main>
    </div>
  </div>
</template>
