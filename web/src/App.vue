<script setup lang="ts">
import { computed, defineAsyncComponent } from 'vue'
import { RouterView, useRoute } from 'vue-router'

const route = useRoute()
const DefaultLayout = defineAsyncComponent(() => import('@/layouts/DefaultLayout.vue'))
const DashboardLayout = defineAsyncComponent(() => import('@/layouts/DashboardLayout.vue'))

const layout = computed(() => {
  return route.meta.layout === 'dashboard' ? DashboardLayout : DefaultLayout
})
</script>

<template>
  <component :is="layout">
    <RouterView />
  </component>
</template>
