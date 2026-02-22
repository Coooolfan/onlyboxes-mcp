<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted } from 'vue'

import ErrorBanner from '@/components/common/ErrorBanner.vue'
import ConsoleHeader from '@/components/dashboard/ConsoleHeader.vue'
import TrustedTokensPanel from '@/components/dashboard/TrustedTokensPanel.vue'
import { useAuthStore } from '@/stores/auth'
import { useTokensStore } from '@/stores/tokens'

const authStore = useAuthStore()
const tokensStore = useTokensStore()

const refreshedAtText = computed(() => {
  if (!tokensStore.refreshedAt) {
    return 'never'
  }
  return tokensStore.formatDateTime(tokensStore.refreshedAt.toISOString())
})

onMounted(async () => {
  await tokensStore.loadTokens()
})

onBeforeUnmount(() => {
  tokensStore.teardown()
})
</script>

<template>
  <main class="relative z-2 mx-auto w-[min(1240px,100%)] grid gap-6">
    <ConsoleHeader
      eyebrow="Onlyboxes / Token Console"
      title="Trusted Token Management"
      :loading="tokensStore.loading"
      :refreshed-at-text="refreshedAtText"
      @refresh="tokensStore.loadTokens"
    >
      <template #subtitle>
        Account: <strong>{{ authStore.currentAccount?.username ?? '--' }}</strong>
      </template>
    </ConsoleHeader>

    <ErrorBanner v-if="tokensStore.errorMessage" :message="tokensStore.errorMessage" />

    <TrustedTokensPanel
      :tokens="tokensStore.trustedTokens"
      :creating-token="tokensStore.creatingTrustedToken"
      :deleting-token-id="tokensStore.deletingTrustedTokenID"
      :delete-button-text="tokensStore.trustedTokenDeleteButtonText"
      :create-token="tokensStore.createTrustedToken"
      :format-date-time="tokensStore.formatDateTime"
      @delete-token="tokensStore.deleteTrustedToken"
    />
  </main>
</template>
