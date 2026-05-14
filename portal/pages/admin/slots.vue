<script setup lang="ts">
import type { AdSlot } from '~/types'

definePageMeta({ middleware: 'role' })

const { apiFetch } = useApi()

const slots = ref<AdSlot[]>([])
const loading = ref(true)
const errorMessage = ref('')
const showForm = ref(false)
const submitting = ref(false)
const formError = ref('')

const DAY_LABELS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

const form = reactive({
  device_id: '',
  label: '',
  days_of_week: [] as number[],
  start_time: '',
  end_time: '',
  duration_sec: 30,
  price_cents: 0,
})

async function loadSlots() {
  loading.value = true
  try {
    slots.value = await apiFetch<AdSlot[]>('/api/v1/slots')
  } catch {
    errorMessage.value = 'Failed to load slots.'
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
        device_id: form.device_id,
        label: form.label || null,
        days_of_week: form.days_of_week,
        start_time: form.start_time,
        end_time: form.end_time,
        duration_sec: Number(form.duration_sec),
        price_cents: Number(form.price_cents),
      },
    })
    form.device_id = ''
    form.label = ''
    form.days_of_week = []
    form.start_time = ''
    form.end_time = ''
    form.duration_sec = 30
    form.price_cents = 0
    showForm.value = false
    await loadSlots()
  } catch {
    formError.value = 'Failed to create slot.'
  } finally {
    submitting.value = false
  }
}

function formatDays(days: number[]): string {
  return days.map((d) => DAY_LABELS[d]).join(' ')
}

function formatCents(cents: number): string {
  return `$${(cents / 100).toFixed(2)}`
}

function toggleDay(day: number) {
  const idx = form.days_of_week.indexOf(day)
  if (idx === -1) {
    form.days_of_week.push(day)
  } else {
    form.days_of_week.splice(idx, 1)
  }
}

const slotStatusClass: Record<string, string> = {
  available: 'text-bg-success',
  booked: 'text-bg-primary',
  paused: 'text-bg-warning',
}

onMounted(loadSlots)
</script>

<template>
  <div>
    <div class="d-flex align-items-center justify-content-between mb-4">
      <h1 class="h4 fw-bold mb-0">Ad Slots</h1>
      <button class="btn btn-primary btn-sm" @click="showForm = !showForm">
        {{ showForm ? 'Cancel' : 'Add Slot' }}
      </button>
    </div>

    <div v-if="showForm" class="card mb-4">
      <div class="card-header fw-semibold">New Slot</div>
      <div class="card-body">
        <form @submit.prevent="handleSubmit">
          <div class="row g-3 mb-3">
            <div class="col-md-6">
              <label class="form-label" for="device_id">Device ID <span class="text-danger">*</span></label>
              <input id="device_id" v-model="form.device_id" class="form-control font-monospace" type="text" required />
            </div>
            <div class="col-md-6">
              <label class="form-label" for="label">Label</label>
              <input id="label" v-model="form.label" class="form-control" type="text" placeholder="Morning Rush, Lunch Hour…" />
            </div>
          </div>
          <div class="mb-3">
            <div class="form-label">Days of Week <span class="text-danger">*</span></div>
            <div class="d-flex flex-wrap gap-3">
              <div v-for="(dayLabel, idx) in DAY_LABELS" :key="idx" class="form-check">
                <input
                  :id="`day-${idx}`"
                  class="form-check-input"
                  type="checkbox"
                  :checked="form.days_of_week.includes(idx)"
                  @change="toggleDay(idx)"
                />
                <label class="form-check-label" :for="`day-${idx}`">{{ dayLabel }}</label>
              </div>
            </div>
          </div>
          <div class="row g-3 mb-3">
            <div class="col-md-6">
              <label class="form-label" for="start_time">Start Time <span class="text-danger">*</span></label>
              <input id="start_time" v-model="form.start_time" class="form-control" type="time" required />
            </div>
            <div class="col-md-6">
              <label class="form-label" for="end_time">End Time <span class="text-danger">*</span></label>
              <input id="end_time" v-model="form.end_time" class="form-control" type="time" required />
            </div>
          </div>
          <div class="row g-3 mb-3">
            <div class="col-md-6">
              <label class="form-label" for="duration_sec">Duration (seconds) <span class="text-danger">*</span></label>
              <input id="duration_sec" v-model="form.duration_sec" class="form-control" type="number" min="1" required />
            </div>
            <div class="col-md-6">
              <label class="form-label" for="price_cents">Price (cents) <span class="text-danger">*</span></label>
              <input id="price_cents" v-model="form.price_cents" class="form-control" type="number" min="0" required />
            </div>
          </div>
          <div v-if="formError" class="alert alert-danger py-2 small">{{ formError }}</div>
          <div class="d-flex justify-content-end">
            <button class="btn btn-primary" type="submit" :disabled="submitting">
              {{ submitting ? 'Saving…' : 'Save Slot' }}
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
            <th>Label</th>
            <th>Device ID</th>
            <th>Days</th>
            <th>Time Window</th>
            <th>Duration</th>
            <th>Price</th>
            <th>Status</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="slot in slots" :key="slot.id">
            <td>{{ slot.label ?? '—' }}</td>
            <td class="font-monospace small text-muted">{{ slot.device_id }}</td>
            <td>{{ formatDays(slot.days_of_week) }}</td>
            <td>{{ slot.start_time }} – {{ slot.end_time }}</td>
            <td>{{ slot.duration_sec }}s</td>
            <td>{{ formatCents(slot.price_cents) }}</td>
            <td>
              <span class="badge" :class="slotStatusClass[slot.status] ?? 'text-bg-secondary'">
                {{ slot.status }}
              </span>
            </td>
          </tr>
          <tr v-if="slots.length === 0">
            <td colspan="7" class="text-center text-muted">No slots yet.</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
