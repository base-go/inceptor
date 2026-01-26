<script setup lang="ts">
const colorMode = useColorMode()

// Set dark mode as default
onMounted(() => {
  if (colorMode.preference === 'system') {
    colorMode.preference = 'dark'
  }
})
</script>

<template>
  <NuxtLayout>
    <NuxtPage />
  </NuxtLayout>
</template>
