<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue'

import type { TrustedTokenCreateResponse, TrustedTokenItem } from '@/types/workers'
import { writeTextToClipboard } from '@/utils/clipboard'

const props = defineProps<{
  tokens: TrustedTokenItem[]
  creatingToken: boolean
  deletingTokenId: string
  deleteButtonText: (tokenID: string) => string
  createToken: (payload: { name: string }) => Promise<TrustedTokenCreateResponse>
  formatDateTime: (value: string) => string
}>()

const emit = defineEmits<{
  deleteToken: [tokenID: string]
}>()

const collapsed = ref(true)
const showCreateModal = ref(false)
const nameInput = ref('')
const modalError = ref('')
const createdToken = ref<TrustedTokenCreateResponse | null>(null)
const copyingCreatedToken = ref(false)
const copiedCreatedToken = ref(false)
const copyFailed = ref(false)

let createdTokenCopyFeedbackTimer: ReturnType<typeof setTimeout> | null = null

const createdTokenCopyButtonText = computed(() => {
  if (copyingCreatedToken.value) {
    return 'Copying...'
  }
  if (copiedCreatedToken.value) {
    return 'Copied'
  }
  if (copyFailed.value) {
    return 'Copy Failed'
  }
  return 'Copy Token'
})

function resetCreatedTokenCopyFeedback(): void {
  if (createdTokenCopyFeedbackTimer) {
    clearTimeout(createdTokenCopyFeedbackTimer)
    createdTokenCopyFeedbackTimer = null
  }
  copyingCreatedToken.value = false
  copiedCreatedToken.value = false
  copyFailed.value = false
}

function clearSensitiveToken(): void {
  createdToken.value = null
  resetCreatedTokenCopyFeedback()
}

function openCreateModal(): void {
  showCreateModal.value = true
  modalError.value = ''
  nameInput.value = ''
  clearSensitiveToken()
}

function closeCreateModal(): void {
  showCreateModal.value = false
  modalError.value = ''
  nameInput.value = ''
  clearSensitiveToken()
}

function scheduleCreatedTokenCopyFeedbackReset(): void {
  if (createdTokenCopyFeedbackTimer) {
    clearTimeout(createdTokenCopyFeedbackTimer)
  }
  createdTokenCopyFeedbackTimer = setTimeout(() => {
    copiedCreatedToken.value = false
    copyFailed.value = false
    createdTokenCopyFeedbackTimer = null
  }, 1500)
}

async function submitCreateToken(): Promise<void> {
  if (props.creatingToken) {
    return
  }

  const name = nameInput.value.trim()
  if (!name) {
    modalError.value = 'name is required'
    return
  }

  modalError.value = ''

  try {
    const payload = await props.createToken({ name })
    const tokenValue = payload.token.trim()
    if (!tokenValue) {
      throw new Error('API returned empty token value.')
    }
    createdToken.value = {
      ...payload,
      token: tokenValue,
    }
  } catch (error) {
    modalError.value = error instanceof Error ? error.message : 'Failed to create trusted token.'
  }
}

async function copyCreatedToken(): Promise<void> {
  const tokenValue = createdToken.value?.token?.trim() ?? ''
  if (!tokenValue || copyingCreatedToken.value) {
    return
  }

  resetCreatedTokenCopyFeedback()
  copyingCreatedToken.value = true
  try {
    await writeTextToClipboard(tokenValue, {
      fallbackErrorMessage: 'Failed to copy token.',
    })
    copiedCreatedToken.value = true
    scheduleCreatedTokenCopyFeedbackReset()
  } catch {
    copyFailed.value = true
    scheduleCreatedTokenCopyFeedbackReset()
  } finally {
    copyingCreatedToken.value = false
  }
}

onBeforeUnmount(() => {
  clearSensitiveToken()
})
</script>

