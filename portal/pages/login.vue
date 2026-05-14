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
  <form class="login-form" @submit.prevent="handleSubmit">
    <div class="form-group">
      <label class="form-label" for="email">Email</label>
      <input
        id="email"
        v-model="email"
        class="form-input"
        type="email"
        required
        autocomplete="email"
        placeholder="you@example.com"
      />
    </div>
    <div class="form-group">
      <label class="form-label" for="password">Password</label>
      <input
        id="password"
        v-model="password"
        class="form-input"
        type="password"
        required
        autocomplete="current-password"
        placeholder="••••••••"
      />
    </div>
    <p v-if="errorMessage" class="form-error">{{ errorMessage }}</p>
    <button class="btn btn--primary btn--full" type="submit" :disabled="loading">
      {{ loading ? 'Signing in…' : 'Sign in' }}
    </button>
  </form>
</template>

<style scoped>
.login-form {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.form-label {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-text);
}

.form-input {
  width: 100%;
  padding: 0.625rem 0.75rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  font-size: 0.9375rem;
  color: var(--color-text);
  background-color: var(--color-surface);
  transition: border-color 0.15s;
}

.form-input:focus {
  outline: none;
  border-color: var(--color-accent);
}

.form-error {
  font-size: 0.875rem;
  color: var(--color-danger);
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.625rem 1.25rem;
  border-radius: var(--radius);
  font-size: 0.9375rem;
  font-weight: 500;
  border: 1px solid transparent;
  transition: opacity 0.15s;
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn--primary {
  background-color: var(--color-accent);
  color: #ffffff;
}

.btn--primary:hover:not(:disabled) {
  opacity: 0.9;
}

.btn--full {
  width: 100%;
}
</style>
