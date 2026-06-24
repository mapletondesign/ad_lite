<script setup lang="ts">
import type { Booking } from '~/types'

definePageMeta({ middleware: 'role' })

const { apiFetch } = useApi()

const allBookings   = ref<Booking[]>([])
const loading       = ref(true)
const error         = ref('')
const statusFilter  = ref('all')
const cancelDialog  = ref(false)
const cancelTarget  = ref<Booking | null>(null)
const cancelling    = ref(false)
const snackbar      = ref(false)
const snackbarMsg   = ref('')

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
  { title: 'Slot ID',   key: 'slot_id' },
  { title: 'Starts On', key: 'starts_on' },
  { title: 'Ends On',   key: 'ends_on' },
  { title: 'Price',     key: 'price_cents' },
  { title: 'Status',    key: 'status' },
  { title: 'Creative',  key: 'creative_url' },
  { title: '',          key: 'actions', sortable: false },
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

async function load() {
  loading.value = true
  try {
    allBookings.value = await apiFetch<Booking[]>('/api/v1/advertiser/bookings')
  } catch {
    error.value = 'Failed to load bookings.'
  } finally {
    loading.value = false
  }
}

function promptCancel(booking: Booking) {
  cancelTarget.value = booking
  cancelDialog.value = true
}

async function confirmCancel() {
  if (!cancelTarget.value) return
  cancelling.value = true
  try {
    await apiFetch(`/api/v1/advertiser/bookings/${cancelTarget.value.id}`, {
      method: 'PATCH',
      body: { cancel: true },
    })
    cancelDialog.value = false
    cancelTarget.value = null
    snackbarMsg.value = 'Booking cancelled.'
    snackbar.value = true
    await load()
  } catch (e: any) {
    error.value = e?.data?.error ?? 'Failed to cancel booking.'
  } finally {
    cancelling.value = false
  }
}

onMounted(load)
</script>

<template>
  <div>
    <div class="d-flex align-center justify-space-between mb-6">
      <div class="text-h5 font-weight-bold">My Bookings</div>
      <div class="d-flex align-center ga-3">
        <v-btn
          prepend-icon="mdi-plus"
          color="primary"
          variant="tonal"
          to="/advertiser/new-booking"
        >
          New Booking
        </v-btn>
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
      <template #item.starts_on="{ item }">{{ formatDate(item.starts_on) }}</template>
      <template #item.ends_on="{ item }">{{ formatDate(item.ends_on) }}</template>
      <template #item.price_cents="{ item }">{{ formatCents(item.price_cents) }}</template>
      <template #item.status="{ item }">
        <v-chip :color="bookingStatusColor[item.status] ?? 'secondary'" size="small" variant="tonal">
          {{ item.status }}
        </v-chip>
      </template>
      <template #item.creative_url="{ item }">
        <a v-if="item.creative_url" :href="item.creative_url" target="_blank" class="text-primary text-caption">
          View
        </a>
        <span v-else class="text-medium-emphasis">—</span>
      </template>
      <template #item.actions="{ item }">
        <v-btn
          v-if="item.status === 'pending' || item.status === 'active'"
          size="small"
          variant="text"
          color="error"
          @click="promptCancel(item)"
        >
          Cancel
        </v-btn>
      </template>
      <template #no-data>No bookings found.</template>
    </v-data-table>

    <!-- Cancel confirmation dialog -->
    <v-dialog v-model="cancelDialog" max-width="420">
      <v-card rounded="lg">
        <v-card-title class="text-subtitle-1 font-weight-semibold pa-4 pb-2">Cancel Booking</v-card-title>
        <v-card-text>
          Are you sure you want to cancel this booking? This action cannot be undone.
        </v-card-text>
        <v-card-actions class="pa-4 pt-2">
          <v-spacer />
          <v-btn variant="text" @click="cancelDialog = false">Keep</v-btn>
          <v-btn color="error" variant="tonal" :loading="cancelling" @click="confirmCancel">
            Cancel Booking
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-snackbar v-model="snackbar" color="success" timeout="3000">{{ snackbarMsg }}</v-snackbar>
  </div>
</template>
