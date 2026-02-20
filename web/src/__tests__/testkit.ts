import { flushPromises, mount } from '@vue/test-utils'
import { createPinia } from 'pinia'

import App from '../App.vue'
import router from '../router'

export const adminSessionPayload = {
  authenticated: true,
  account: {
    account_id: 'acc-admin',
    username: 'admin-test',
    is_admin: true,
  },
  registration_enabled: true,
}

export const memberSessionPayload = {
  authenticated: true,
  account: {
    account_id: 'acc-member',
    username: 'member-test',
    is_admin: false,
  },
  registration_enabled: false,
}

export const statsPayload = {
  total: 150,
  online: 120,
  offline: 30,
  stale: 10,
  stale_after_sec: 45,
  generated_at: '2026-02-16T10:00:10Z',
}

export const workersPayload = {
  items: [
    {
      node_id: 'node-1',
      node_name: 'worker-1',
      executor_kind: 'docker',
      capabilities: [{ name: 'echo' }],
      labels: { zone: 'a' },
      version: 'v0.1.0',
      status: 'online',
      registered_at: '2026-02-16T10:00:00Z',
      last_seen_at: '2026-02-16T10:00:05Z',
    },
  ],
  total: 1,
  page: 1,
  page_size: 25,
}

export function defaultTokensPayload() {
  return {
    items: [
      {
        id: 'tok-1',
        name: 'ci-prod',
        token_masked: 'obx_****9fa1',
        created_at: '2026-02-16T10:00:00Z',
        updated_at: '2026-02-16T10:00:00Z',
      },
    ],
    total: 1,
  }
}

export function jsonResponse(payload: unknown) {
  return {
    ok: true,
    status: 200,
    statusText: 'OK',
    json: async () => payload,
  }
}

export function noContentResponse() {
  return {
    ok: true,
    status: 204,
    statusText: 'No Content',
    json: async () => ({}),
  }
}

export function unauthorizedResponse() {
  return {
    ok: false,
    status: 401,
    statusText: 'Unauthorized',
    json: async () => ({ error: 'authentication required' }),
  }
}

export async function mountApp(path: string) {
  const pinia = createPinia()
  const wrapper = mount(App, {
    global: {
      plugins: [pinia, router],
    },
  })

  await router.push('/login')
  await flushPromises()
  await router.push(path)
  await flushPromises()

  return wrapper
}

export async function waitForRoute(path: string, maxAttempts = 8) {
  for (let i = 0; i < maxAttempts; i += 1) {
    if (router.currentRoute.value.path === path) {
      return
    }
    await flushPromises()
    await new Promise<void>((resolve) => {
      setTimeout(resolve, 0)
    })
  }
}
