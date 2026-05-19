<script setup lang="ts">
import type { Advertiser } from '~/types'

definePageMeta({ middleware: 'role' })

const { apiFetch } = useApi()

const advertisers = ref<Advertiser[]>([])
const loading = ref(true)
const errorMessage = ref('')
const showForm = ref(false)
const submitting = ref(false)
const formError = ref('')

const form = reactive({
  name: '',
  email: '',
})

async function loadAdvertisers() {
  loading.value = true
  try {
    advertisers.value = await apiFetch<Advertiser[]>('/api/v1/advertisers')
  } catch {
    errorMessage.value = 'Failed to load advertisers.'
  } finally {
    loading.value = false
  }
}

async function handleSubmit() {
  formError.value = ''
  submitting.value = true
  try {
    await apiFetch<Advertiser>('/api/v1/advertisers', {
      method: 'POST',
      body: { name: form.name, email: form.email },
    })
    form.name = ''
    form.email = ''
    showForm.value = false
    await loadAdvertisers()
  } catch {
    formError.value = 'Failed to create advertiser.'
  } finally {
    submitting.value = false
  }
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString()
}

onMounted(loadAdvertisers)
</script>

<template>
  <div>
    <div class="d-flex align-items-center justify-content-between mb-4">
      <h1 class="h4 fw-bold mb-0">Advertisers</h1>
      <button class="btn btn-primary btn-sm" @click="showForm = !showForm">
        {{ showForm ? 'Cancel' : 'Add Advertiser' }}
      </button>
    </div>

    <div v-if="showForm" class="card mb-4">
      <div class="card-header fw-semibold">New Advertiser</div>
      <div class="card-body">
        <form @submit.prevent="handleSubmit">
          <div class="row g-3 mb-3">
            <div class="col-md-6">
              <label class="form-label" for="name">Name <span class="text-danger">*</span></label>
              <input id="name" v-model="form.name" class="form-control" type="text" required />
            </div>
            <div class="col-md-6">
              <label class="form-label" for="email">Email <span class="text-danger">*</span></label>
              <input id="email" v-model="form.email" class="form-control" type="email" required />
            </div>
          </div>
          <div v-if="formError" class="alert alert-danger py-2 small">{{ formError }}</div>
          <div class="d-flex justify-content-end">
            <button class="btn btn-primary" type="submit" :disabled="submitting">
              {{ submitting ? 'Saving…' : 'Save Advertiser' }}
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
            <th>Email</th>
            <th>Created</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="advertiser in advertisers" :key="advertiser.id">
            <td>{{ advertiser.name }}</td>
            <td>{{ advertiser.email }}</td>
            <td>{{ formatDate(advertiser.created_at) }}</td>
          </tr>
          <tr v-if="advertisers.length === 0">
            <td colspan="3" class="text-center text-muted">No advertisers yet.</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
