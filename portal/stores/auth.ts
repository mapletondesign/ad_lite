import { defineStore } from 'pinia'
import type { AuthUser } from '~/types'

interface JwtPayload {
  sub: string
  role: 'admin' | 'advertiser' | 'venue'
  advertiser_id?: string
  venue_id?: string
  exp: number
}

function decodeJwtPayload(token: string): JwtPayload {
  const base64 = token.split('.')[1]
  const json = atob(base64.replace(/-/g, '+').replace(/_/g, '/'))
  return JSON.parse(json) as JwtPayload
}

const ACCESS_TOKEN_KEY = 'adlite_access_token'
const REFRESH_TOKEN_KEY = 'adlite_refresh_token'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    accessToken: null as string | null,
    refreshToken: null as string | null,
    user: null as AuthUser | null,
  }),

  getters: {
    isAuthenticated: (state): boolean => state.accessToken !== null && state.user !== null,
    role: (state): string | null => state.user?.role ?? null,
  },

  actions: {
    setTokens(accessToken: string, refreshToken: string) {
      this.accessToken = accessToken
      this.refreshToken = refreshToken
      localStorage.setItem(ACCESS_TOKEN_KEY, accessToken)
      localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken)

      const payload = decodeJwtPayload(accessToken)
      this.user = {
        id: payload.sub,
        role: payload.role,
        advertiser_id: payload.advertiser_id,
        venue_id: payload.venue_id,
      }
    },

    loadFromStorage() {
      if (typeof localStorage === 'undefined') return
      const accessToken = localStorage.getItem(ACCESS_TOKEN_KEY)
      const refreshToken = localStorage.getItem(REFRESH_TOKEN_KEY)
      if (accessToken && refreshToken) {
        try {
          this.setTokens(accessToken, refreshToken)
        } catch {
          this.logout()
        }
      }
    },

    logout() {
      this.accessToken = null
      this.refreshToken = null
      this.user = null
      if (typeof localStorage !== 'undefined') {
        localStorage.removeItem(ACCESS_TOKEN_KEY)
        localStorage.removeItem(REFRESH_TOKEN_KEY)
      }
    },
  },
})
