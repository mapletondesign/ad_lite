<script setup lang="ts">
import type { Device } from '~/types'

definePageMeta({ middleware: 'role' })

const { apiFetch } = useApi()

const devices = ref<Device[]>([])
const loading = ref(true)
const errorMessage = ref('')

const onlineCount  = computed(() => devices.value.filter((d) => d.status === 'online').length)
const offlineCount = computed(() => devices.value.filter((d) => d.status === 'offline').length)

function formatDate(iso: string | null): string {
  if (!iso) return '—'
  return new Date(iso).toLocaleString()
}

onMounted(async () => {
  try {
    devices.value = await apiFetch<Device[]>('/api/v1/venue/devices')
  } catch {
    errorMessage.value = 'Failed to load device data.'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div>
    <h1 class="h4 fw-bold mb-4">Venue Dashboard</h1>

    <div v-if="errorMessage" class="alert alert-danger">{{ errorMessage }}</div>

    <div class="row g-3 mb-4">
      <template v-if="loading">
        <div v-for="n in 3" :key="n" class="col-6 col-md-4">
          <div class="skeleton" />
        </div>
      </template>
      <template v-else>
        <div class="col-6 col-md-4">
          <div class="card h-100">
            <div class="card-body">
              <div class="fs-2 fw-bold">{{ devices.length }}</div>
              <div class="text-muted small">Total Devices</div>
            </div>
          </div>
        </div>
        <div class="col-6 col-md-4">
          <div class="card h-100">
            <div class="card-body">
              <div class="fs-2 fw-bold text-success">{{ onlineCount }}</div>
              <div class="text-muted small">Online</div>
            </div>
          </div>
        </div>
        <div class="col-6 col-md-4">
          <div class="card h-100">
            <div class="card-body">
              <div class="fs-2 fw-bold text-danger">{{ offlineCount }}</div>
              <div class="text-muted small">Offline</div>
            </div>
          </div>
        </div>
      </template>
    </div>

    <div v-if="!loading && devices.length > 0">
      <div class="d-flex align-items-center justify-content-between mb-2">
        <h2 class="h6 fw-semibold mb-0">Devices</h2>
        <NuxtLink to="/venue/devices" class="small">View all</NuxtLink>
      </div>
      <div class="card">
        <table class="table table-dark table-hover mb-0">
          <thead>
            <tr>
              <th>Name</th>
              <th>Status</th>
              <th>Last Seen</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="device in devices" :key="device.id">
              <td>{{ device.name }}</td>
              <td>
                <span
                  class="badge"
                  :class="device.status === 'online' ? 'text-bg-success' : 'text-bg-danger'"
                >
                  {{ device.status }}
                </span>
              </td>
              <td class="small text-muted">{{ formatDate(device.last_seen) }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div v-else-if="!loading && devices.length === 0" class="card">
      <div class="card-body text-muted">No devices registered for your venue yet.</div>
    </div>
  </div>
</template>
