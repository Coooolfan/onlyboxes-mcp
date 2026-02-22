<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

import { useAuthStore } from '@/stores/auth'

interface ConsoleRouteTab {
  label: string
  to: string
  requiresAdmin?: boolean
}

const authStore = useAuthStore()
const route = useRoute()

const tabs = computed<ConsoleRouteTab[]>(() => {
  const items: ConsoleRouteTab[] = [
    {
      label: 'Workers',
      to: '/workers',
      requiresAdmin: true,
    },
    {
      label: 'Accounts',
      to: '/accounts',
      requiresAdmin: true,
    },
    {
      label: 'Tokens',
      to: '/tokens',
    },
  ]
  return items.filter((item) => !item.requiresAdmin || authStore.isAdmin)
})

function isActive(path: string): boolean {
  return route.path === path
}
</script>

<template>
  <nav
    class="flex items-center gap-2 rounded-lg border border-stroke bg-surface-soft px-2 py-2 shadow-card overflow-x-auto"
    aria-label="Console navigation"
  >
    <RouterLink
      v-for="item in tabs"
      :key="item.to"
      :to="item.to"
      class="rounded-md px-3.5 py-2 text-sm font-medium h-9 inline-flex items-center justify-center border transition-all duration-200 whitespace-nowrap"
      :class="
        isActive(item.to)
          ? 'text-white bg-accent border-accent'
          : 'text-primary bg-surface border-stroke hover:border-stroke-hover hover:bg-surface-soft'
      "
    >
      {{ item.label }}
    </RouterLink>
  </nav>
</template>
