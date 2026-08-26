import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import MainLayout from '@/components/common/MainLayout.vue'
import type { UserRole } from '@/types/enums'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', component: () => import('@/pages/LoginPage.vue'), meta: { public: true } },
    {
      path: '/', component: MainLayout, redirect: '/ponds', children: [
        { path: 'ponds', component: () => import('@/pages/PondsPage.vue') },
        { path: 'readings', component: () => import('@/pages/ReadingsPage.vue') },
        { path: 'plans', component: () => import('@/pages/PlansPage.vue') },
        { path: 'executions', component: () => import('@/pages/ExecutionsPage.vue') },
        { path: 'audit', component: () => import('@/pages/AuditPage.vue'), meta: { roles: ['admin', 'manager'] satisfies UserRole[] } },
      ],
    },
    { path: '/:pathMatch(.*)*', redirect: '/ponds' },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (!auth.initialized) await auth.refresh()
  if (to.meta.public) return auth.authenticated ? '/ponds' : true
  if (!auth.authenticated) return { path: '/login', query: { redirect: to.fullPath } }
  const roles = to.meta.roles as UserRole[] | undefined
  if (roles && !auth.hasRole(...roles)) return '/ponds'
  return true
})

export default router
