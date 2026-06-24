<script setup lang="ts">
import type { Device } from '~/types'

definePageMeta({ middleware: 'role' })

const { apiFetch } = useApi()

const devices = ref<Device[]>([])
const loading = ref(true)
const error   = ref('')

const onlineCount  = computed(() => devices.value.filter((d) => d.status === 'online').length)
const offlineCount = computed(() => devices.value.filter((d) => d.status === 'offline').length)

const recentHeaders = [
  { title: 'Name',      key: 'name' },
  { title: 'Status',    key: 'status' },
  { title: 'Last Seen', key: 'last_seen' },
]

const stats = computed(() => [
  { label: 'Total Devices', value: devices.value.length, icon: 'mdi-monitor-multiple', color: 'primary' },
  { label: 'Online',        value: onlineCount.value,    icon: 'mdi-wifi',              color: 'success' },
  { label: 'Offline',       value: offlineCount.value,   icon: 'mdi-wifi-off',          color: 'error' },
])

function formatDate(iso: string | null) {
  if (!iso) return '—'
  return new Date(iso).toLocaleString()
}

onMounted(async () => {
  try {
    devices.value = await apiFetch<Device[]>('/api/v1/venue/devices')
  } catch {
    error.value = 'Failed to load device data.'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div>
    <div class="text-h5 font-weight-bold mb-6">Venue Dashboard</div>

    <v-alert v-if="error" type="error" class="mb-4" rounded="lg">{{ error }}</v-alert>

    <v-row class="mb-6">
      <v-col v-if="loading" v-for="n in 3" :key="n" cols="6" md="4">
        <v-skeleton-loader type="card" rounded="lg" />
      </v-col>
      <template v-else>
        <v-col v-for="stat in stats" :key="stat.label" cols="6" md="4">
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

    <template v-if="!loading">
      <div class="d-flex align-center justify-space-between mb-3">
        <div class="text-subtitle-1 font-weight-medium">Devices</div>
        <NuxtLink to="/venue/devices" class="text-caption text-primary">View all</NuxtLink>
      </div>

      <v-data-table
        :headers="recentHeaders"
        :items="devices"
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
        <template #no-data>No devices registered for your venue yet.</template>
      </v-data-table>
    </template>
  </div>
</template>
