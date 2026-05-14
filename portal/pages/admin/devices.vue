<script setup lang="ts">
import type { Device } from '~/types'

definePageMeta({ middleware: 'role' })

const { apiFetch } = useApi()

const devices = ref<Device[]>([])
const loading = ref(true)
const errorMessage = ref('')

async function loadDevices() {
  loading.value = true
  try {
    devices.value = await apiFetch<Device[]>('/api/v1/devices')
  } catch {
    errorMessage.value = 'Failed to load devices.'
  } finally {
    loading.value = false
  }
}

function formatDate(iso: string | null): string {
  if (!iso) return '—'
  return new Date(iso).toLocaleDateString()
}

onMounted(loadDevices)
</script>

<template>
  <div>
    <h1 class="h4 fw-bold mb-4">Devices</h1>

    <div v-if="errorMessage" class="alert alert-danger">{{ errorMessage }}</div>

    <div class="card" :class="{ 'opacity-50': loading }">
      <div v-if="loading" class="card-body text-muted">Loading…</div>
      <table v-else class="table table-dark table-hover mb-0">
        <thead>
          <tr>
            <th>Name</th>
            <th>Venue ID</th>
            <th>Status</th>
            <th>Last Seen</th>
            <th>Firmware</th>
            <th>Created</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="device in devices" :key="device.id">
            <td>{{ device.name }}</td>
            <td class="font-monospace small text-muted">{{ device.venue_id }}</td>
            <td>
              <span
                class="badge"
                :class="device.status === 'online' ? 'text-bg-success' : 'text-bg-danger'"
              >
                {{ device.status }}
              </span>
            </td>
            <td>{{ formatDate(device.last_seen) }}</td>
            <td>{{ device.firmware_version ?? '—' }}</td>
            <td>{{ formatDate(device.created_at) }}</td>
          </tr>
          <tr v-if="devices.length === 0">
            <td colspan="6" class="text-center text-muted">No devices registered yet.</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
