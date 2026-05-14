<script setup lang="ts">
import type { Venue, Advertiser, Device, Booking } from '~/types'

definePageMeta({ middleware: 'role' })

const { apiFetch } = useApi()

const loading = ref(true)
const errorMessage = ref('')

const venues = ref<Venue[]>([])
const advertisers = ref<Advertiser[]>([])
const devices = ref<Device[]>([])
const bookings = ref<Booking[]>([])

const onlineCount = computed(() => devices.value.filter((d) => d.status === 'online').length)

onMounted(async () => {
  try {
    const [v, a, d, b] = await Promise.all([
      apiFetch<Venue[]>('/api/v1/venues'),
      apiFetch<Advertiser[]>('/api/v1/advertisers'),
      apiFetch<Device[]>('/api/v1/devices'),
      apiFetch<Booking[]>('/api/v1/bookings'),
    ])
    venues.value = v
    advertisers.value = a
    devices.value = d
    bookings.value = b
  } catch {
    errorMessage.value = 'Failed to load dashboard data.'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div>
    <h1 class="h4 fw-bold mb-4">Dashboard</h1>

    <div v-if="errorMessage" class="alert alert-danger">{{ errorMessage }}</div>

    <div class="row g-3">
      <template v-if="loading">
        <div v-for="n in 4" :key="n" class="col-6 col-md-3">
          <div class="skeleton" />
        </div>
      </template>
      <template v-else>
        <div class="col-6 col-md-3">
          <div class="card h-100">
            <div class="card-body">
              <div class="fs-2 fw-bold">{{ venues.length }}</div>
              <div class="text-muted small">Total Venues</div>
            </div>
          </div>
        </div>
        <div class="col-6 col-md-3">
          <div class="card h-100">
            <div class="card-body">
              <div class="fs-2 fw-bold">{{ advertisers.length }}</div>
              <div class="text-muted small">Total Advertisers</div>
            </div>
          </div>
        </div>
        <div class="col-6 col-md-3">
          <div class="card h-100">
            <div class="card-body">
              <div class="fs-2 fw-bold">
                {{ devices.length }}
                <span class="fs-6 fw-normal text-success">{{ onlineCount }} online</span>
              </div>
              <div class="text-muted small">Total Devices</div>
            </div>
          </div>
        </div>
        <div class="col-6 col-md-3">
          <div class="card h-100">
            <div class="card-body">
              <div class="fs-2 fw-bold">{{ bookings.length }}</div>
              <div class="text-muted small">Total Bookings</div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>
