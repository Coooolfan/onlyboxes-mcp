import { defineStore } from 'pinia'

import {
  isInvalidCredentialsError,
  loginAPI,
  logoutAPI,
  probeSessionAPI,
} from '@/services/auth.api'
import { isUnauthorizedError } from '@/services/http'
import type { AccountProfile, AuthState, ConsoleSessionPayload } from '@/types/auth'

interface AuthStoreState {
  authState: AuthState
  bootstrapped: boolean
  bootstrapPromise: Promise<void> | null
  currentAccount: AccountProfile | null
  isAdmin: boolean
  registrationEnabled: boolean
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthStoreState => ({
    authState: 'loading',
    bootstrapped: false,
    bootstrapPromise: null,
    currentAccount: null,
    isAdmin: false,
    registrationEnabled: false,
  }),
  getters: {
    isAuthenticated: (state) => state.authState === 'authenticated',
    homePath: (state) => (state.isAdmin ? '/workers' : '/tokens'),
  },
  actions: {
    applySession(payload: ConsoleSessionPayload): void {
      this.authState = 'authenticated'
      this.currentAccount = payload.account
      this.isAdmin = payload.account.is_admin
      this.registrationEnabled = payload.registration_enabled
      this.bootstrapped = true
    },

    clearSession(): void {
      this.currentAccount = null
      this.isAdmin = false
      this.registrationEnabled = false
      this.authState = 'unauthenticated'
      this.bootstrapped = true
    },

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
          const sessionPayload = await probeSessionAPI()
          this.applySession(sessionPayload)
        } catch (error) {
          this.clearSession()
          if (!isUnauthorizedError(error)) {
            throw error
          }
        } finally {
          this.bootstrapPromise = null
        }
      })()

      this.bootstrapPromise = task
      await task
    },

    async login(username: string, password: string): Promise<void> {
      try {
        const sessionPayload = await loginAPI(username, password)
        this.applySession(sessionPayload)
      } catch (error) {
        if (isInvalidCredentialsError(error)) {
          throw error
        }
        throw error instanceof Error ? error : new Error('Failed to sign in.')
      }
    },

    logoutLocal(): void {
      this.clearSession()
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
