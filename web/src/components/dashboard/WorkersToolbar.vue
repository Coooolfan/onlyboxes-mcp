<script setup lang="ts">
import type { WorkerStatus } from '@/types/workers'

defineProps<{
  statusFilter: WorkerStatus
  refreshedAtText: string
}>()

const emit = defineEmits<{
  setStatus: [status: WorkerStatus]
}>()
</script>

<template>
  <div class="panel-topbar">
    <div class="tabs">
      <button
        type="button"
        :class="['tab-btn', { active: statusFilter === 'all' }]"
        @click="emit('setStatus', 'all')"
      >
        All
      </button>
      <button
        type="button"
        :class="['tab-btn', { active: statusFilter === 'online' }]"
        @click="emit('setStatus', 'online')"
      >
        Online
      </button>
      <button
        type="button"
        :class="['tab-btn', { active: statusFilter === 'offline' }]"
        @click="emit('setStatus', 'offline')"
      >
        Offline
      </button>
    </div>

    <p class="panel-meta">
      Last refresh:
      <span>{{ refreshedAtText }}</span>
    </p>
  </div>
</template>

<style scoped>
.panel-topbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 18px 22px;
  border-bottom: 1px solid var(--stroke);
  background: linear-gradient(180deg, #ffffff 0%, #f8f9fc 100%);
}

.tabs {
  display: inline-flex;
  gap: 8px;
  background: var(--surface-soft);
  border-radius: 999px;
  padding: 4px;
}

.tab-btn {
  border-radius: 999px;
  padding: 8px 14px;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
  background: transparent;
}

.tab-btn.active {
  color: #ffffff;
  background: var(--accent);
}

.panel-meta {
  margin: 0;
  color: var(--text-secondary);
  font-size: 13px;
}

.panel-meta span {
  color: var(--text-primary);
  font-weight: 600;
}

@media (max-width: 960px) {
  .panel-topbar {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
