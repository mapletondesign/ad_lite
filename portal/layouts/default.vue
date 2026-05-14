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
  <div class="app-layout">
    <aside class="sidebar">
      <div class="sidebar__brand">AdLite</div>
      <nav class="sidebar__nav">
        <NuxtLink
          v-for="link in navLinks"
          :key="link.to"
          :to="link.to"
          class="sidebar__link"
          active-class="sidebar__link--active"
          exact-active-class="sidebar__link--active"
        >
          {{ link.label }}
        </NuxtLink>
      </nav>
      <div class="sidebar__footer">
        <div class="sidebar__user">
          <span class="sidebar__user-role">{{ auth.user?.role }}</span>
          <span class="sidebar__user-id">{{ auth.user?.id?.slice(0, 8) }}</span>
        </div>
        <button class="sidebar__logout" @click="logout">Logout</button>
      </div>
    </aside>
    <main class="main-content">
      <slot />
    </main>
  </div>
</template>

<style scoped>
.app-layout {
  display: flex;
  min-height: 100vh;
}

.sidebar {
  width: var(--sidebar-width);
  flex-shrink: 0;
  background-color: var(--color-primary);
  color: #ffffff;
  display: flex;
  flex-direction: column;
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  overflow-y: auto;
}

.sidebar__brand {
  padding: 1.5rem 1.25rem 1rem;
  font-size: 1.25rem;
  font-weight: 700;
  letter-spacing: -0.01em;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  margin-bottom: 0.5rem;
}

.sidebar__nav {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 0.5rem 0;
}

.sidebar__link {
  display: block;
  padding: 0.625rem 1.25rem;
  font-size: 0.9375rem;
  color: rgba(255, 255, 255, 0.75);
  transition: background-color 0.15s, color 0.15s;
  border-radius: 0;
}

.sidebar__link:hover {
  background-color: rgba(255, 255, 255, 0.08);
  color: #ffffff;
}

.sidebar__link--active {
  background-color: var(--color-accent);
  color: #ffffff;
}

.sidebar__footer {
  padding: 1rem 1.25rem;
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.sidebar__user {
  display: flex;
  flex-direction: column;
  gap: 0.125rem;
}

.sidebar__user-role {
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--color-accent);
  background-color: rgba(79, 70, 229, 0.2);
  padding: 0.125rem 0.5rem;
  border-radius: 999px;
  align-self: flex-start;
}

.sidebar__user-id {
  font-size: 0.75rem;
  color: rgba(255, 255, 255, 0.5);
  font-family: monospace;
}

.sidebar__logout {
  width: 100%;
  padding: 0.5rem;
  background-color: transparent;
  border: 1px solid rgba(255, 255, 255, 0.2);
  border-radius: var(--radius);
  color: rgba(255, 255, 255, 0.75);
  font-size: 0.875rem;
  transition: background-color 0.15s, color 0.15s;
}

.sidebar__logout:hover {
  background-color: rgba(255, 255, 255, 0.08);
  color: #ffffff;
}

.main-content {
  flex: 1;
  margin-left: var(--sidebar-width);
  background-color: var(--color-bg);
  padding: 2rem;
  min-height: 100vh;
  overflow-y: auto;
}
</style>
