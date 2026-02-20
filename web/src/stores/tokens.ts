import { ref } from 'vue'
import { defineStore } from 'pinia'

import { isUnauthorizedError } from '@/services/http'
import {
  createTrustedTokenAPI,
  deleteTrustedTokenAPI,
  fetchTrustedTokensAPI,
} from '@/services/workers.api'
import { redirectToLogin } from '@/stores/auth-redirect'
import { formatDateTime } from '@/utils/datetime'
import type { TrustedTokenCreateResponse, TrustedTokenItem } from '@/types/workers'

function isAbortError(error: unknown): boolean {
  if (typeof DOMException !== 'undefined' && error instanceof DOMException) {
    return error.name === 'AbortError'
  }
  return error instanceof Error && error.name === 'AbortError'
}

export const useTokensStore = defineStore('tokens', () => {
  const loading = ref(false)
  const errorMessage = ref('')
  const trustedTokens = ref<TrustedTokenItem[]>([])
  const refreshedAt = ref<Date | null>(null)
  const creatingTrustedToken = ref(false)
  const deletingTrustedTokenID = ref('')

  let activeController: AbortController | null = null
  let requestSerial = 0

  async function handleUnauthorized(): Promise<void> {
    await redirectToLogin(() => {
      reset()
    })
  }

  function reset(): void {
    trustedTokens.value = []
    loading.value = false
    errorMessage.value = ''
    refreshedAt.value = null
    deletingTrustedTokenID.value = ''
    creatingTrustedToken.value = false
  }

  function trustedTokenDeleteButtonText(tokenID: string): string {
    if (deletingTrustedTokenID.value === tokenID) {
      return 'Deleting...'
    }
    return 'Delete'
  }

  async function loadTokens(): Promise<void> {
    const serial = ++requestSerial
    activeController?.abort()
    const controller = new AbortController()
    activeController = controller

    loading.value = true
    errorMessage.value = ''
    try {
      const payload = await fetchTrustedTokensAPI(controller.signal)
      if (controller.signal.aborted || serial !== requestSerial) {
        return
      }
      trustedTokens.value = payload.items ?? []
      refreshedAt.value = new Date()
    } catch (error) {
      if (isAbortError(error) || serial !== requestSerial) {
        return
      }
      if (isUnauthorizedError(error)) {
        await handleUnauthorized()
        return
      }
      errorMessage.value = error instanceof Error ? error.message : 'Failed to load trusted tokens.'
    } finally {
      if (serial === requestSerial) {
        loading.value = false
      }
      if (activeController === controller) {
        activeController = null
      }
    }
  }

  async function createTrustedToken(payload: {
    name: string
  }): Promise<TrustedTokenCreateResponse> {
    if (creatingTrustedToken.value) {
      throw new Error('Trusted token creation already in progress.')
    }

    const name = payload.name.trim()
    if (!name) {
      errorMessage.value = 'name is required'
      throw new Error('name is required')
    }

    creatingTrustedToken.value = true
    errorMessage.value = ''
    try {
      const created = await createTrustedTokenAPI({ name })
      const tokenValue = created.token.trim()
      if (!tokenValue) {
        throw new Error('API returned empty token value.')
      }
      await loadTokens()
      return {
        ...created,
        token: tokenValue,
      }
    } catch (error) {
      if (isUnauthorizedError(error)) {
        await handleUnauthorized()
      }
      errorMessage.value =
        error instanceof Error ? error.message : 'Failed to create trusted token.'
      throw error
    } finally {
      creatingTrustedToken.value = false
    }
  }

  function confirmDeleteTrustedToken(tokenID: string): boolean {
    if (typeof window === 'undefined' || typeof window.confirm !== 'function') {
      return true
    }
    return window.confirm(`Delete token ${tokenID}?`)
  }

  async function deleteTrustedToken(tokenID: string): Promise<void> {
    if (
      !tokenID ||
      deletingTrustedTokenID.value === tokenID ||
      !confirmDeleteTrustedToken(tokenID)
    ) {
      return
    }
    deletingTrustedTokenID.value = tokenID
    errorMessage.value = ''
    try {
      await deleteTrustedTokenAPI(tokenID)
      await loadTokens()
    } catch (error) {
      if (isUnauthorizedError(error)) {
        await handleUnauthorized()
        return
      }
      errorMessage.value =
        error instanceof Error ? error.message : 'Failed to delete trusted token.'
    } finally {
      if (deletingTrustedTokenID.value === tokenID) {
        deletingTrustedTokenID.value = ''
      }
    }
  }

  function teardown(): void {
    requestSerial += 1
    activeController?.abort()
    activeController = null
    loading.value = false
  }

  return {
    loading,
    errorMessage,
    trustedTokens,
    refreshedAt,
    creatingTrustedToken,
    deletingTrustedTokenID,
    trustedTokenDeleteButtonText,
    formatDateTime,
    loadTokens,
    createTrustedToken,
    deleteTrustedToken,
    teardown,
    reset,
  }
})
