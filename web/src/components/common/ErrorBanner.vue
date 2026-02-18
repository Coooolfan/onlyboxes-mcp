<script setup lang="ts">
defineProps<{
  message: string
}>()
</script>

<template>
  <p class="error-banner">{{ message }}</p>
</template>

<style scoped>
.error-banner {
  margin: 0;
  border: 1px solid #ffccc7;
  border-radius: 12px;
  background: #fff4f2;
  color: #9f2f24;
  padding: 10px 12px;
  font-size: 13px;
}
</style>
