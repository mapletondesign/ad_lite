<script setup lang="ts">
import type { AdSlot } from '~/types'

definePageMeta({ middleware: 'role' })

const { apiFetch } = useApi()

const slots      = ref<AdSlot[]>([])
const loading    = ref(true)
const error      = ref('')
const showForm   = ref(false)
const submitting = ref(false)
const formError  = ref('')
const snackbar   = ref(false)

const DAY_LABELS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

const form = reactive({
  device_id:   '',
  label:       '',
  days_of_week: [] as number[],
  start_time:  '',
  end_time:    '',
  duration_sec: 30,
  price_cents:  0,
})

const headers = [
  { title: 'Label',       key: 'label' },
  { title: 'Device ID',   key: 'device_id' },
  { title: 'Days',        key: 'days_of_week' },
  { title: 'Time Window', key: 'window' },
  { title: 'Duration',    key: 'duration_sec' },
  { title: 'Price',       key: 'price_cents' },
  { title: 'Status',      key: 'status' },
]

const slotStatusColor: Record<string, string> = {
  available: 'success',
  booked:    'primary',
  paused:    'warning',
}

async function load() {
  loading.value = true
  try {
    slots.value = await apiFetch<AdSlot[]>('/api/v1/slots')
  } catch {
    error.value = 'Failed to load slots.'
  } finally {
    loading.value = false
  }
}

async function handleSubmit() {
  formError.value = ''
  submitting.value = true
  try {
    await apiFetch<AdSlot>('/api/v1/slots', {
      method: 'POST',
      body: {
        device_id:    form.device_id,
        label:        form.label || null,
        days_of_week: form.days_of_week,
        start_time:   form.start_time,
        end_time:     form.end_time,
        duration_sec: Number(form.duration_sec),
        price_cents:  Number(form.price_cents),
      },
    })
    Object.assign(form, { device_id: '', label: '', days_of_week: [], start_time: '', end_time: '', duration_sec: 30, price_cents: 0 })
    showForm.value = false
    snackbar.value = true
    await load()
  } catch {
    formError.value = 'Failed to create slot.'
  } finally {
    submitting.value = false
  }
}

function toggleDay(day: number) {
  const idx = form.days_of_week.indexOf(day)
  if (idx === -1) form.days_of_week.push(day)
  else form.days_of_week.splice(idx, 1)
}

function formatDays(days: number[]) {
  return days.map((d) => DAY_LABELS[d]).join(' ')
}

function formatCents(cents: number) {
  return `$${(cents / 100).toFixed(2)}`
}

onMounted(load)
</script>

<template>
  <div>
    <div class="d-flex align-center justify-space-between mb-6">
      <div class="text-h5 font-weight-bold">Ad Slots</div>
      <v-btn
        :prepend-icon="showForm ? 'mdi-close' : 'mdi-plus'"
        color="primary"
        variant="tonal"
        @click="showForm = !showForm"
      >
        {{ showForm ? 'Cancel' : 'Add Slot' }}
      </v-btn>
    </div>

    <v-expand-transition>
      <v-card v-if="showForm" class="mb-6" rounded="lg">
        <v-card-title class="text-subtitle-1 font-weight-semibold pa-4 pb-2">New Slot</v-card-title>
        <v-card-text>
          <v-form @submit.prevent="handleSubmit">
            <v-row>
              <v-col cols="12" md="6">
                <v-text-field v-model="form.device_id" label="Device ID" variant="outlined" density="comfortable" required />
              </v-col>
              <v-col cols="12" md="6">
                <v-text-field v-model="form.label" label="Label" variant="outlined" density="comfortable" placeholder="Morning Rush, Lunch Hour…" />
              </v-col>
            </v-row>

            <div class="text-body-2 font-weight-medium mb-2">Days of Week</div>
            <div class="d-flex flex-wrap gap-2 mb-4">
              <v-btn
                v-for="(dayLabel, idx) in DAY_LABELS"
                :key="idx"
                :color="form.days_of_week.includes(idx) ? 'primary' : 'default'"
                :variant="form.days_of_week.includes(idx) ? 'tonal' : 'outlined'"
                size="small"
                rounded="lg"
                @click="toggleDay(idx)"
              >
                {{ dayLabel }}
              </v-btn>
            </div>

            <v-row>
              <v-col cols="12" md="6">
                <v-text-field v-model="form.start_time" label="Start Time" type="time" variant="outlined" density="comfortable" required />
              </v-col>
              <v-col cols="12" md="6">
                <v-text-field v-model="form.end_time" label="End Time" type="time" variant="outlined" density="comfortable" required />
              </v-col>
              <v-col cols="12" md="6">
                <v-text-field v-model="form.duration_sec" label="Duration (seconds)" type="number" variant="outlined" density="comfortable" required />
              </v-col>
              <v-col cols="12" md="6">
                <v-text-field v-model="form.price_cents" label="Price (cents)" type="number" variant="outlined" density="comfortable" required />
              </v-col>
            </v-row>

            <v-alert v-if="formError" type="error" density="compact" class="mb-4" rounded="lg">{{ formError }}</v-alert>
            <div class="d-flex justify-end">
              <v-btn type="submit" color="primary" :loading="submitting">Save Slot</v-btn>
            </div>
          </v-form>
        </v-card-text>
      </v-card>
    </v-expand-transition>

    <v-alert v-if="error" type="error" class="mb-4" rounded="lg">{{ error }}</v-alert>

    <v-data-table
      :headers="headers"
      :items="slots"
      :loading="loading"
      rounded="lg"
      hover
    >
      <template #item.label="{ item }">{{ item.label ?? '—' }}</template>
      <template #item.device_id="{ item }">
        <span class="monospace text-medium-emphasis">{{ item.device_id.slice(0, 8) }}…</span>
      </template>
      <template #item.days_of_week="{ item }">{{ formatDays(item.days_of_week) }}</template>
      <template #item.window="{ item }">{{ item.start_time }} – {{ item.end_time }}</template>
      <template #item.duration_sec="{ item }">{{ item.duration_sec }}s</template>
      <template #item.price_cents="{ item }">{{ formatCents(item.price_cents) }}</template>
      <template #item.status="{ item }">
        <v-chip :color="slotStatusColor[item.status] ?? 'secondary'" size="small" variant="tonal">
          {{ item.status }}
        </v-chip>
      </template>
      <template #no-data>No slots yet.</template>
    </v-data-table>

    <v-snackbar v-model="snackbar" color="success" timeout="3000">Slot created.</v-snackbar>
  </div>
</template>
