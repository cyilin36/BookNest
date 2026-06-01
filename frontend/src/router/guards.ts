import type { Router } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

export function installRouterGuards(router: Router) {
  router.beforeEach(async (to) => {
    const auth = useAuthStore()
    await auth.restore()

    if ((to.name === 'login' || to.name === 'register') && auth.isAuthenticated) {
      return (to.query.redirect as string) || '/bookshelf'
    }

    if (to.meta.requiresAuth && !auth.isAuthenticated) {
      return { path: '/login', query: { redirect: to.fullPath } }
    }

    if (to.meta.requiresAdmin && !auth.isAdmin) {
      return { path: '/bookshelf', query: { forbidden: '1' } }
    }

    return true
  })
}
