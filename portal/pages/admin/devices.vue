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
  <div class="page">
    <div class="page__header">
      <h1 class="page__title">Devices</h1>
    </div>

    <p v-if="errorMessage" class="error-text">{{ errorMessage }}</p>

    <div class="table-wrap" :class="{ loading: loading }">
      <p v-if="loading" class="loading-text">Loading…</p>
      <table v-else class="data-table">
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
            <td class="mono">{{ device.venue_id }}</td>
            <td>
              <span class="badge" :class="device.status === 'online' ? 'badge--online' : 'badge--offline'">
                {{ device.status }}
              </span>
            </td>
            <td>{{ formatDate(device.last_seen) }}</td>
            <td>{{ device.firmware_version ?? '—' }}</td>
            <td>{{ formatDate(device.created_at) }}</td>
          </tr>
          <tr v-if="devices.length === 0">
            <td colspan="6" class="empty-cell">No devices registered yet.</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
.page {
  max-width: 1100px;
}

.page__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.5rem;
}

.page__title {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--color-text);
}

.error-text {
  color: var(--color-danger);
  font-size: 0.875rem;
  margin-bottom: 1rem;
}

.table-wrap {
  background-color: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  overflow: hidden;
}

.table-wrap.loading {
  opacity: 0.5;
}

.loading-text {
  padding: 1.5rem;
  color: var(--color-text-muted);
  font-size: 0.9375rem;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
}

.data-table th {
  background-color: var(--color-primary);
  color: #ffffff;
  text-align: left;
  padding: 0.75rem 1rem;
  font-size: 0.8125rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.data-table td {
  padding: 0.75rem 1rem;
  font-size: 0.9375rem;
  border-bottom: 1px solid var(--color-border);
  color: var(--color-text);
}

.data-table tbody tr:nth-child(even) td {
  background-color: #f8fafc;
}

.data-table tbody tr:last-child td {
  border-bottom: none;
}

.empty-cell {
  color: var(--color-text-muted);
  text-align: center;
}

.mono {
  font-family: monospace;
  font-size: 0.8125rem;
  color: var(--color-text-muted);
}

.badge {
  display: inline-block;
  padding: 0.2rem 0.6rem;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: capitalize;
}

.badge--online {
  background-color: #dcfce7;
  color: var(--color-success);
}

.badge--offline {
  background-color: #fee2e2;
  color: var(--color-danger);
}
</style>
