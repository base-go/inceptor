// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  devtools: { enabled: true },

  modules: ['@nuxt/ui'],

  ui: {
    icons: ['heroicons', 'simple-icons'],
  },

  colorMode: {
    preference: 'dark',
  },

  app: {
    head: {
      title: 'Inceptor - Crash Logging Dashboard',
      meta: [
        { name: 'description', content: 'Self-hosted crash logging and error tracking' },
      ],
    },
  },

  runtimeConfig: {
    public: {
      apiBase: process.env.NUXT_PUBLIC_API_BASE || '/api/v1',
    },
  },

  typescript: {
    strict: true,
  },

  compatibilityDate: '2024-01-01',
})
