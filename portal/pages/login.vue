<script setup lang="ts">
import { useAuthStore } from '~/stores/auth'
import type { TokenPair } from '~/types'

definePageMeta({ layout: 'auth' })

const auth = useAuthStore()
const config = useRuntimeConfig()

const email    = ref('')
const password = ref('')
const loading  = ref(false)
const error    = ref('')

async function handleSubmit() {
  error.value = ''
  loading.value = true
  try {
    const result = await $fetch<TokenPair>('/api/v1/auth/login', {
      baseURL: config.public.apiBase,
      method: 'POST',
      body: { email: email.value, password: password.value },
    })
    auth.setTokens(result.access_token, result.refresh_token)
    await navigateTo('/')
  } catch {
    error.value = 'Invalid email or password.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <v-form @submit.prevent="handleSubmit">
    <v-text-field
      v-model="email"
      label="Email"
      type="email"
      variant="outlined"
      density="comfortable"
      autocomplete="email"
      required
      class="mb-2"
    />
    <v-text-field
      v-model="password"
      label="Password"
      type="password"
      variant="outlined"
      density="comfortable"
      autocomplete="current-password"
      required
      class="mb-4"
    />
    <v-alert v-if="error" type="error" density="compact" class="mb-4" rounded="lg">
      {{ error }}
    </v-alert>
    <v-btn type="submit" color="primary" size="large" block :loading="loading">
      Sign in
    </v-btn>
  </v-form>
</template>
