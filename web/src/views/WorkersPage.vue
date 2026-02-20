<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import ErrorBanner from '@/components/common/ErrorBanner.vue'
import DashboardHeader from '@/components/dashboard/DashboardHeader.vue'
import PaginationBar from '@/components/dashboard/PaginationBar.vue'
import StatsGrid from '@/components/dashboard/StatsGrid.vue'
import TrustedTokensPanel from '@/components/dashboard/TrustedTokensPanel.vue'
import WorkerCreateResultModal from '@/components/dashboard/WorkerCreateResultModal.vue'
import WorkersTable from '@/components/dashboard/WorkersTable.vue'
import WorkersToolbar from '@/components/dashboard/WorkersToolbar.vue'
import { createAccountAPI } from '@/services/auth.api'
import { useAuthStore } from '@/stores/auth'
import { useTokensStore } from '@/stores/tokens'
import { useWorkersStore } from '@/stores/workers'
import type { WorkerStartupCommandResponse, WorkerStatus } from '@/types/workers'

const authStore = useAuthStore()
const tokensStore = useTokensStore()
const workersStore = useWorkersStore()
const route = useRoute()
const router = useRouter()
const createdWorkerPayload = ref<WorkerStartupCommandResponse | null>(null)
const createAccountUsername = ref('')
const createAccountPassword = ref('')
const creatingAccount = ref(false)
const createAccountError = ref('')
const createAccountSuccess = ref('')

const showCreateAccountPanel = computed(() => authStore.isAdmin && authStore.registrationEnabled)

function parseStatus(raw: unknown): WorkerStatus {
  return raw === 'online' || raw === 'offline' || raw === 'all' ? raw : 'all'
}

function parsePage(raw: unknown): number {
  if (typeof raw !== 'string') {
    return 1
  }
  const parsed = Number.parseInt(raw, 10)
  if (!Number.isFinite(parsed) || parsed < 1) {
    return 1
  }
  return parsed
}

function syncStoreFromRoute(load: boolean): void {
  const targetStatus = parseStatus(route.query.status)
  const targetPage = parsePage(route.query.page)

  const statusChanged = targetStatus !== workersStore.statusFilter
  const pageChanged = targetPage !== workersStore.page
  if (!statusChanged && !pageChanged) {
    return
  }

  workersStore.setFilter(targetStatus, { load: false })
  workersStore.setPage(targetPage, { load: false })

  if (load) {
    void workersStore.loadDashboard()
  }
}

function syncRouteFromStore(): void {
  const currentStatus = parseStatus(route.query.status)
  const currentPage = parsePage(route.query.page)

  if (currentStatus === workersStore.statusFilter && currentPage === workersStore.page) {
    return
  }

  const query: Record<string, string> = {}
  if (workersStore.statusFilter !== 'all') {
    query.status = workersStore.statusFilter
  }
  if (workersStore.page > 1) {
    query.page = String(workersStore.page)
  }

  void router.replace({
    path: '/workers',
    query,
  })
}

const refreshedAtText = computed(() => {
  if (!workersStore.refreshedAt) {
    return 'never'
  }
  return workersStore.formatDateTime(workersStore.refreshedAt.toISOString())
})

function handleVisibilityChange(): void {
  workersStore.onPageVisibilityChange()
}

async function handleLogout(): Promise<void> {
  await authStore.logout()
  workersStore.teardown()
  tokensStore.teardown()
  tokensStore.reset()
  await router.replace('/login')
}

async function handleRefresh(): Promise<void> {
  await Promise.all([workersStore.loadDashboard(), tokensStore.loadTokens()])
}

async function handleAddWorker(): Promise<void> {
  const payload = await workersStore.createWorker()
  if (!payload) {
    return
  }
  createdWorkerPayload.value = payload
}

function closeWorkerCreateResultModal(): void {
  createdWorkerPayload.value = null
}

async function submitCreateAccount(): Promise<void> {
  if (creatingAccount.value) {
    return
  }
  const username = createAccountUsername.value.trim()
  const password = createAccountPassword.value
  if (!username || !password) {
    createAccountError.value = 'username and password are required'
    createAccountSuccess.value = ''
    return
  }

  createAccountError.value = ''
  createAccountSuccess.value = ''
  creatingAccount.value = true
  try {
    const payload = await createAccountAPI(username, password)
    createAccountUsername.value = ''
    createAccountPassword.value = ''
    createAccountSuccess.value = `Created account ${payload.account.username}`
  } catch (error) {
    createAccountError.value = error instanceof Error ? error.message : 'Failed to create account.'
  } finally {
    creatingAccount.value = false
  }
}

watch(
  () => route.query,
  () => {
    syncStoreFromRoute(true)
  },
)

watch(
  () => [workersStore.statusFilter, workersStore.page],
  () => {
    syncRouteFromStore()
  },
)

onMounted(async () => {
  syncStoreFromRoute(false)
  await Promise.all([workersStore.loadDashboard(), tokensStore.loadTokens()])
  workersStore.startAutoRefresh()
  document.addEventListener('visibilitychange', handleVisibilityChange)
})

onBeforeUnmount(() => {
  workersStore.teardown()
  tokensStore.teardown()
  document.removeEventListener('visibilitychange', handleVisibilityChange)
})
</script>

