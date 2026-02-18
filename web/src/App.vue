<script setup lang="ts">
import { RouterView } from 'vue-router'
</script>

<template>
  <div class="dashboard-shell">
    <div class="gradient-orb orb-left" />
    <div class="gradient-orb orb-right" />

    <RouterView />
  </div>
</template>

<style scoped>
.dashboard-shell {
  position: relative;
  min-height: 100vh;
  overflow: hidden;
  padding: 36px 18px 48px;
}

.gradient-orb {
  position: absolute;
  z-index: 1;
  width: 440px;
  height: 440px;
  border-radius: 999px;
  filter: blur(60px);
  opacity: 0.35;
}

.orb-left {
  top: -180px;
  left: -120px;
  background: #ffe1b9;
}

.orb-right {
  top: -180px;
  right: -90px;
  background: #bfe8ff;
}

@media (max-width: 620px) {
  .dashboard-shell {
    padding: 22px 12px 32px;
  }
}
</style>
