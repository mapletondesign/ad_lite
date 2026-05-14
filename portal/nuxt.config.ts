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
  css: ['~/assets/css/main.css'],
})
