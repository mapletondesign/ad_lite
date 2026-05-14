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
  <div>
    <div class="d-flex align-items-center justify-content-between mb-4">
      <h1 class="h4 fw-bold mb-0">Venues</h1>
      <button class="btn btn-primary btn-sm" @click="showForm = !showForm">
        {{ showForm ? 'Cancel' : 'Add Venue' }}
      </button>
    </div>

    <div v-if="showForm" class="card mb-4">
      <div class="card-header fw-semibold">New Venue</div>
      <div class="card-body">
        <form @submit.prevent="handleSubmit">
          <div class="row g-3 mb-3">
            <div class="col-md-6">
              <label class="form-label" for="name">Name <span class="text-danger">*</span></label>
              <input id="name" v-model="form.name" class="form-control" type="text" required />
            </div>
            <div class="col-md-6">
              <label class="form-label" for="category">Category</label>
              <input id="category" v-model="form.category" class="form-control" type="text" placeholder="restaurant, gym, retail…" />
            </div>
          </div>
          <div class="mb-3">
            <label class="form-label" for="address">Address</label>
            <input id="address" v-model="form.address" class="form-control" type="text" />
          </div>
          <div class="row g-3 mb-3">
            <div class="col-md-6">
              <label class="form-label" for="city">City</label>
              <input id="city" v-model="form.city" class="form-control" type="text" />
            </div>
            <div class="col-md-6">
              <label class="form-label" for="state">State</label>
              <input id="state" v-model="form.state" class="form-control" type="text" />
            </div>
          </div>
          <div v-if="formError" class="alert alert-danger py-2 small">{{ formError }}</div>
          <div class="d-flex justify-content-end">
            <button class="btn btn-primary" type="submit" :disabled="submitting">
              {{ submitting ? 'Saving…' : 'Save Venue' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <div v-if="errorMessage" class="alert alert-danger">{{ errorMessage }}</div>

    <div class="card" :class="{ 'opacity-50': loading }">
      <div v-if="loading" class="card-body text-muted">Loading…</div>
      <table v-else class="table table-dark table-hover mb-0">
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
            <td colspan="5" class="text-center text-muted">No venues yet.</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
