<script setup lang="ts">
import type { Advertiser } from '~/types'

definePageMeta({ middleware: 'role' })

const { apiFetch } = useApi()

const advertisers = ref<Advertiser[]>([])
const loading     = ref(true)
const error       = ref('')
const showForm    = ref(false)
const submitting  = ref(false)
const formError   = ref('')
const snackbar    = ref(false)

const form = reactive({ name: '', email: '' })

const headers = [
  { title: 'Name',    key: 'name' },
  { title: 'Email',   key: 'email' },
  { title: 'Created', key: 'created_at' },
]

async function load() {
  loading.value = true
  try {
    advertisers.value = await apiFetch<Advertiser[]>('/api/v1/advertisers')
  } catch {
    error.value = 'Failed to load advertisers.'
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
    Object.assign(form, { name: '', email: '' })
    showForm.value = false
    snackbar.value = true
    await load()
  } catch {
    formError.value = 'Failed to create advertiser.'
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
      <div class="text-h5 font-weight-bold">Advertisers</div>
      <v-btn
        :prepend-icon="showForm ? 'mdi-close' : 'mdi-plus'"
        color="primary"
        variant="tonal"
        @click="showForm = !showForm"
      >
        {{ showForm ? 'Cancel' : 'Add Advertiser' }}
      </v-btn>
    </div>

    <v-expand-transition>
      <v-card v-if="showForm" class="mb-6" rounded="lg">
        <v-card-title class="text-subtitle-1 font-weight-semibold pa-4 pb-2">New Advertiser</v-card-title>
        <v-card-text>
          <v-form @submit.prevent="handleSubmit">
            <v-row>
              <v-col cols="12" md="6">
                <v-text-field v-model="form.name" label="Name" variant="outlined" density="comfortable" required />
              </v-col>
              <v-col cols="12" md="6">
                <v-text-field v-model="form.email" label="Email" type="email" variant="outlined" density="comfortable" required />
              </v-col>
            </v-row>
            <v-alert v-if="formError" type="error" density="compact" class="mb-4" rounded="lg">{{ formError }}</v-alert>
            <div class="d-flex justify-end">
              <v-btn type="submit" color="primary" :loading="submitting">Save Advertiser</v-btn>
            </div>
          </v-form>
        </v-card-text>
      </v-card>
    </v-expand-transition>

    <v-alert v-if="error" type="error" class="mb-4" rounded="lg">{{ error }}</v-alert>

    <v-data-table
      :headers="headers"
      :items="advertisers"
      :loading="loading"
      rounded="lg"
      hover
    >
      <template #item.created_at="{ item }">{{ formatDate(item.created_at) }}</template>
      <template #no-data>No advertisers yet.</template>
    </v-data-table>

    <v-snackbar v-model="snackbar" color="success" timeout="3000">Advertiser created.</v-snackbar>
  </div>
</template>
