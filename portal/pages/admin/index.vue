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
  <div class="page">
    <h1 class="page__title">Dashboard</h1>

    <p v-if="errorMessage" class="error-text">{{ errorMessage }}</p>

    <div class="stats-grid" :class="{ loading: loading }">
      <template v-if="loading">
        <div v-for="n in 4" :key="n" class="stat-card stat-card--skeleton" />
      </template>
      <template v-else>
        <div class="stat-card">
          <div class="stat-card__value">{{ venues.length }}</div>
          <div class="stat-card__label">Total Venues</div>
        </div>
        <div class="stat-card">
          <div class="stat-card__value">{{ advertisers.length }}</div>
          <div class="stat-card__label">Total Advertisers</div>
        </div>
        <div class="stat-card">
          <div class="stat-card__value">
            {{ devices.length }}
            <span class="stat-card__sub">{{ onlineCount }} online</span>
          </div>
          <div class="stat-card__label">Total Devices</div>
        </div>
        <div class="stat-card">
          <div class="stat-card__value">{{ bookings.length }}</div>
          <div class="stat-card__label">Total Bookings</div>
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped>
.page {
  max-width: 1100px;
}

.page__title {
  font-size: 1.5rem;
  font-weight: 700;
  margin-bottom: 1.5rem;
  color: var(--color-text);
}

.error-text {
  color: var(--color-danger);
  font-size: 0.875rem;
  margin-bottom: 1rem;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 1rem;
}

.stats-grid.loading {
  opacity: 0.5;
}

.stat-card {
  background-color: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 1.5rem;
}

.stat-card--skeleton {
  height: 100px;
  background-color: #e2e8f0;
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.stat-card__value {
  font-size: 2rem;
  font-weight: 700;
  color: var(--color-text);
  display: flex;
  align-items: baseline;
  gap: 0.5rem;
}

.stat-card__sub {
  font-size: 0.875rem;
  font-weight: 400;
  color: var(--color-success);
}

.stat-card__label {
  font-size: 0.875rem;
  color: var(--color-text-muted);
  margin-top: 0.25rem;
}
</style>
