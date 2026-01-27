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

  // Explicitly set empty public config to prevent any defaults
  runtimeConfig: {
    public: {}
  },

  typescript: {
    strict: true,
  },

  compatibilityDate: '2024-01-01',
})
