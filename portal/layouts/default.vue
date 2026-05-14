<script setup lang="ts">
import { useAuthStore } from '~/stores/auth'

const auth = useAuthStore()

interface NavLink {
  label: string
  to: string
}

const adminLinks: NavLink[] = [
  { label: 'Dashboard', to: '/admin' },
  { label: 'Venues', to: '/admin/venues' },
  { label: 'Advertisers', to: '/admin/advertisers' },
  { label: 'Devices', to: '/admin/devices' },
  { label: 'Slots', to: '/admin/slots' },
  { label: 'Bookings', to: '/admin/bookings' },
]

const advertiserLinks: NavLink[] = [
  { label: 'Dashboard', to: '/advertiser' },
  { label: 'Upload Creative', to: '/advertiser/creative' },
  { label: 'My Bookings', to: '/advertiser/bookings' },
]

const venueLinks: NavLink[] = [
  { label: 'Dashboard', to: '/venue' },
  { label: 'My Devices', to: '/venue/devices' },
]

const navLinks = computed<NavLink[]>(() => {
  const role = auth.user?.role
  if (role === 'admin') return adminLinks
  if (role === 'advertiser') return advertiserLinks
  if (role === 'venue') return venueLinks
  return []
})

function logout() {
  auth.logout()
  navigateTo('/login')
}
</script>

<template>
  <div class="d-flex">
    <aside class="sidebar d-flex flex-column border-end border-secondary-subtle">
      <div class="px-3 py-4 fw-bold fs-5 border-bottom border-secondary-subtle text-white">
        AdLite
      </div>
      <nav class="flex-grow-1 px-2 py-2 d-flex flex-column gap-1">
        <NuxtLink
          v-for="link in navLinks"
          :key="link.to"
          :to="link.to"
          class="nav-link px-3 py-2"
          active-class="active"
          exact-active-class="active"
        >
          {{ link.label }}
        </NuxtLink>
      </nav>
      <div class="px-3 py-3 border-top border-secondary-subtle d-flex flex-column gap-2">
        <div class="d-flex flex-column gap-1">
          <span class="badge text-bg-secondary align-self-start text-uppercase" style="letter-spacing:.05em">
            {{ auth.user?.role }}
          </span>
          <span class="text-muted small font-monospace">{{ auth.user?.id?.slice(0, 8) }}</span>
        </div>
        <button class="btn btn-outline-secondary btn-sm w-100" @click="logout">Logout</button>
      </div>
    </aside>
    <main class="main-content p-4">
      <slot />
    </main>
  </div>
</template>
