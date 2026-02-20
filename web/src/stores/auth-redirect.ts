import router from '@/router'
import { useAuthStore } from '@/stores/auth'

export async function redirectToLogin(resetState: () => void): Promise<void> {
  const authStore = useAuthStore()
  authStore.logoutLocal()
  resetState()

  if (router.currentRoute.value.path !== '/login') {
    await router.replace({
      path: '/login',
      query: { redirect: router.currentRoute.value.fullPath },
    })
  }
}
