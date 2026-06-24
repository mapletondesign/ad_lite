<script setup lang="ts">
import type { Venue, Advertiser, Device, Booking } from '~/types'

definePageMeta({ middleware: 'role' })

const { apiFetch } = useApi()

const loading     = ref(true)
const error       = ref('')
const venues      = ref<Venue[]>([])
const advertisers = ref<Advertiser[]>([])
const devices     = ref<Device[]>([])
const bookings    = ref<Booking[]>([])

const onlineCount = computed(() => devices.value.filter((d) => d.status === 'online').length)

const stats = computed(() => [
  { label: 'Total Venues',      value: venues.value.length,      icon: 'mdi-store',          color: 'primary' },
  { label: 'Total Advertisers', value: advertisers.value.length, icon: 'mdi-account-group',  color: 'secondary' },
  { label: 'Total Devices',     value: devices.value.length,     icon: 'mdi-monitor',         color: 'info',    sub: `${onlineCount.value} online` },
  { label: 'Total Bookings',    value: bookings.value.length,    icon: 'mdi-calendar-check',  color: 'success' },
])

onMounted(async () => {
  try {
    const [v, a, d, b] = await Promise.all([
      apiFetch<Venue[]>('/api/v1/venues'),
      apiFetch<Advertiser[]>('/api/v1/advertisers'),
      apiFetch<Device[]>('/api/v1/devices'),
      apiFetch<Booking[]>('/api/v1/bookings'),
    ])
    venues.value      = v
    advertisers.value = a
    devices.value     = d
    bookings.value    = b
  } catch {
    error.value = 'Failed to load dashboard data.'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div>
    <div class="text-h5 font-weight-bold mb-6">Dashboard</div>

    <v-alert v-if="error" type="error" class="mb-6" rounded="lg">{{ error }}</v-alert>

    <v-row>
      <v-col v-for="stat in stats" :key="stat.label" cols="6" md="3">
        <v-skeleton-loader v-if="loading" type="card" />
        <v-card v-else rounded="lg" height="100%">
          <v-card-text>
            <div class="d-flex align-center justify-space-between mb-2">
              <v-icon :color="stat.color" size="28">{{ stat.icon }}</v-icon>
            </div>
            <div class="text-h4 font-weight-bold">{{ stat.value }}</div>
            <div v-if="stat.sub" class="text-caption text-success">{{ stat.sub }}</div>
            <div class="text-body-2 text-medium-emphasis mt-1">{{ stat.label }}</div>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>
