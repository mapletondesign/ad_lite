import { useAuthStore } from '~/stores/auth'

export default defineNuxtRouteMiddleware((to) => {
  if (to.path.startsWith('/login')) return

  const auth = useAuthStore()

  if (!auth.isAuthenticated) {
    auth.loadFromStorage()
  }

  if (!auth.isAuthenticated) {
    return navigateTo('/login')
  }
})
