<script setup lang="ts">
import { useAuthStore } from '~/stores/auth'

const auth = useAuthStore()

onMounted(() => {
  if (!auth.isAuthenticated) {
    auth.loadFromStorage()
  }

  const role = auth.user?.role
  if (role === 'admin') {
    navigateTo('/admin')
  } else if (role === 'advertiser') {
    navigateTo('/advertiser')
  } else if (role === 'venue') {
    navigateTo('/venue')
  } else {
    navigateTo('/login')
  }
})
</script>

<template>
  <div />
</template>