<template>
  <main class="dashboard-content">
    <DashboardHeader
      :creating-worker="workersStore.creatingWorker"
      :auto-refresh-enabled="workersStore.autoRefreshEnabled"
      :loading="workersStore.loading"
      @add-worker="handleAddWorker"
      @toggle-auto-refresh="workersStore.toggleAutoRefresh"
      @refresh="handleRefresh"
      @logout="handleLogout"
    />

    <StatsGrid
      :total-workers="workersStore.totalWorkers"
      :online-workers="workersStore.onlineWorkers"
      :offline-workers="workersStore.offlineWorkers"
      :stale-workers="workersStore.staleWorkers"
      :stale-workers-label="workersStore.staleWorkersLabel"
    />

    <TrustedTokensPanel
      :tokens="tokensStore.trustedTokens"
      :creating-token="tokensStore.creatingTrustedToken"
      :deleting-token-id="tokensStore.deletingTrustedTokenID"
      :delete-button-text="tokensStore.trustedTokenDeleteButtonText"
      :create-token="tokensStore.createTrustedToken"
      :format-date-time="tokensStore.formatDateTime"
      @delete-token="tokensStore.deleteTrustedToken"
    />

    <section v-if="showCreateAccountPanel" class="account-panel">
      <div class="account-panel-head">
        <h2>Create Account</h2>
        <p>Registration is enabled. New accounts are always non-admin.</p>
      </div>

      <form class="account-form" @submit.prevent="submitCreateAccount">
        <label class="field-label" for="create-account-username">Username</label>
        <input
          id="create-account-username"
          v-model="createAccountUsername"
          class="field-input"
          type="text"
          autocomplete="off"
          spellcheck="false"
        />

        <label class="field-label" for="create-account-password">Password</label>
        <input
          id="create-account-password"
          v-model="createAccountPassword"
          class="field-input"
          type="password"
          autocomplete="new-password"
        />

        <p v-if="createAccountError" class="account-error">{{ createAccountError }}</p>
        <p v-if="createAccountSuccess" class="account-success">{{ createAccountSuccess }}</p>

        <div class="account-actions">
          <button class="primary-btn small" type="submit" :disabled="creatingAccount">
            {{ creatingAccount ? 'Creating...' : 'Create Account' }}
          </button>
        </div>
      </form>
    </section>

    <section class="board-panel">
      <WorkersToolbar
        :status-filter="workersStore.statusFilter"
        :refreshed-at-text="refreshedAtText"
        @set-status="workersStore.setFilter"
      />

      <ErrorBanner
        v-if="workersStore.errorMessage"
        :message="workersStore.errorMessage"
        class="panel-error"
      />

      <WorkersTable
        :worker-rows="workersStore.workerRows"
        :loading="workersStore.loading"
        :deleting-node-id="workersStore.deletingNodeID"
        :format-capabilities="workersStore.formatCapabilities"
        :format-labels="workersStore.formatLabels"
        :format-date-time="workersStore.formatDateTime"
        :format-age="workersStore.formatAge"
        :delete-worker-button-text="workersStore.deleteWorkerButtonText"
        @delete-worker="workersStore.deleteWorker"
      />

      <PaginationBar
        :footer-text="workersStore.footerText"
        :page="workersStore.page"
        :total-pages="workersStore.totalPages"
        :can-prev="workersStore.canPrev"
        :can-next="workersStore.canNext"
        :loading="workersStore.loading"
        @prev="workersStore.previousPage"
        @next="workersStore.nextPage"
      />
    </section>

    <WorkerCreateResultModal
      :payload="createdWorkerPayload"
      @close="closeWorkerCreateResultModal"
    />
  </main>
</template>

<style scoped>
.dashboard-content {
  position: relative;
  z-index: 2;
  margin: 0 auto;
  width: min(1240px, 100%);
  display: grid;
  gap: 20px;
}

.board-panel {
  border: 1px solid var(--stroke);
  border-radius: 24px;
  background: var(--surface);
  box-shadow: var(--shadow);
  overflow: hidden;
  animation: rise-in 620ms ease-out;
}

.panel-error {
  margin: 16px 20px 0;
}

.account-panel {
  border: 1px solid var(--stroke);
  border-radius: 18px;
  background: var(--surface);
  box-shadow: var(--shadow);
  padding: 18px 20px;
  display: grid;
  gap: 14px;
}

.account-panel-head h2 {
  margin: 0;
  font-size: 1.1rem;
}

.account-panel-head p {
  margin: 6px 0 0;
  color: var(--text-secondary);
  font-size: 13px;
}

.account-form {
  display: grid;
  gap: 10px;
}

.account-error {
  margin: 0;
  border: 1px solid #ffccc7;
  border-radius: 10px;
  background: #fff4f2;
  color: #9f2f24;
  padding: 10px 12px;
  font-size: 13px;
}

.account-success {
  margin: 0;
  border: 1px solid #c7efd4;
  border-radius: 10px;
  background: #f3fff8;
  color: #0e5f3a;
  padding: 10px 12px;
  font-size: 13px;
}

.account-actions {
  display: flex;
  justify-content: flex-end;
}

@media (max-width: 620px) {
  .board-panel {
    border-radius: 16px;
  }
}
</style>
