<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'

import ErrorBanner from '@/components/common/ErrorBanner.vue'
import AccountsPanel from '@/components/dashboard/AccountsPanel.vue'
import ChangePasswordModal from '@/components/dashboard/ChangePasswordModal.vue'
import ConsoleRouteTabs from '@/components/dashboard/ConsoleRouteTabs.vue'
import CreateAccountModal from '@/components/dashboard/CreateAccountModal.vue'
import { useAccountsStore } from '@/stores/accounts'
import { useAuthStore } from '@/stores/auth'
import { useTokensStore } from '@/stores/tokens'
import { useWorkersStore } from '@/stores/workers'

const authStore = useAuthStore()
const accountsStore = useAccountsStore()
const tokensStore = useTokensStore()
const workersStore = useWorkersStore()
const router = useRouter()

const showChangePasswordModal = ref(false)
const showCreateAccountModal = ref(false)

const refreshedAtText = computed(() => {
  if (!accountsStore.refreshedAt) {
    return 'never'
  }
  return accountsStore.formatDateTime(accountsStore.refreshedAt.toISOString())
})

const showCreateAccountPanel = computed(() => authStore.isAdmin && authStore.registrationEnabled)

async function handleLogout(): Promise<void> {
  await authStore.logout()
  accountsStore.teardown()
  accountsStore.reset()
  workersStore.teardown()
  workersStore.reset()
  tokensStore.teardown()
  tokensStore.reset()
  await router.replace('/login')
}

async function handleRefresh(): Promise<void> {
  await accountsStore.loadAccounts(accountsStore.page)
}

function openChangePasswordModal(): void {
  showChangePasswordModal.value = true
}

function closeChangePasswordModal(): void {
  showChangePasswordModal.value = false
}

function openCreateAccountModal(): void {
  showCreateAccountModal.value = true
}

function closeCreateAccountModal(): void {
  showCreateAccountModal.value = false
}

onMounted(async () => {
  await accountsStore.loadAccounts(accountsStore.page)
})

onBeforeUnmount(() => {
  accountsStore.teardown()
})
</script>

<template>
  <main class="relative z-2 mx-auto w-[min(1240px,100%)] grid gap-6">
    <ConsoleRouteTabs />

    <header
      class="flex items-start justify-between gap-5 bg-surface border border-stroke rounded-lg p-8 shadow-card animate-rise-in max-[960px]:flex-col"
    >
      <div>
        <p class="m-0 font-mono text-xs tracking-[0.05em] uppercase text-secondary">
          Onlyboxes / Account Console
        </p>
        <h1 class="mt-3 mb-2 text-2xl font-semibold leading-[1.2] tracking-[-0.02em]">
          Account Administration
        </h1>
        <p class="m-0 text-secondary text-sm leading-normal">
          Account: <strong>{{ authStore.currentAccount?.username ?? '--' }}</strong>
        </p>
      </div>

      <div class="flex items-center gap-3 max-[960px]:w-full max-[960px]:flex-wrap">
        <button
          v-if="showCreateAccountPanel"
          class="ghost-btn rounded-md px-3.5 py-2 text-sm font-medium h-9 inline-flex items-center justify-center text-primary bg-surface border border-stroke transition-all duration-200 hover:not-disabled:border-stroke-hover hover:not-disabled:bg-surface-soft disabled:cursor-not-allowed disabled:opacity-50"
          type="button"
          @click="openCreateAccountModal"
        >
          Create Account
        </button>
        <button
          class="rounded-md px-3.5 py-2 text-sm font-medium h-9 inline-flex items-center justify-center text-white bg-accent border border-accent transition-all duration-200 hover:not-disabled:bg-[#333] hover:not-disabled:border-[#333] disabled:cursor-not-allowed disabled:opacity-50"
          type="button"
          :disabled="accountsStore.loading"
          @click="handleRefresh"
        >
          {{ accountsStore.loading ? 'Refreshing...' : 'Refresh' }}
        </button>
        <button
          class="rounded-md px-3.5 py-2 text-sm font-medium h-9 inline-flex items-center justify-center text-primary bg-surface border border-stroke transition-all duration-200 hover:not-disabled:border-stroke-hover hover:not-disabled:bg-surface-soft disabled:cursor-not-allowed disabled:opacity-50"
          type="button"
          @click="openChangePasswordModal"
        >
          Change Password
        </button>
        <button
          class="rounded-md px-3.5 py-2 text-sm font-medium h-9 inline-flex items-center justify-center text-primary bg-surface border border-stroke transition-all duration-200 hover:not-disabled:border-stroke-hover hover:not-disabled:bg-surface-soft disabled:cursor-not-allowed disabled:opacity-50"
          type="button"
          @click="handleLogout"
        >
          Logout
        </button>
      </div>
    </header>

    <ErrorBanner v-if="accountsStore.errorMessage" :message="accountsStore.errorMessage" />

    <AccountsPanel
      v-if="authStore.isAdmin"
      :accounts="accountsStore.accounts"
      :total="accountsStore.total"
      :page="accountsStore.page"
      :total-pages="accountsStore.totalPages"
      :can-prev="accountsStore.canPrev"
      :can-next="accountsStore.canNext"
      :footer-text="accountsStore.footerText"
      :loading="accountsStore.loading"
      :refreshed-at-text="refreshedAtText"
      :current-account-id="authStore.currentAccount?.account_id ?? ''"
      :deleting-account-id="accountsStore.deletingAccountID"
      :delete-button-text="accountsStore.deleteAccountButtonText"
      :format-date-time="accountsStore.formatDateTime"
      @refresh="accountsStore.loadAccounts(accountsStore.page)"
      @prev-page="accountsStore.previousPage"
      @next-page="accountsStore.nextPage"
      @delete-account="accountsStore.deleteAccount"
    />

    <CreateAccountModal v-if="showCreateAccountModal" @close="closeCreateAccountModal" />
    <ChangePasswordModal v-if="showChangePasswordModal" @close="closeChangePasswordModal" />
  </main>
</template>
