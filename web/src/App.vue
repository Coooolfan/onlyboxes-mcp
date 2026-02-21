<script setup lang="ts">
import { computed } from 'vue'
import { RouterView } from 'vue-router'

import { defaultConsoleRepoURL, defaultConsoleVersion } from '@/constants/console'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

const consoleVersionText = computed(() => authStore.consoleVersion || defaultConsoleVersion)
const consoleRepoURL = computed(() => authStore.consoleRepoURL || defaultConsoleRepoURL)
</script>

<template>
  <div class="dashboard-shell">
    <div class="dashboard-main">
      <RouterView />
    </div>
    <footer class="console-footer">
      <span class="console-footer-version">Console {{ consoleVersionText }}</span>
      <span class="console-footer-separator">·</span>
      <a
        class="console-footer-link"
        :href="consoleRepoURL"
        target="_blank"
        rel="noopener noreferrer"
      >
        GitHub
      </a>
    </footer>
  </div>
</template>

<style scoped>
.dashboard-shell {
  position: relative;
  min-height: 100vh;
  padding: 32px 24px 20px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.dashboard-main {
  flex: 1;
}

.console-footer {
  margin: 0 auto;
  width: min(1240px, 100%);
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  color: var(--text-secondary);
  font-size: 12px;
  line-height: 1.5;
  font-family: 'JetBrains Mono', monospace;
}

.console-footer-link {
  color: var(--text-secondary);
  text-decoration: underline;
  text-underline-offset: 2px;
}

.console-footer-link:hover {
  color: var(--text-primary);
}

@media (max-width: 620px) {
  .dashboard-shell {
    padding: 24px 16px 16px;
  }

  .console-footer {
    justify-content: flex-start;
  }
}
</style>
