<script setup lang="ts">
import type { Booking } from '~/types'

definePageMeta({ middleware: 'role' })

const { apiFetch } = useApi()

const allBookings = ref<Booking[]>([])
const loading = ref(true)
const errorMessage = ref('')
const statusFilter = ref<string>('all')

const statusOptions = ['all', 'pending', 'active', 'completed', 'cancelled'] as const

const filteredBookings = computed(() => {
  if (statusFilter.value === 'all') return allBookings.value
  return allBookings.value.filter((b) => b.status === statusFilter.value)
})

async function loadBookings() {
  loading.value = true
  try {
    allBookings.value = await apiFetch<Booking[]>('/api/v1/bookings')
  } catch {
    errorMessage.value = 'Failed to load bookings.'
  } finally {
    loading.value = false
  }
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString()
}

function formatCents(cents: number): string {
  return `$${(cents / 100).toFixed(2)}`
}

const bookingStatusClass: Record<string, string> = {
  pending:   'text-bg-warning',
  active:    'text-bg-success',
  completed: 'text-bg-secondary',
  cancelled: 'text-bg-danger',
}

onMounted(loadBookings)
</script>

<template>
  <div>
    <div class="d-flex align-items-center justify-content-between mb-4">
      <h1 class="h4 fw-bold mb-0">Bookings</h1>
      <div class="d-flex align-items-center gap-2">
        <label class="form-label text-muted small mb-0" for="status-filter">Status</label>
        <select id="status-filter" v-model="statusFilter" class="form-select form-select-sm w-auto">
          <option v-for="opt in statusOptions" :key="opt" :value="opt">
            {{ opt.charAt(0).toUpperCase() + opt.slice(1) }}
          </option>
        </select>
      </div>
    </div>

    <div v-if="errorMessage" class="alert alert-danger">{{ errorMessage }}</div>

    <div class="card" :class="{ 'opacity-50': loading }">
      <div v-if="loading" class="card-body text-muted">Loading…</div>
      <table v-else class="table table-dark table-hover mb-0">
        <thead>
          <tr>
            <th>Slot ID</th>
            <th>Advertiser ID</th>
            <th>Starts On</th>
            <th>Ends On</th>
            <th>Price</th>
            <th>Status</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="booking in filteredBookings" :key="booking.id">
            <td class="font-monospace small text-muted">{{ booking.slot_id }}</td>
            <td class="font-monospace small text-muted">{{ booking.advertiser_id }}</td>
            <td>{{ formatDate(booking.starts_on) }}</td>
            <td>{{ formatDate(booking.ends_on) }}</td>
            <td>{{ formatCents(booking.price_cents) }}</td>
            <td>
              <span class="badge" :class="bookingStatusClass[booking.status] ?? 'text-bg-secondary'">
                {{ booking.status }}
              </span>
            </td>
          </tr>
          <tr v-if="filteredBookings.length === 0">
            <td colspan="6" class="text-center text-muted">No bookings found.</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
