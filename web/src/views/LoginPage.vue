<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { isInvalidCredentialsError } from '@/services/auth.api'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const route = useRoute()
const router = useRouter()

const loginUsername = ref('')
const loginPassword = ref('')
const loginErrorMessage = ref('')
const loginSubmitting = ref(false)

function resolveRedirect(): string {
  const redirect = route.query.redirect
  if (typeof redirect === 'string' && redirect.startsWith('/')) {
    return redirect
  }
  return '/workers'
}

async function submitLogin(): Promise<void> {
  if (loginSubmitting.value) {
    return
  }

  loginErrorMessage.value = ''
  if (loginUsername.value.trim() === '' || loginPassword.value === '') {
    loginErrorMessage.value = '请输入账号和密码。'
    return
  }

  loginSubmitting.value = true
  try {
    await authStore.login(loginUsername.value, loginPassword.value)
    await router.replace(resolveRedirect())
  } catch (error) {
    if (isInvalidCredentialsError(error)) {
      loginErrorMessage.value = '账号或密码错误'
    } else {
      loginErrorMessage.value = error instanceof Error ? error.message : '登录失败，请稍后重试。'
    }
  } finally {
    loginSubmitting.value = false
  }
}
</script>

<template>
  <section class="auth-panel">
    <p class="eyebrow">Onlyboxes / Console Login</p>
    <h1>Sign In to Control Panel</h1>
    <p class="subtitle">Use the dashboard username and password printed in the console startup logs.</p>

    <form class="login-form" @submit.prevent="submitLogin">
      <label class="field-label" for="dashboard-username">Username</label>
      <input
        id="dashboard-username"
        v-model="loginUsername"
        class="field-input"
        type="text"
        name="username"
        autocomplete="username"
        spellcheck="false"
      />

      <label class="field-label" for="dashboard-password">Password</label>
      <input
        id="dashboard-password"
        v-model="loginPassword"
        class="field-input"
        type="password"
        name="password"
        autocomplete="current-password"
      />

      <p v-if="loginErrorMessage" class="auth-error">{{ loginErrorMessage }}</p>

      <button class="primary-btn" type="submit" :disabled="loginSubmitting">
        {{ loginSubmitting ? 'Signing In...' : 'Sign In' }}
      </button>
    </form>
  </section>
</template>

<style scoped>
.auth-panel {
  width: min(520px, 100%);
  margin: 48px auto 0;
  background: linear-gradient(135deg, #fcfdff 0%, #f2f5f9 100%);
  border: 1px solid var(--stroke);
  border-radius: 24px;
  padding: 28px;
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

.login-form {
  margin-top: 18px;
  display: grid;
  gap: 12px;
}

.field-label {
  font-family: 'IBM Plex Mono', monospace;
  font-size: 12px;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: var(--text-secondary);
}

.field-input {
  width: 100%;
  border: 1px solid var(--stroke);
  border-radius: 12px;
  background: #fff;
  color: var(--text-primary);
  padding: 10px 12px;
  font: inherit;
  outline: none;
  transition: border-color 140ms ease, box-shadow 140ms ease;
}

.field-input:focus {
  border-color: #5b7cff;
  box-shadow: 0 0 0 3px rgba(91, 124, 255, 0.18);
}

.auth-error {
  margin: 0;
  border: 1px solid #ffccc7;
  border-radius: 10px;
  background: #fff4f2;
  color: #9f2f24;
  padding: 10px 12px;
  font-size: 13px;
}
</style>