<template>
  <section class="token-panel">
    <div class="token-panel-header">
      <div class="token-title-block">
        <h2>Trusted Tokens</h2>
        <p>Total: {{ tokens.length }}</p>
      </div>
      <div class="token-header-actions">
        <button type="button" class="primary-btn small" @click="openCreateModal">New Token</button>
        <button type="button" class="ghost-btn small" @click="collapsed = !collapsed">
          {{ collapsed ? 'Expand' : 'Collapse' }}
        </button>
      </div>
    </div>

    <p v-if="collapsed" class="collapsed-hint">列表已折叠。展开后查看所有 Token 简要信息。</p>

    <template v-else>
      <p v-if="tokens.length === 0" class="empty-hint">未配置，MCP 与受保护 HTTP 端点当前全部拒绝。</p>

      <ul v-else class="token-list">
        <li v-for="item in tokens" :key="item.id" class="token-item">
          <div class="token-summary">
            <p class="token-name">{{ item.name }}</p>
            <p class="token-meta">
              <span class="token-meta-label">ID</span>
              <code>{{ item.id }}</code>
            </p>
            <p class="token-meta">
              <span class="token-meta-label">Masked</span>
              <code>{{ item.token_masked }}</code>
            </p>
            <p class="token-meta">
              <span class="token-meta-label">Created</span>
              <span>{{ formatDateTime(item.created_at) }}</span>
            </p>
          </div>

          <div class="token-actions">
            <button
              type="button"
              class="ghost-btn small danger"
              :disabled="deletingTokenId === item.id"
              @click="emit('deleteToken', item.id)"
            >
              {{ deleteButtonText(item.id) }}
            </button>
          </div>
        </li>
      </ul>
    </template>
  </section>

  <div v-if="showCreateModal" class="token-modal-backdrop" @click.self="closeCreateModal">
    <div class="token-modal" role="dialog" aria-modal="true" aria-labelledby="trusted-token-dialog-title">
      <div class="token-modal-header">
        <h3 id="trusted-token-dialog-title">{{ createdToken ? 'Token Created' : 'New Trusted Token' }}</h3>
        <button type="button" class="ghost-btn small" @click="closeCreateModal">
          {{ createdToken ? 'Done' : 'Cancel' }}
        </button>
      </div>

      <template v-if="!createdToken">
        <p class="token-modal-copy">创建后会展示一次明文 token，关闭后不可再次查看。</p>
        <form class="token-modal-form" @submit.prevent="submitCreateToken">
          <label class="token-field">
            <span>Name</span>
            <input v-model="nameInput" type="text" maxlength="64" required placeholder="ci-prod" />
          </label>

          <p v-if="modalError" class="token-modal-error">{{ modalError }}</p>

          <div class="token-modal-actions">
            <button type="button" class="ghost-btn small" :disabled="creatingToken" @click="closeCreateModal">Cancel</button>
            <button type="submit" class="primary-btn small" :disabled="creatingToken || nameInput.trim() === ''">
              {{ creatingToken ? 'Creating...' : 'Create Token' }}
            </button>
          </div>
        </form>
      </template>

      <template v-else>
        <p class="token-modal-copy">这是唯一一次展示明文 token，请立即复制并安全保存。</p>
        <code class="token-plain-value">{{ createdToken.token }}</code>
        <div class="token-result-meta">
          <p><span>Name</span>{{ createdToken.name }}</p>
          <p><span>ID</span>{{ createdToken.id }}</p>
          <p><span>Masked</span>{{ createdToken.token_masked }}</p>
        </div>

        <div class="token-modal-actions">
          <button type="button" class="ghost-btn small" :disabled="copyingCreatedToken" @click="copyCreatedToken">
            {{ createdTokenCopyButtonText }}
          </button>
          <button type="button" class="primary-btn small" @click="closeCreateModal">Done</button>
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped>
.token-panel {
  border: 1px solid var(--stroke);
  border-radius: 18px;
  background: var(--surface);
  box-shadow: var(--shadow);
  padding: 16px 18px;
  animation: rise-in 540ms ease-out;
}

.token-panel-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.token-title-block h2 {
  margin: 0;
  font-size: 1.1rem;
}

.token-title-block p {
  margin: 6px 0 0;
  color: var(--text-secondary);
  font-size: 13px;
}

.token-header-actions {
  display: flex;
  gap: 8px;
}

.collapsed-hint,
.empty-hint {
  margin: 14px 0 0;
  color: var(--text-secondary);
  font-size: 13px;
}

.token-list {
  list-style: none;
  margin: 14px 0 0;
  padding: 0;
  display: grid;
  gap: 10px;
}

.token-item {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  border: 1px solid var(--stroke);
  background: linear-gradient(180deg, #ffffff 0%, #fafbfd 100%);
  border-radius: 14px;
  padding: 12px;
}

.token-summary {
  min-width: 0;
  display: grid;
  gap: 6px;
}

.token-name {
  margin: 0;
  font-size: 14px;
  font-weight: 700;
}

.token-meta {
  margin: 0;
  display: flex;
  align-items: baseline;
  gap: 8px;
  color: var(--text-primary);
  font-size: 13px;
}

.token-meta-label {
  width: 52px;
  color: var(--text-secondary);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.token-meta code {
  font-family: 'IBM Plex Mono', monospace;
  background: #f1f4f8;
  border-radius: 6px;
  padding: 2px 6px;
  word-break: break-all;
}

.token-modal-backdrop {
  position: fixed;
  inset: 0;
  z-index: 1000;
  background: rgba(10, 14, 22, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 16px;
}

.token-modal {
  width: min(560px, 100%);
  border-radius: 16px;
  border: 1px solid var(--stroke);
  background: var(--surface);
  box-shadow: 0 24px 80px rgba(10, 14, 22, 0.26);
  padding: 18px;
  display: grid;
  gap: 12px;
}

.token-modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.token-modal-header h3 {
  margin: 0;
  font-size: 1.1rem;
}

.token-modal-copy {
  margin: 0;
  color: var(--text-secondary);
  font-size: 13px;
}

.token-modal-form {
  display: grid;
  gap: 10px;
}

.token-field {
  display: grid;
  gap: 6px;
}

.token-field span {
  color: var(--text-secondary);
  font-size: 12px;
}

.token-field input {
  border: 1px solid var(--stroke);
  border-radius: 10px;
  padding: 9px 10px;
  font-size: 13px;
}

.token-field input:focus {
  outline: none;
  border-color: #3f8cff;
  box-shadow: 0 0 0 3px rgba(63, 140, 255, 0.15);
}

.token-modal-error {
  margin: 0;
  border: 1px solid #ffccc7;
  border-radius: 10px;
  background: #fff4f2;
  color: #9f2f24;
  padding: 8px 10px;
  font-size: 13px;
}

.token-modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.token-plain-value {
  display: block;
  border: 1px solid var(--stroke);
  border-radius: 10px;
  background: #f8faff;
  padding: 10px;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 13px;
  line-height: 1.5;
  word-break: break-all;
}

.token-result-meta {
  display: grid;
  gap: 6px;
}

.token-result-meta p {
  margin: 0;
  display: flex;
  align-items: baseline;
  gap: 8px;
  font-size: 13px;
}

.token-result-meta span {
  width: 52px;
  color: var(--text-secondary);
  font-family: 'IBM Plex Mono', monospace;
  font-size: 11px;
  text-transform: uppercase;
}

@media (max-width: 700px) {
  .token-panel-header {
    flex-direction: column;
  }

  .token-item {
    flex-direction: column;
  }

  .token-modal-actions {
    justify-content: stretch;
  }

  .token-modal-actions button {
    flex: 1;
  }
}
</style>
