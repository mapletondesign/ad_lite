<script setup lang="ts">
import type { Booking } from '~/types'

definePageMeta({ middleware: 'role' })

const { apiFetch } = useApi()

const bookings = ref<Booking[]>([])
const loading  = ref(true)
const error    = ref('')

const activeCount     = computed(() => bookings.value.filter((b) => b.status === 'active').length)
const pendingCount    = computed(() => bookings.value.filter((b) => b.status === 'pending').length)
const totalSpentCents = computed(() =>
  bookings.value
    .filter((b) => b.status !== 'cancelled')
    .reduce((sum, b) => sum + b.price_cents, 0)
)

const recentBookings = computed(() => bookings.value.slice(0, 5))

const recentHeaders = [
  { title: 'Slot ID',    key: 'slot_id' },
  { title: 'Starts On', key: 'starts_on' },
  { title: 'Ends On',   key: 'ends_on' },
  { title: 'Price',     key: 'price_cents' },
  { title: 'Status',    key: 'status' },
]

const bookingStatusColor: Record<string, string> = {
  pending:   'warning',
  active:    'success',
  completed: 'secondary',
  cancelled: 'error',
}

function formatCents(cents: number) {
  return `$${(cents / 100).toFixed(2)}`
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleDateString()
}

const stats = computed(() => [
  { label: 'Total Bookings', value: bookings.value.length, icon: 'mdi-calendar-multiple',   color: 'primary' },
  { label: 'Active',         value: activeCount.value,     icon: 'mdi-play-circle-outline',  color: 'success' },
  { label: 'Pending',        value: pendingCount.value,    icon: 'mdi-clock-outline',         color: 'warning' },
  { label: 'Total Spend',    value: formatCents(totalSpentCents.value), icon: 'mdi-currency-usd', color: 'info' },
])

onMounted(async () => {
  try {
    bookings.value = await apiFetch<Booking[]>('/api/v1/advertiser/bookings')
  } catch {
    error.value = 'Failed to load dashboard data.'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div>
    <div class="text-h5 font-weight-bold mb-6">Advertiser Dashboard</div>

    <v-alert v-if="error" type="error" class="mb-4" rounded="lg">{{ error }}</v-alert>

    <v-row class="mb-6">
      <v-col v-if="loading" v-for="n in 4" :key="n" cols="6" md="3">
        <v-skeleton-loader type="card" rounded="lg" />
      </v-col>
      <template v-else>
        <v-col v-for="stat in stats" :key="stat.label" cols="6" md="3">
          <v-card rounded="lg" elevation="0" border>
            <v-card-text class="d-flex align-center ga-3">
              <v-icon :color="stat.color" size="36">{{ stat.icon }}</v-icon>
              <div>
                <div class="text-h5 font-weight-bold">{{ stat.value }}</div>
                <div class="text-caption text-medium-emphasis">{{ stat.label }}</div>
              </div>
            </v-card-text>
          </v-card>
        </v-col>
      </template>
    </v-row>

    <template v-if="!loading && recentBookings.length > 0">
      <div class="d-flex align-center justify-space-between mb-3">
        <div class="text-subtitle-1 font-weight-medium">Recent Bookings</div>
        <NuxtLink to="/advertiser/bookings" class="text-caption text-primary">View all</NuxtLink>
      </div>

      <v-data-table
        :headers="recentHeaders"
        :items="recentBookings"
        rounded="lg"
        hover
        hide-default-footer
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
      </v-data-table>
    </template>
  </div>
</template>
