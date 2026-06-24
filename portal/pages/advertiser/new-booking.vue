<script setup lang="ts">
import type { AdSlot } from '~/types'

definePageMeta({ middleware: 'role' })

const { apiFetch } = useApi()
const router = useRouter()

// Step 1 — slot browser
const slots      = ref<AdSlot[]>([])
const loading    = ref(true)
const error      = ref('')

// Step 2 — booking form
const selected   = ref<AdSlot | null>(null)
const submitting = ref(false)
const formError  = ref('')
const form = reactive({
  starts_on:    '',
  ends_on:      '',
  creative_url: '',
})

const DAY_LABELS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

const slotHeaders = [
  { title: 'Label',       key: 'label' },
  { title: 'Days',        key: 'days_of_week' },
  { title: 'Time Window', key: 'window' },
  { title: 'Duration',    key: 'duration_sec' },
  { title: 'Price / Day', key: 'price_cents' },
  { title: '',            key: 'action', sortable: false },
]

function formatDays(days: number[]) {
  return days.map((d) => DAY_LABELS[d]).join(' ')
}

function formatCents(cents: number) {
  return `$${(cents / 100).toFixed(2)}`
}

function selectSlot(slot: AdSlot) {
  selected.value = slot
  formError.value = ''
  Object.assign(form, { starts_on: '', ends_on: '', creative_url: '' })
}

function clearSelection() {
  selected.value = null
}

const totalCents = computed(() => {
  if (!selected.value || !form.starts_on || !form.ends_on) return null
  const start = new Date(form.starts_on)
  const end   = new Date(form.ends_on)
  if (isNaN(start.getTime()) || isNaN(end.getTime()) || end < start) return null
  const days = Math.round((end.getTime() - start.getTime()) / 86_400_000) + 1
  return days * selected.value.price_cents
})

async function handleSubmit() {
  if (!selected.value) return
  formError.value = ''
  submitting.value = true
  try {
    await apiFetch('/api/v1/advertiser/bookings', {
      method: 'POST',
      body: {
        slot_id:      selected.value.id,
        starts_on:    form.starts_on,
        ends_on:      form.ends_on,
        creative_url: form.creative_url || undefined,
      },
    })
    router.push('/advertiser/bookings')
  } catch (e: any) {
    formError.value = e?.data?.error ?? 'Failed to create booking. The slot may no longer be available.'
  } finally {
    submitting.value = false
  }
}

onMounted(async () => {
  try {
    slots.value = await apiFetch<AdSlot[]>('/api/v1/advertiser/slots?status=available')
  } catch {
    error.value = 'Failed to load available slots.'
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div>
    <div class="d-flex align-center mb-6 ga-3">
      <v-btn icon="mdi-arrow-left" variant="text" size="small" :to="'/advertiser/bookings'" />
      <div class="text-h5 font-weight-bold">New Booking</div>
    </div>

    <v-alert v-if="error" type="error" class="mb-4" rounded="lg">{{ error }}</v-alert>

    <!-- Step 1: pick a slot -->
    <template v-if="!selected">
      <div class="text-subtitle-1 font-weight-medium mb-3">Available Slots</div>

      <v-data-table
        :headers="slotHeaders"
        :items="slots"
        :loading="loading"
        rounded="lg"
        hover
      >
        <template #item.label="{ item }">{{ item.label ?? '—' }}</template>
        <template #item.days_of_week="{ item }">{{ formatDays(item.days_of_week) }}</template>
        <template #item.window="{ item }">{{ item.start_time }} – {{ item.end_time }}</template>
        <template #item.duration_sec="{ item }">{{ item.duration_sec }}s</template>
        <template #item.price_cents="{ item }">{{ formatCents(item.price_cents) }}</template>
        <template #item.action="{ item }">
          <v-btn color="primary" size="small" variant="tonal" @click="selectSlot(item)">
            Select
          </v-btn>
        </template>
        <template #no-data>No available slots at the moment.</template>
      </v-data-table>
    </template>

    <!-- Step 2: fill booking details -->
    <template v-else>
      <v-card rounded="lg" class="mb-4" border elevation="0">
        <v-card-title class="text-subtitle-2 font-weight-semibold pa-4 pb-2">Selected Slot</v-card-title>
        <v-card-text class="pt-0">
          <div class="d-flex align-center justify-space-between flex-wrap ga-2">
            <div>
              <div class="font-weight-medium">{{ selected.label ?? 'Unnamed slot' }}</div>
              <div class="text-caption text-medium-emphasis">
                {{ formatDays(selected.days_of_week) }} · {{ selected.start_time }} – {{ selected.end_time }} · {{ selected.duration_sec }}s · {{ formatCents(selected.price_cents) }}/day
              </div>
            </div>
            <v-btn size="small" variant="text" color="error" @click="clearSelection">Change slot</v-btn>
          </div>
        </v-card-text>
      </v-card>

      <v-card rounded="lg" elevation="0" border>
        <v-card-title class="text-subtitle-1 font-weight-semibold pa-4 pb-2">Booking Details</v-card-title>
        <v-card-text>
          <v-form @submit.prevent="handleSubmit">
            <v-row>
              <v-col cols="12" md="6">
                <v-text-field
                  v-model="form.starts_on"
                  label="Start Date"
                  type="date"
                  variant="outlined"
                  density="comfortable"
                  required
                />
              </v-col>
              <v-col cols="12" md="6">
                <v-text-field
                  v-model="form.ends_on"
                  label="End Date"
                  type="date"
                  variant="outlined"
                  density="comfortable"
                  required
                />
              </v-col>
              <v-col cols="12">
                <v-text-field
                  v-model="form.creative_url"
                  label="Creative URL (optional)"
                  variant="outlined"
                  density="comfortable"
                  hint="Paste the URL from Upload Creative, or leave blank to add later."
                  persistent-hint
                  placeholder="https://…"
                />
              </v-col>
            </v-row>

            <v-alert v-if="formError" type="error" density="compact" class="my-4" rounded="lg">
              {{ formError }}
            </v-alert>

            <div v-if="totalCents !== null" class="text-body-2 text-medium-emphasis mb-4">
              Estimated total: <strong class="text-high-emphasis">{{ formatCents(totalCents) }}</strong>
            </div>

            <div class="d-flex justify-end">
              <v-btn type="submit" color="primary" :loading="submitting">Confirm Booking</v-btn>
            </div>
          </v-form>
        </v-card-text>
      </v-card>
    </template>
  </div>
</template>
