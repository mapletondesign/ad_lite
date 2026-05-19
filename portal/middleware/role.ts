import { useAuthStore } from '~/stores/auth'

export default defineNuxtRouteMiddleware((to) => {
  const auth = useAuthStore()
  const role = auth.user?.role

  if (to.path.startsWith('/admin') && role !== 'admin') {
    return navigateTo('/login')
  }

  if (to.path.startsWith('/advertiser') && role !== 'advertiser' && role !== 'admin') {
    return navigateTo('/login')
  }

  if (to.path.startsWith('/venue') && role !== 'venue' && role !== 'admin') {
    return navigateTo('/login')
  }
})
