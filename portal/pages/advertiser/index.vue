<script setup lang="ts">
import type { Booking } from '~/types'

definePageMeta({ middleware: 'role' })

const { apiFetch } = useApi()

const bookings = ref<Booking[]>([])
const loading = ref(true)
const errorMessage = ref('')

const activeCount    = computed(() => bookings.value.filter((b) => b.status === 'active').length)
const pendingCount   = computed(() => bookings.value.filter((b) => b.status === 'pending').length)
const totalSpentCents = computed(() =>
  bookings.value
    .filter((b) => b.status !== 'cancelled')
    .reduce((sum, b) => sum + b.price_cents, 0)
)

function formatCents(cents: number): string {
  return `$${(cents / 100).toFixed(2)}`
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString()
}

const recentBookings = computed(() => bookings.value.slice(0, 5))

const statusClass: Record<string, string> = {
  pending:   'text-bg-warning',
  active:    'text-bg-success',
  completed: 'text-bg-secondary',
  cancelled: 'text-bg-danger',
}

onMounted(async () => {
  try {
    bookings.value = await apiFetch<Booking[]>('/api/v1/advertiser/bookings')
  } catch {
    errorMessage.value = 'Failed to load dashboard data.'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div>
    <h1 class="h4 fw-bold mb-4">Advertiser Dashboard</h1>

    <div v-if="errorMessage" class="alert alert-danger">{{ errorMessage }}</div>

    <div class="row g-3 mb-4">
      <template v-if="loading">
        <div v-for="n in 4" :key="n" class="col-6 col-md-3">
          <div class="skeleton" />
        </div>
      </template>
      <template v-else>
        <div class="col-6 col-md-3">
          <div class="card h-100">
            <div class="card-body">
              <div class="fs-2 fw-bold">{{ bookings.length }}</div>
              <div class="text-muted small">Total Bookings</div>
            </div>
          </div>
        </div>
        <div class="col-6 col-md-3">
          <div class="card h-100">
            <div class="card-body">
              <div class="fs-2 fw-bold text-success">{{ activeCount }}</div>
              <div class="text-muted small">Active</div>
            </div>
          </div>
        </div>
        <div class="col-6 col-md-3">
          <div class="card h-100">
            <div class="card-body">
              <div class="fs-2 fw-bold text-warning">{{ pendingCount }}</div>
              <div class="text-muted small">Pending</div>
            </div>
          </div>
        </div>
        <div class="col-6 col-md-3">
          <div class="card h-100">
            <div class="card-body">
              <div class="fs-2 fw-bold">{{ formatCents(totalSpentCents) }}</div>
              <div class="text-muted small">Total Spend</div>
            </div>
          </div>
        </div>
      </template>
    </div>

    <div v-if="!loading && recentBookings.length > 0">
      <div class="d-flex align-items-center justify-content-between mb-2">
        <h2 class="h6 fw-semibold mb-0">Recent Bookings</h2>
        <NuxtLink to="/advertiser/bookings" class="small">View all</NuxtLink>
      </div>
      <div class="card">
        <table class="table table-dark table-hover mb-0">
          <thead>
            <tr>
              <th>Slot ID</th>
              <th>Starts On</th>
              <th>Ends On</th>
              <th>Price</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="booking in recentBookings" :key="booking.id">
              <td class="font-monospace small text-muted">{{ booking.slot_id }}</td>
              <td>{{ formatDate(booking.starts_on) }}</td>
              <td>{{ formatDate(booking.ends_on) }}</td>
              <td>{{ formatCents(booking.price_cents) }}</td>
              <td>
                <span class="badge" :class="statusClass[booking.status] ?? 'text-bg-secondary'">
                  {{ booking.status }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
