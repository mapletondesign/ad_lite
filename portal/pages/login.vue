<script setup lang="ts">
import { useAuthStore } from '~/stores/auth'
import type { TokenPair } from '~/types'

definePageMeta({ layout: 'auth' })

const auth = useAuthStore()
const config = useRuntimeConfig()

const email = ref('')
const password = ref('')
const loading = ref(false)
const errorMessage = ref('')

async function handleSubmit() {
  errorMessage.value = ''
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
    errorMessage.value = 'Invalid email or password.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <form @submit.prevent="handleSubmit">
    <div class="mb-3">
      <label class="form-label" for="email">Email</label>
      <input
        id="email"
        v-model="email"
        class="form-control"
        type="email"
        required
        autocomplete="email"
        placeholder="you@example.com"
      />
    </div>
    <div class="mb-3">
      <label class="form-label" for="password">Password</label>
      <input
        id="password"
        v-model="password"
        class="form-control"
        type="password"
        required
        autocomplete="current-password"
        placeholder="••••••••"
      />
    </div>
    <div v-if="errorMessage" class="alert alert-danger py-2 small">{{ errorMessage }}</div>
    <button class="btn btn-primary w-100" type="submit" :disabled="loading">
      {{ loading ? 'Signing in…' : 'Sign in' }}
    </button>
  </form>
</template>
