<script setup lang="ts">
import { useAuthStore } from '~/stores/auth'

const auth = useAuthStore()

interface NavLink {
  label: string
  to: string
  icon: string
  exact?: boolean
}

const adminLinks: NavLink[] = [
  { label: 'Dashboard',   to: '/admin',              icon: 'mdi-view-dashboard', exact: true },
  { label: 'Venues',      to: '/admin/venues',       icon: 'mdi-store' },
  { label: 'Advertisers', to: '/admin/advertisers',  icon: 'mdi-account-group' },
  { label: 'Devices',     to: '/admin/devices',      icon: 'mdi-monitor' },
  { label: 'Slots',       to: '/admin/slots',        icon: 'mdi-clock-outline' },
  { label: 'Bookings',    to: '/admin/bookings',     icon: 'mdi-calendar-check' },
]

const advertiserLinks: NavLink[] = [
  { label: 'Dashboard',       to: '/advertiser',          icon: 'mdi-view-dashboard', exact: true },
  { label: 'Upload Creative', to: '/advertiser/creative', icon: 'mdi-upload' },
  { label: 'My Bookings',     to: '/advertiser/bookings', icon: 'mdi-calendar' },
]

const venueLinks: NavLink[] = [
  { label: 'Dashboard', to: '/venue',         icon: 'mdi-view-dashboard', exact: true },
  { label: 'My Devices', to: '/venue/devices', icon: 'mdi-monitor' },
]

const navLinks = computed<NavLink[]>(() => {
  const role = auth.user?.role
  if (role === 'admin')      return adminLinks
  if (role === 'advertiser') return advertiserLinks
  if (role === 'venue')      return venueLinks
  return []
})

function logout() {
  auth.logout()
  navigateTo('/login')
}
</script>

<template>
  <v-navigation-drawer permanent width="220">
    <v-list-item
      title="AdLite"
      subtitle="DOOH Platform"
      class="py-4"
    />
    <v-divider />
    <v-list density="compact" nav class="py-2">
      <v-list-item
        v-for="link in navLinks"
        :key="link.to"
        :to="link.to"
        :exact="link.exact ?? false"
        :prepend-icon="link.icon"
        :title="link.label"
        rounded="lg"
        active-color="primary"
      />
    </v-list>
    <template #append>
      <v-divider />
      <div class="pa-3 d-flex flex-column gap-2">
        <v-chip size="small" variant="tonal" color="secondary" class="align-self-start text-uppercase">
          {{ auth.user?.role }}
        </v-chip>
        <span class="text-caption text-medium-emphasis monospace">
          {{ auth.user?.id?.slice(0, 8) }}
        </span>
        <v-btn variant="outlined" size="small" block @click="logout">Logout</v-btn>
      </div>
    </template>
  </v-navigation-drawer>

  <v-main>
    <v-container fluid class="pa-6">
      <slot />
    </v-container>
  </v-main>
</template>
