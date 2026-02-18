import { createRouter, createWebHistory } from 'vue-router'

import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: '/workers',
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginPage.vue'),
    },
    {
      path: '/workers',
      name: 'workers',
      component: () => import('@/views/WorkersPage.vue'),
      meta: {
        requiresAuth: true,
      },
    },
    {
      path: '/:pathMatch(.*)*',
      redirect: '/workers',
    },
  ],
})

router.beforeEach(async (to) => {
  const authStore = useAuthStore()

  try {
    await authStore.bootstrap()
  } catch {
    // fall through with unauthenticated state and let pages render their own errors
  }

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    return {
      path: '/login',
      query: {
        redirect: to.fullPath,
      },
    }
  }

  if (to.path === '/login' && authStore.isAuthenticated) {
    const redirect =
      typeof to.query.redirect === 'string' && to.query.redirect.startsWith('/')
        ? to.query.redirect
        : '/workers'
    return redirect
  }

  return true
})

export default router
