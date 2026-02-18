<script setup lang="ts">
defineProps<{
  totalWorkers: number
  onlineWorkers: number
  offlineWorkers: number
  staleWorkers: number
  staleWorkersLabel: string
}>()
</script>

<template>
  <section class="stat-grid">
    <article class="stat-card">
      <p class="stat-label">Total Workers</p>
      <p class="stat-value">{{ totalWorkers }}</p>
    </article>
    <article class="stat-card online">
      <p class="stat-label">Online</p>
      <p class="stat-value">{{ onlineWorkers }}</p>
    </article>
    <article class="stat-card offline">
      <p class="stat-label">Offline</p>
      <p class="stat-value">{{ offlineWorkers }}</p>
    </article>
    <article class="stat-card stale">
      <p class="stat-label">{{ staleWorkersLabel }}</p>
      <p class="stat-value">{{ staleWorkers }}</p>
    </article>
  </section>
</template>

<style scoped>
.stat-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 14px;
}

.stat-card {
  background: var(--surface);
  border: 1px solid var(--stroke);
  border-radius: 18px;
  padding: 18px;
  box-shadow: 0 8px 30px rgba(12, 20, 29, 0.05);
  animation: rise-in 560ms ease-out;
}

.stat-card.online {
  border-top: 4px solid var(--online);
}

.stat-card.offline {
  border-top: 4px solid var(--offline);
}

.stat-card.stale {
  border-top: 4px solid var(--stale);
}

.stat-label {
  margin: 0;
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 500;
}

.stat-value {
  margin: 8px 0 0;
  font-size: clamp(1.8rem, 2vw, 2.3rem);
  font-weight: 700;
  letter-spacing: -0.03em;
}

@media (max-width: 960px) {
  .stat-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 620px) {
  .stat-grid {
    grid-template-columns: 1fr;
  }
}
</style>
