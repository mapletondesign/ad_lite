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

onMounted(loadSlots)
</script>

<template>
  <div class="page">
    <div class="page__header">
      <h1 class="page__title">Ad Slots</h1>
      <button class="btn btn--primary" @click="showForm = !showForm">
        {{ showForm ? 'Cancel' : 'Add Slot' }}
      </button>
    </div>

    <div v-if="showForm" class="inline-form">
      <h2 class="inline-form__title">New Slot</h2>
      <form @submit.prevent="handleSubmit">
        <div class="form-row">
          <div class="form-group">
            <label class="form-label" for="device_id">Device ID <span class="required">*</span></label>
            <input id="device_id" v-model="form.device_id" class="form-input" type="text" required />
          </div>
          <div class="form-group">
            <label class="form-label" for="label">Label</label>
            <input id="label" v-model="form.label" class="form-input" type="text" />
          </div>
        </div>
        <div class="form-group">
          <span class="form-label">Days of Week <span class="required">*</span></span>
          <div class="day-checkboxes">
            <label v-for="(dayLabel, idx) in DAY_LABELS" :key="idx" class="day-checkbox">
              <input
                type="checkbox"
                :checked="form.days_of_week.includes(idx)"
                @change="toggleDay(idx)"
              />
              {{ dayLabel }}
            </label>
          </div>
        </div>
        <div class="form-row">
          <div class="form-group">
            <label class="form-label" for="start_time">Start Time <span class="required">*</span></label>
            <input id="start_time" v-model="form.start_time" class="form-input" type="time" required />
          </div>
          <div class="form-group">
            <label class="form-label" for="end_time">End Time <span class="required">*</span></label>
            <input id="end_time" v-model="form.end_time" class="form-input" type="time" required />
          </div>
        </div>
        <div class="form-row">
          <div class="form-group">
            <label class="form-label" for="duration_sec">Duration (seconds) <span class="required">*</span></label>
            <input id="duration_sec" v-model="form.duration_sec" class="form-input" type="number" min="1" required />
          </div>
          <div class="form-group">
            <label class="form-label" for="price_cents">Price (cents) <span class="required">*</span></label>
            <input id="price_cents" v-model="form.price_cents" class="form-input" type="number" min="0" required />
          </div>
        </div>
        <p v-if="formError" class="form-error">{{ formError }}</p>
        <div class="form-actions">
          <button class="btn btn--primary" type="submit" :disabled="submitting">
            {{ submitting ? 'Saving…' : 'Save Slot' }}
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
            <td class="mono">{{ slot.device_id }}</td>
            <td>{{ formatDays(slot.days_of_week) }}</td>
            <td>{{ slot.start_time }} – {{ slot.end_time }}</td>
            <td>{{ slot.duration_sec }}s</td>
            <td>{{ formatCents(slot.price_cents) }}</td>
            <td>
              <span class="badge" :class="`badge--${slot.status}`">{{ slot.status }}</span>
            </td>
          </tr>
          <tr v-if="slots.length === 0">
            <td colspan="7" class="empty-cell">No slots yet.</td>
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

.day-checkboxes {
  display: flex;
  flex-wrap: wrap;
  gap: 0.75rem;
  margin-top: 0.25rem;
}

.day-checkbox {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.875rem;
  cursor: pointer;
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

.mono {
  font-family: monospace;
  font-size: 0.8125rem;
  color: var(--color-text-muted);
}

.badge {
  display: inline-block;
  padding: 0.2rem 0.6rem;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: capitalize;
}

.badge--available {
  background-color: #dcfce7;
  color: var(--color-success);
}

.badge--booked {
  background-color: #dbeafe;
  color: #1d4ed8;
}

.badge--paused {
  background-color: #fef9c3;
  color: var(--color-warning);
}
</style>
