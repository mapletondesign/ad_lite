export default defineNuxtConfig({
  ssr: false,
  modules: ['@pinia/nuxt', 'vuetify-nuxt-module'],

  vuetify: {
    vuetifyOptions: {
      theme: {
        defaultTheme: 'dark',
        themes: {
          dark: {
            dark: true,
            colors: {
              primary: '#2979FF',
              secondary: '#546E7A',
              success: '#4CAF50',
              warning: '#FFC107',
              error: '#FF5252',
              info: '#29B6F6',
              surface: '#1E1E2E',
              background: '#121212',
            },
          },
        },
      },
      icons: {
        defaultSet: 'mdi',
      },
    },
  },

  css: ['@mdi/font/css/materialdesignicons.min.css', '~/assets/css/main.css'],

  runtimeConfig: {
    public: {
      apiBase: 'http://localhost:8080',
    },
  },

  typescript: {
    strict: true,
  },

  compatibilityDate: '2026-05-13',
})
