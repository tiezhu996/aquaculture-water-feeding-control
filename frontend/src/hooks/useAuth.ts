import { storeToRefs } from 'pinia'
import { useAuthStore } from '@/stores/auth'
import type { UserRole } from '@/types/enums'

export function useAuth() {
  const store = useAuthStore()
  const { user, authenticated, loading } = storeToRefs(store)
  const canManagePonds = () => store.hasRole('admin', 'manager')
  const canOperate = () => store.hasRole('admin', 'manager', 'operator')
  const canReview = () => store.hasRole('admin', 'manager')
  const can = (...roles: UserRole[]) => store.hasRole(...roles)
  return { user, authenticated, loading, canManagePonds, canOperate, canReview, can, logout: store.logout }
}
