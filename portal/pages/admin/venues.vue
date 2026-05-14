<script setup lang="ts">
import type { Venue } from '~/types'

definePageMeta({ middleware: 'role' })

const { apiFetch } = useApi()

const venues = ref<Venue[]>([])
const loading = ref(true)
const errorMessage = ref('')
const showForm = ref(false)
const submitting = ref(false)
const formError = ref('')

const form = reactive({
  name: '',
  category: '',
  address: '',
  city: '',
  state: '',
})

async function loadVenues() {
  loading.value = true
  try {
    venues.value = await apiFetch<Venue[]>('/api/v1/venues')
  } catch {
    errorMessage.value = 'Failed to load venues.'
  } finally {
    loading.value = false
  }
}

async function handleSubmit() {
  formError.value = ''
  submitting.value = true
  try {
    await apiFetch<Venue>('/api/v1/venues', {
      method: 'POST',
      body: {
        name: form.name,
        category: form.category || null,
        address: form.address || null,
        city: form.city || null,
        state: form.state || null,
      },
    })
    form.name = ''
    form.category = ''
    form.address = ''
    form.city = ''
    form.state = ''
    showForm.value = false
    await loadVenues()
  } catch {
    formError.value = 'Failed to create venue.'
  } finally {
    submitting.value = false
  }
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString()
}

onMounted(loadVenues)
</script>

<template>
  <div class="page">
    <div class="page__header">
      <h1 class="page__title">Venues</h1>
      <button class="btn btn--primary" @click="showForm = !showForm">
        {{ showForm ? 'Cancel' : 'Add Venue' }}
      </button>
    </div>

    <div v-if="showForm" class="inline-form">
      <h2 class="inline-form__title">New Venue</h2>
      <form @submit.prevent="handleSubmit">
        <div class="form-row">
          <div class="form-group">
            <label class="form-label" for="name">Name <span class="required">*</span></label>
            <input id="name" v-model="form.name" class="form-input" type="text" required />
          </div>
          <div class="form-group">
            <label class="form-label" for="category">Category</label>
            <input id="category" v-model="form.category" class="form-input" type="text" />
          </div>
        </div>
        <div class="form-group">
          <label class="form-label" for="address">Address</label>
          <input id="address" v-model="form.address" class="form-input" type="text" />
        </div>
        <div class="form-row">
          <div class="form-group">
            <label class="form-label" for="city">City</label>
            <input id="city" v-model="form.city" class="form-input" type="text" />
          </div>
          <div class="form-group">
            <label class="form-label" for="state">State</label>
            <input id="state" v-model="form.state" class="form-input" type="text" />
          </div>
        </div>
        <p v-if="formError" class="form-error">{{ formError }}</p>
        <div class="form-actions">
          <button class="btn btn--primary" type="submit" :disabled="submitting">
            {{ submitting ? 'Saving…' : 'Save Venue' }}
          </button>
        </div>
      </form>
    </div>

    <p v-if="errorMessage" class="error-text">{{ errorMessage }}</p>

    <div class="table-wrap" :class="{ loading: loading }">
      <p v-if="loading" class="loading-text">Loading…</p>
      <table v-else class="data-table">
        <thead>
          <tr>
            <th>Name</th>
            <th>Category</th>
            <th>City</th>
            <th>State</th>
            <th>Created</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="venue in venues" :key="venue.id">
            <td>{{ venue.name }}</td>
            <td>{{ venue.category ?? '—' }}</td>
            <td>{{ venue.city ?? '—' }}</td>
            <td>{{ venue.state ?? '—' }}</td>
            <td>{{ formatDate(venue.created_at) }}</td>
          </tr>
          <tr v-if="venues.length === 0">
            <td colspan="5" class="empty-cell">No venues yet.</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.page {
  max-width: 1100px;
}

.page__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.5rem;
}

.page__title {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--color-text);
}

.error-text {
  color: var(--color-danger);
  font-size: 0.875rem;
  margin-bottom: 1rem;
}

.inline-form {
  background-color: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 1.5rem;
  margin-bottom: 1.5rem;
}

.inline-form__title {
  font-size: 1rem;
  font-weight: 600;
  margin-bottom: 1rem;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
  margin-bottom: 1rem;
}

.form-label {
  font-size: 0.875rem;
  font-weight: 500;
}

.required {
  color: var(--color-danger);
}

.form-input {
  width: 100%;
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  font-size: 0.9375rem;
  color: var(--color-text);
  background-color: var(--color-surface);
}

.form-input:focus {
  outline: none;
  border-color: var(--color-accent);
}

.form-error {
  font-size: 0.875rem;
  color: var(--color-danger);
  margin-bottom: 0.75rem;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
}

.btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.5rem 1.25rem;
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

.table-wrap {
  background-color: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  overflow: hidden;
}

.table-wrap.loading {
  opacity: 0.5;
}

.loading-text {
  padding: 1.5rem;
  color: var(--color-text-muted);
  font-size: 0.9375rem;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th {
  background-color: var(--color-primary);
  color: #ffffff;
  text-align: left;
  padding: 0.75rem 1rem;
  font-size: 0.8125rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.data-table td {
  padding: 0.75rem 1rem;
  font-size: 0.9375rem;
  border-bottom: 1px solid var(--color-border);
  color: var(--color-text);
}

.data-table tbody tr:nth-child(even) td {
  background-color: #f8fafc;
}

.data-table tbody tr:last-child td {
  border-bottom: none;
}

.empty-cell {
  color: var(--color-text-muted);
  text-align: center;
}
</style>
