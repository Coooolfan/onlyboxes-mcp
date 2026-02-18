<script setup lang="ts">
defineProps<{
  tokens: string[]
  copyingToken: string
  copyButtonText: (token: string) => string
}>()

const emit = defineEmits<{
  copyToken: [token: string]
}>()
</script>

<template>
  <section class="token-panel">
    <div class="token-panel-header">
      <h2>MCP Token Whitelist</h2>
      <p>Total: {{ tokens.length }}</p>
    </div>

    <p v-if="tokens.length === 0" class="empty-hint">未配置，MCP 当前全部拒绝。</p>

    <ul v-else class="token-list">
      <li v-for="token in tokens" :key="token" class="token-item">
        <code class="token-value">{{ token }}</code>
        <button
          type="button"
          class="ghost-btn small"
          :disabled="copyingToken === token"
          @click="emit('copyToken', token)"
        >
          {{ copyButtonText(token) }}
        </button>
      </li>
    </ul>
  </section>
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
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
}

h2 {
  margin: 0;
  font-size: 1.1rem;
}

.token-panel-header p {
  margin: 0;
  color: var(--text-secondary);
  font-size: 13px;
}

.empty-hint {
  margin: 12px 0 0;
  color: var(--text-secondary);
  font-size: 13px;
}

.token-list {
  list-style: none;
  margin: 14px 0 0;
  padding: 0;
  display: grid;
  gap: 8px;
}

.token-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border: 1px solid var(--stroke);
  background: #fafbfd;
  border-radius: 12px;
  padding: 10px 12px;
}

.token-value {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  color: #1b2230;
  word-break: break-all;
}
</style>
