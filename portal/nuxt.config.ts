export default defineNuxtConfig({
  ssr: false,
  modules: ['@pinia/nuxt'],

  runtimeConfig: {
    public: {
      apiBase: 'http://localhost:8080',
    },
  },

  typescript: {
    strict: true,
  },

  app: {
    head: {
      htmlAttrs: { 'data-bs-theme': 'dark' },
    },
  },

  css: ['bootstrap/dist/css/bootstrap.min.css', '~/assets/css/main.css'],
  compatibilityDate: '2026-05-13',
})