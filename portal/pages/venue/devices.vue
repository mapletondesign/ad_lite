<script setup lang="ts">
import type { Device } from '~/types'

definePageMeta({ middleware: 'role' })

const { apiFetch } = useApi()

const devices = ref<Device[]>([])
const loading = ref(true)
const errorMessage = ref('')

async function load() {
  loading.value = true
  try {
    devices.value = await apiFetch<Device[]>('/api/v1/venue/devices')
  } catch {
    errorMessage.value = 'Failed to load devices.'
  } finally {
    loading.value = false
  }
}

function formatDate(iso: string | null): string {
  if (!iso) return '—'
  return new Date(iso).toLocaleString()
}

onMounted(load)
</script>

<template>
  <div>
    <h1 class="h4 fw-bold mb-4">My Devices</h1>

    <div v-if="errorMessage" class="alert alert-danger">{{ errorMessage }}</div>

    <div class="card" :class="{ 'opacity-50': loading }">
      <div v-if="loading" class="card-body text-muted">Loading…</div>
      <table v-else class="table table-dark table-hover mb-0">
        <thead>
          <tr>
            <th>Name</th>
            <th>Status</th>
            <th>Last Seen</th>
            <th>IP Address</th>
            <th>Firmware</th>
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
            <td class="small">{{ formatDate(device.last_seen) }}</td>
            <td class="font-monospace small text-muted">{{ device.ip_address ?? '—' }}</td>
            <td class="small">{{ device.firmware_version ?? '—' }}</td>
          </tr>
          <tr v-if="devices.length === 0">
            <td colspan="5" class="text-center text-muted">No devices registered for your venue yet.</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
