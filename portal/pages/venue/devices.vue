<script setup lang="ts">
import type { Device } from '~/types'

definePageMeta({ middleware: 'role' })

const { apiFetch } = useApi()

const devices = ref<Device[]>([])
const loading = ref(true)
const error   = ref('')

const headers = [
  { title: 'Name',      key: 'name' },
  { title: 'Status',    key: 'status' },
  { title: 'Last Seen', key: 'last_seen' },
  { title: 'IP Address', key: 'ip_address' },
  { title: 'Firmware',  key: 'firmware_version' },
]

function formatDate(iso: string | null) {
  if (!iso) return '—'
  return new Date(iso).toLocaleString()
}

onMounted(async () => {
  try {
    devices.value = await apiFetch<Device[]>('/api/v1/venue/devices')
  } catch {
    error.value = 'Failed to load devices.'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div>
    <div class="text-h5 font-weight-bold mb-6">My Devices</div>

    <v-alert v-if="error" type="error" class="mb-4" rounded="lg">{{ error }}</v-alert>

    <v-data-table
      :headers="headers"
      :items="devices"
      :loading="loading"
      rounded="lg"
      hover
    >
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
      <template #item.ip_address="{ item }">
        <span class="monospace text-medium-emphasis">{{ item.ip_address ?? '—' }}</span>
      </template>
      <template #item.firmware_version="{ item }">{{ item.firmware_version ?? '—' }}</template>
      <template #no-data>No devices registered for your venue yet.</template>
    </v-data-table>
  </div>
</template>
