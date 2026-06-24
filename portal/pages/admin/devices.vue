<script setup lang="ts">
import type { Device } from '~/types'

definePageMeta({ middleware: 'role' })

const { apiFetch } = useApi()

const devices = ref<Device[]>([])
const loading = ref(true)
const error   = ref('')

const headers = [
  { title: 'Name',      key: 'name' },
  { title: 'Venue ID',  key: 'venue_id' },
  { title: 'Status',    key: 'status' },
  { title: 'Last Seen', key: 'last_seen' },
  { title: 'Firmware',  key: 'firmware_version' },
  { title: 'Created',   key: 'created_at' },
]

function formatDate(iso: string | null) {
  if (!iso) return '—'
  return new Date(iso).toLocaleDateString()
}

onMounted(async () => {
  try {
    devices.value = await apiFetch<Device[]>('/api/v1/devices')
  } catch {
    error.value = 'Failed to load devices.'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div>
    <div class="text-h5 font-weight-bold mb-6">Devices</div>

    <v-alert v-if="error" type="error" class="mb-4" rounded="lg">{{ error }}</v-alert>

    <v-data-table
      :headers="headers"
      :items="devices"
      :loading="loading"
      rounded="lg"
      hover
    >
      <template #item.venue_id="{ item }">
        <span class="monospace text-medium-emphasis">{{ item.venue_id.slice(0, 8) }}…</span>
      </template>
      <template #item.status="{ item }">
        <v-chip
          :color="item.status === 'online' ? 'success' : 'error'"
          size="small"
          variant="tonal"
        >
          {{ item.status }}
        </v-chip>
      </template>
      <template #item.last_seen="{ item }">{{ formatDate(item.last_seen) }}</template>
      <template #item.firmware_version="{ item }">{{ item.firmware_version ?? '—' }}</template>
      <template #item.created_at="{ item }">{{ formatDate(item.created_at) }}</template>
      <template #no-data>No devices registered yet.</template>
    </v-data-table>
  </div>
</template>
