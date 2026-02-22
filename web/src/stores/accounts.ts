import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

import { deleteAccountAPI, fetchAccountsAPI } from '@/services/auth.api'
import { isUnauthorizedError } from '@/services/http'
import { redirectToLogin } from '@/stores/auth-redirect'
import { formatDateTime } from '@/utils/datetime'
import type { AccountListItem } from '@/types/auth'

const accountPageSize = 20

function isAbortError(error: unknown): boolean {
  if (typeof DOMException !== 'undefined' && error instanceof DOMException) {
    return error.name === 'AbortError'
  }
  return error instanceof Error && error.name === 'AbortError'
}

export const useAccountsStore = defineStore('accounts', () => {
  const loading = ref(false)
  const errorMessage = ref('')
  const accounts = ref<AccountListItem[]>([])
  const total = ref(0)
  const page = ref(1)
  const refreshedAt = ref<Date | null>(null)
  const deletingAccountID = ref('')

  let activeController: AbortController | null = null
  let requestSerial = 0

  const totalPages = computed(() => Math.max(1, Math.ceil(total.value / accountPageSize)))
  const canPrev = computed(() => page.value > 1)
  const canNext = computed(() => page.value < totalPages.value)

  const footerText = computed(() => {
    const start = total.value === 0 ? 0 : (page.value - 1) * accountPageSize + 1
    const end = Math.min(page.value * accountPageSize, total.value)
    return `${start}-${end} / ${total.value}`
  })

  async function handleUnauthorized(): Promise<void> {
    await redirectToLogin(() => {
      reset()
    })
  }

  function reset(): void {
    loading.value = false
    errorMessage.value = ''
    accounts.value = []
    total.value = 0
    page.value = 1
    refreshedAt.value = null
    deletingAccountID.value = ''
  }

  async function loadAccounts(targetPage = page.value): Promise<void> {
    const nextPage = Math.max(1, Math.floor(targetPage))
    page.value = nextPage

    const serial = ++requestSerial
    activeController?.abort()
    const controller = new AbortController()
    activeController = controller

    loading.value = true
    errorMessage.value = ''

    try {
      const payload = await fetchAccountsAPI(page.value, accountPageSize, controller.signal)
      if (controller.signal.aborted || serial !== requestSerial) {
        return
      }

      accounts.value = payload.items ?? []
      total.value = payload.total ?? 0
      const serverPage = typeof payload.page === 'number' ? Math.floor(payload.page) : page.value
      const normalizedPage = Math.max(1, serverPage)
      page.value = Math.min(normalizedPage, totalPages.value)
      refreshedAt.value = new Date()
    } catch (error) {
      if (isAbortError(error) || serial !== requestSerial) {
        return
      }
      if (isUnauthorizedError(error)) {
        await handleUnauthorized()
        return
      }
      errorMessage.value = error instanceof Error ? error.message : 'Failed to load accounts.'
    } finally {
      if (serial === requestSerial) {
        loading.value = false
      }
      if (activeController === controller) {
        activeController = null
      }
    }
  }

  function setPage(targetPage: number): void {
    const nextPage = Math.max(1, Math.floor(targetPage))
    if (nextPage === page.value) {
      return
    }
    page.value = nextPage
    void loadAccounts(nextPage)
  }

  function previousPage(): void {
    if (!canPrev.value) {
      return
    }
    page.value -= 1
    void loadAccounts(page.value)
  }

  function nextPage(): void {
    if (!canNext.value) {
      return
    }
    page.value += 1
    void loadAccounts(page.value)
  }

  function deleteAccountButtonText(accountID: string): string {
    if (deletingAccountID.value === accountID) {
      return 'Deleting...'
    }
    return 'Delete'
  }

  function confirmDeleteAccount(accountID: string): boolean {
    if (typeof window === 'undefined' || typeof window.confirm !== 'function') {
      return true
    }
    return window.confirm(`Delete account ${accountID}?`)
  }

  async function deleteAccount(accountID: string): Promise<void> {
    if (!accountID || deletingAccountID.value === accountID || !confirmDeleteAccount(accountID)) {
      return
    }

    deletingAccountID.value = accountID
    errorMessage.value = ''

    try {
      await deleteAccountAPI(accountID)
      const targetPage = accounts.value.length === 1 && page.value > 1 ? page.value - 1 : page.value
      await loadAccounts(targetPage)
    } catch (error) {
      if (isUnauthorizedError(error)) {
        await handleUnauthorized()
        return
      }
      errorMessage.value = error instanceof Error ? error.message : 'Failed to delete account.'
    } finally {
      if (deletingAccountID.value === accountID) {
        deletingAccountID.value = ''
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
    accountPageSize,
    loading,
    errorMessage,
    accounts,
    total,
    page,
    totalPages,
    canPrev,
    canNext,
    footerText,
    refreshedAt,
    deletingAccountID,
    deleteAccountButtonText,
    formatDateTime,
    loadAccounts,
    setPage,
    previousPage,
    nextPage,
    deleteAccount,
    teardown,
    reset,
  }
})
