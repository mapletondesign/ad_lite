<script setup lang="ts">
import type { Booking } from '~/types'

definePageMeta({ middleware: 'role' })

const { apiFetch } = useApi()

const allBookings  = ref<Booking[]>([])
const loading      = ref(true)
const error        = ref('')
const statusFilter = ref('all')

const statusOptions = [
  { title: 'All',       value: 'all' },
  { title: 'Pending',   value: 'pending' },
  { title: 'Active',    value: 'active' },
  { title: 'Completed', value: 'completed' },
  { title: 'Cancelled', value: 'cancelled' },
]

const filteredBookings = computed(() =>
  statusFilter.value === 'all'
    ? allBookings.value
    : allBookings.value.filter((b) => b.status === statusFilter.value)
)

const headers = [
  { title: 'Slot ID',       key: 'slot_id' },
  { title: 'Advertiser ID', key: 'advertiser_id' },
  { title: 'Starts On',     key: 'starts_on' },
  { title: 'Ends On',       key: 'ends_on' },
  { title: 'Price',         key: 'price_cents' },
  { title: 'Status',        key: 'status' },
]

const bookingStatusColor: Record<string, string> = {
  pending:   'warning',
  active:    'success',
  completed: 'secondary',
  cancelled: 'error',
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString()
}

function formatCents(cents: number) {
  return `$${(cents / 100).toFixed(2)}`
}

onMounted(async () => {
  try {
    allBookings.value = await apiFetch<Booking[]>('/api/v1/bookings')
  } catch {
    error.value = 'Failed to load bookings.'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div>
    <div class="d-flex align-center justify-space-between mb-6">
      <div class="text-h5 font-weight-bold">Bookings</div>
      <v-select
        v-model="statusFilter"
        :items="statusOptions"
        item-title="title"
        item-value="value"
        label="Status"
        variant="outlined"
        density="compact"
        hide-details
        style="max-width: 160px;"
      />
    </div>

    <v-alert v-if="error" type="error" class="mb-4" rounded="lg">{{ error }}</v-alert>

    <v-data-table
      :headers="headers"
      :items="filteredBookings"
      :loading="loading"
      rounded="lg"
      hover
    >
      <template #item.slot_id="{ item }">
        <span class="monospace text-medium-emphasis">{{ item.slot_id.slice(0, 8) }}…</span>
      </template>
      <template #item.advertiser_id="{ item }">
        <span class="monospace text-medium-emphasis">{{ item.advertiser_id.slice(0, 8) }}…</span>
      </template>
      <template #item.starts_on="{ item }">{{ formatDate(item.starts_on) }}</template>
      <template #item.ends_on="{ item }">{{ formatDate(item.ends_on) }}</template>
      <template #item.price_cents="{ item }">{{ formatCents(item.price_cents) }}</template>
      <template #item.status="{ item }">
        <v-chip :color="bookingStatusColor[item.status] ?? 'secondary'" size="small" variant="tonal">
          {{ item.status }}
        </v-chip>
      </template>
      <template #no-data>No bookings found.</template>
    </v-data-table>
  </div>
</template>
