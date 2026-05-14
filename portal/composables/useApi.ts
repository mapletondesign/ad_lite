import { useAuthStore } from '~/stores/auth'

interface JwtPayload {
  exp: number
}

function getTokenExpiry(token: string): number {
  const base64 = token.split('.')[1]
  const json = atob(base64.replace(/-/g, '+').replace(/_/g, '/'))
  const payload = JSON.parse(json) as JwtPayload
  return payload.exp * 1000
}

export function useApi() {
  const config = useRuntimeConfig()
  const auth = useAuthStore()

  async function apiFetch<T>(path: string, options: Parameters<typeof $fetch>[1] = {}): Promise<T> {
    if (auth.accessToken && getTokenExpiry(auth.accessToken) < Date.now()) {
      if (!auth.refreshToken) {
        auth.logout()
        await navigateTo('/login')
        throw new Error('Session expired')
      }

      try {
        const refreshed = await $fetch<{ access_token: string; refresh_token: string }>(
          '/api/v1/auth/refresh',
          {
            baseURL: config.public.apiBase,
            method: 'POST',
            body: { refresh_token: auth.refreshToken },
          }
        )
        auth.setTokens(refreshed.access_token, refreshed.refresh_token)
      } catch {
        auth.logout()
        await navigateTo('/login')
        throw new Error('Session expired')
      }
    }

    return $fetch<T>(path, {
      baseURL: config.public.apiBase,
      ...options,
      headers: {
        ...(options.headers as Record<string, string> | undefined),
        Authorization: `Bearer ${auth.accessToken}`,
      },
    })
  }

  return { apiFetch }
}
