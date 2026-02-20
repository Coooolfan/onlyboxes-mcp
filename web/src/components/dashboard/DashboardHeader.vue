<script setup lang="ts">
defineProps<{
  creatingWorker: boolean
  autoRefreshEnabled: boolean
  loading: boolean
}>()

const emit = defineEmits<{
  addWorker: []
  toggleAutoRefresh: []
  refresh: []
  logout: []
}>()
</script>

<template>
  <header class="dashboard-header">
    <div>
      <p class="eyebrow">Onlyboxes / Worker Registry</p>
      <h1>Execution Node Control Panel</h1>
      <p class="subtitle">Real-time monitoring for worker registration and heartbeat health.</p>
    </div>

    <div class="header-actions">
      <button
        class="primary-btn"
        type="button"
        :disabled="creatingWorker"
        @click="emit('addWorker')"
      >
        {{ creatingWorker ? 'Adding...' : 'Add Worker' }}
      </button>
      <button class="ghost-btn" type="button" @click="emit('toggleAutoRefresh')">
        {{ autoRefreshEnabled ? 'Auto Refresh: ON' : 'Auto Refresh: OFF' }}
      </button>
      <button class="primary-btn" type="button" :disabled="loading" @click="emit('refresh')">
        {{ loading ? 'Refreshing...' : 'Refresh Now' }}
      </button>
      <button class="ghost-btn" type="button" @click="emit('logout')">Logout</button>
    </div>
  </header>
</template>

<style scoped>
.dashboard-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20px;
  background: linear-gradient(135deg, #fcfdff 0%, #f2f5f9 100%);
  border: 1px solid var(--stroke);
  border-radius: 24px;
  padding: 26px 28px;
  box-shadow: var(--shadow);
  animation: rise-in 500ms ease-out;
}

.eyebrow {
  margin: 0;
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--text-secondary);
}

h1 {
  margin: 8px 0 10px;
  font-size: clamp(1.8rem, 3.3vw, 2.8rem);
  line-height: 1.1;
  letter-spacing: -0.02em;
}

.subtitle {
  margin: 0;
  color: var(--text-secondary);
  max-width: 62ch;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

@media (max-width: 960px) {
  .dashboard-header {
    flex-direction: column;
  }

  .header-actions {
    width: 100%;
    flex-wrap: wrap;
  }
}
</style>
