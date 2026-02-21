<script setup lang="ts">
import { computed } from 'vue'

import type { WorkerItem, WorkerInflightItem } from '@/types/workers'

const props = defineProps<{
  workerRows: WorkerItem[]
  inflightWorkers: WorkerInflightItem[]
  loading: boolean
  deletingNodeId: string
  formatCapabilities: (worker: WorkerItem) => string
  formatLabels: (worker: WorkerItem) => string
  formatDateTime: (value: string) => string
  formatAge: (value: string) => string
  deleteWorkerButtonText: (nodeID: string) => string
}>()

const emit = defineEmits<{
  deleteWorker: [nodeID: string]
}>()

type InflightCapability = WorkerInflightItem['capabilities'][number]

function normalizeCapabilityName(name: string): string {
  return name.trim().toLowerCase()
}

const inflightByWorker = computed(() => {
  const out = new Map<string, Map<string, InflightCapability>>()
  for (const worker of props.inflightWorkers) {
    const capabilities = new Map<string, InflightCapability>()
    for (const capability of worker.capabilities) {
      const normalized = normalizeCapabilityName(capability.name)
      if (!normalized) {
        continue
      }
      capabilities.set(normalized, capability)
    }
    out.set(worker.node_id, capabilities)
  }
  return out
})

function getInflight(nodeId: string, capName: string): InflightCapability | null {
  const normalized = normalizeCapabilityName(capName)
  if (!normalized) {
    return null
  }
  return inflightByWorker.value.get(nodeId)?.get(normalized) ?? null
}
</script>

<template>
  <div class="table-wrap">
    <table>
      <thead>
        <tr>
          <th>Node</th>
          <th>Runtime</th>
          <th>Capabilities</th>
          <th>Labels</th>
          <th>Status</th>
          <th>Registered / Last Heartbeat</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-if="!loading && workerRows.length === 0">
          <td colspan="7" class="empty-cell">No workers found in current filter.</td>
        </tr>
        <tr v-for="worker in workerRows" :key="worker.node_id">
          <td>
            <div class="node-main">{{ worker.node_name || worker.node_id }}</div>
            <div class="node-sub">{{ worker.node_id }}</div>
          </td>
          <td>
            <div class="runtime-main">{{ worker.executor_kind || '--' }}</div>
            <div class="runtime-sub">version: {{ worker.version || '--' }}</div>
          </td>
          <td>
            <div
              class="capabilities-list"
              v-if="worker.capabilities && worker.capabilities.length > 0"
            >
              <span class="capability-badge" v-for="cap in worker.capabilities" :key="cap.name">
                {{ cap.name }}
                <span
                  v-if="getInflight(worker.node_id, cap.name)"
                  class="inflight-tag"
                  :class="{
                    active: getInflight(worker.node_id, cap.name)!.inflight > 0,
                    full:
                      getInflight(worker.node_id, cap.name)!.inflight >=
                      getInflight(worker.node_id, cap.name)!.max_inflight,
                  }"
                >
                  {{ getInflight(worker.node_id, cap.name)!.inflight }}/{{
                    getInflight(worker.node_id, cap.name)!.max_inflight
                  }}
                </span>
              </span>
            </div>
            <span v-else>--</span>
          </td>
          <td>{{ formatLabels(worker) }}</td>
          <td>
            <span :class="['status-pill', worker.status]">{{ worker.status }}</span>
          </td>
          <td>
            <div class="time-main">{{ formatDateTime(worker.registered_at) }}</div>
            <div class="time-sub">{{ formatAge(worker.last_seen_at) }}</div>
          </td>
          <td>
            <div class="row-actions">
              <button
                type="button"
                class="ghost-btn small danger"
                :disabled="deletingNodeId === worker.node_id"
                @click="emit('deleteWorker', worker.node_id)"
              >
                {{ deleteWorkerButtonText(worker.node_id) }}
              </button>
            </div>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.table-wrap {
  overflow: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
  min-width: 920px;
}

th,
td {
  text-align: left;
  padding: 16px 24px;
  border-bottom: 1px solid var(--stroke);
  vertical-align: middle;
}

th {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-secondary);
  background: var(--surface-soft);
  position: sticky;
  top: 0;
  z-index: 1;
}

td {
  font-size: 14px;
  color: var(--text-primary);
}

tr {
  transition: background-color 0.2s ease;
}

tr:hover {
  background-color: var(--surface-soft);
}

.row-actions {
  display: inline-flex;
  gap: 8px;
  align-items: center;
}

.node-main {
  font-weight: 600;
}

.capabilities-list {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.capability-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 4px 8px;
  background: var(--surface-soft);
  border: 1px solid var(--stroke);
  border-radius: var(--radius);
  font-family: 'JetBrains Mono', monospace;
  font-size: 11px;
  color: var(--text-secondary);
  gap: 6px;
}

.inflight-tag {
  font-size: 10px;
  padding: 1px 4px;
  border-radius: 4px;
  background: var(--surface);
  border: 1px solid var(--stroke);
  color: var(--text-tertiary);
}

.inflight-tag.active {
  color: var(--text-primary);
  border-color: var(--stroke-hover);
  background: var(--surface-soft);
}

.inflight-tag.full {
  color: #b45309;
  background: #fffbeb;
  border-color: #fcd34d;
}

.node-sub {
  margin-top: 4px;
  color: var(--text-secondary);
  font-family: 'JetBrains Mono', monospace;
  font-size: 12px;
}

.runtime-main {
  font-size: 14px;
}

.runtime-sub {
  margin-top: 4px;
  color: var(--text-secondary);
  font-family: 'JetBrains Mono', monospace;
  font-size: 12px;
}

.time-main {
  font-size: 14px;
}

.time-sub {
  margin-top: 4px;
  color: var(--text-secondary);
  font-family: 'JetBrains Mono', monospace;
  font-size: 12px;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius);
  padding: 4px 10px;
  font-size: 12px;
  font-weight: 500;
  text-transform: capitalize;
}

.status-pill.online {
  color: #166534;
  background: #f0fdf4;
  border: 1px solid #bbf7d0;
}

.status-pill.offline {
  color: #991b1b;
  background: #fef2f2;
  border: 1px solid #fecaca;
}

.empty-cell {
  color: var(--text-secondary);
  text-align: center;
  padding: 48px 24px;
}
</style>
