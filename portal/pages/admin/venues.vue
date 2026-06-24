<script setup lang="ts">
import type { Venue } from '~/types'

definePageMeta({ middleware: 'role' })

const { apiFetch } = useApi()

const venues     = ref<Venue[]>([])
const loading    = ref(true)
const error      = ref('')
const showForm   = ref(false)
const submitting = ref(false)
const formError  = ref('')
const snackbar   = ref(false)

const form = reactive({ name: '', category: '', address: '', city: '', state: '' })

const headers = [
  { title: 'Name',     key: 'name' },
  { title: 'Category', key: 'category' },
  { title: 'City',     key: 'city' },
  { title: 'State',    key: 'state' },
  { title: 'Created',  key: 'created_at' },
]

async function load() {
  loading.value = true
  try {
    venues.value = await apiFetch<Venue[]>('/api/v1/venues')
  } catch {
    error.value = 'Failed to load venues.'
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
        name:     form.name,
        category: form.category || null,
        address:  form.address  || null,
        city:     form.city     || null,
        state:    form.state    || null,
      },
    })
    Object.assign(form, { name: '', category: '', address: '', city: '', state: '' })
    showForm.value = false
    snackbar.value = true
    await load()
  } catch {
    formError.value = 'Failed to create venue.'
  } finally {
    submitting.value = false
  }
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString()
}

onMounted(load)
</script>

<template>
  <div>
    <div class="d-flex align-center justify-space-between mb-6">
      <div class="text-h5 font-weight-bold">Venues</div>
      <v-btn
        :prepend-icon="showForm ? 'mdi-close' : 'mdi-plus'"
        color="primary"
        variant="tonal"
        @click="showForm = !showForm"
      >
        {{ showForm ? 'Cancel' : 'Add Venue' }}
      </v-btn>
    </div>

    <v-expand-transition>
      <v-card v-if="showForm" class="mb-6" rounded="lg">
        <v-card-title class="text-subtitle-1 font-weight-semibold pa-4 pb-2">New Venue</v-card-title>
        <v-card-text>
          <v-form @submit.prevent="handleSubmit">
            <v-row>
              <v-col cols="12" md="6">
                <v-text-field v-model="form.name" label="Name" variant="outlined" density="comfortable" required />
              </v-col>
              <v-col cols="12" md="6">
                <v-text-field v-model="form.category" label="Category" variant="outlined" density="comfortable" placeholder="restaurant, gym, retail…" />
              </v-col>
              <v-col cols="12">
                <v-text-field v-model="form.address" label="Address" variant="outlined" density="comfortable" />
              </v-col>
              <v-col cols="12" md="6">
                <v-text-field v-model="form.city" label="City" variant="outlined" density="comfortable" />
              </v-col>
              <v-col cols="12" md="6">
                <v-text-field v-model="form.state" label="State" variant="outlined" density="comfortable" />
              </v-col>
            </v-row>
            <v-alert v-if="formError" type="error" density="compact" class="mb-4" rounded="lg">{{ formError }}</v-alert>
            <div class="d-flex justify-end">
              <v-btn type="submit" color="primary" :loading="submitting">Save Venue</v-btn>
            </div>
          </v-form>
        </v-card-text>
      </v-card>
    </v-expand-transition>

    <v-alert v-if="error" type="error" class="mb-4" rounded="lg">{{ error }}</v-alert>

    <v-data-table
      :headers="headers"
      :items="venues"
      :loading="loading"
      rounded="lg"
      hover
    >
      <template #item.category="{ item }">{{ item.category ?? '—' }}</template>
      <template #item.city="{ item }">{{ item.city ?? '—' }}</template>
      <template #item.state="{ item }">{{ item.state ?? '—' }}</template>
      <template #item.created_at="{ item }">{{ formatDate(item.created_at) }}</template>
      <template #no-data>No venues yet.</template>
    </v-data-table>

    <v-snackbar v-model="snackbar" color="success" timeout="3000">Venue created.</v-snackbar>
  </div>
</template>
