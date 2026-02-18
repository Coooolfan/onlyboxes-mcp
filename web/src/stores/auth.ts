import { defineStore } from 'pinia'

import { isInvalidCredentialsError, loginAPI, logoutAPI, probeSessionAPI } from '@/services/auth.api'
import { isUnauthorizedError } from '@/services/http'
import type { AuthState } from '@/types/auth'

interface AuthStoreState {
  authState: AuthState
  bootstrapped: boolean
  bootstrapPromise: Promise<void> | null
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthStoreState => ({
    authState: 'loading',
    bootstrapped: false,
    bootstrapPromise: null,
  }),
  getters: {
    isAuthenticated: (state) => state.authState === 'authenticated',
  },
  actions: {
    async bootstrap(): Promise<void> {
      if (this.bootstrapped) {
        return
      }
      if (this.bootstrapPromise) {
        await this.bootstrapPromise
        return
      }

      const task = (async () => {
        try {
          await probeSessionAPI()
          this.authState = 'authenticated'
        } catch (error) {
          this.authState = 'unauthenticated'
          if (!isUnauthorizedError(error)) {
            throw error
          }
        } finally {
          this.bootstrapped = true
          this.bootstrapPromise = null
        }
      })()

      this.bootstrapPromise = task
      await task
    },

    async login(username: string, password: string): Promise<void> {
      try {
        await loginAPI(username, password)
      } catch (error) {
        if (isInvalidCredentialsError(error)) {
          throw error
        }
        throw error instanceof Error ? error : new Error('Failed to sign in.')
      }

      this.authState = 'authenticated'
      this.bootstrapped = true
    },

    logoutLocal(): void {
      this.authState = 'unauthenticated'
      this.bootstrapped = true
    },

    async logout(): Promise<void> {
      try {
        await logoutAPI()
      } catch {
        // ignore network failures and force local logout state
      }
      this.logoutLocal()
    },
  },
})
