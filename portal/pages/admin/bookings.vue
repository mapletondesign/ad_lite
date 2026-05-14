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

onMounted(loadBookings)
</script>

<template>
  <div class="page">
    <div class="page__header">
      <h1 class="page__title">Bookings</h1>
      <div class="filter-group">
        <label class="filter-label" for="status-filter">Status</label>
        <select id="status-filter" v-model="statusFilter" class="filter-select">
          <option v-for="opt in statusOptions" :key="opt" :value="opt">
            {{ opt.charAt(0).toUpperCase() + opt.slice(1) }}
          </option>
        </select>
      </div>
    </div>

    <p v-if="errorMessage" class="error-text">{{ errorMessage }}</p>

    <div class="table-wrap" :class="{ loading: loading }">
      <p v-if="loading" class="loading-text">Loading…</p>
      <table v-else class="data-table">
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
            <td class="mono">{{ booking.slot_id }}</td>
            <td class="mono">{{ booking.advertiser_id }}</td>
            <td>{{ formatDate(booking.starts_on) }}</td>
            <td>{{ formatDate(booking.ends_on) }}</td>
            <td>{{ formatCents(booking.price_cents) }}</td>
            <td>
              <span class="badge" :class="`badge--${booking.status}`">{{ booking.status }}</span>
            </td>
          </tr>
          <tr v-if="filteredBookings.length === 0">
            <td colspan="6" class="empty-cell">No bookings found.</td>
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

.filter-group {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.filter-label {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--color-text-muted);
}

.filter-select {
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  font-size: 0.875rem;
  color: var(--color-text);
  background-color: var(--color-surface);
}

.filter-select:focus {
  outline: none;
  border-color: var(--color-accent);
}

.error-text {
  color: var(--color-danger);
  font-size: 0.875rem;
  margin-bottom: 1rem;
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

.badge--pending {
  background-color: #fef9c3;
  color: var(--color-warning);
}

.badge--active {
  background-color: #dcfce7;
  color: var(--color-success);
}

.badge--completed {
  background-color: #f1f5f9;
  color: var(--color-text-muted);
}

.badge--cancelled {
  background-color: #fee2e2;
  color: var(--color-danger);
}
</style>
