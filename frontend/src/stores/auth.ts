import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { authApi } from '@/api/auth'
import type { User } from '@/types/models'
import type { UserRole } from '@/types/enums'

function storedUser(): User | null {
  try {
    const value = localStorage.getItem('aquaculture_user')
    return value ? JSON.parse(value) as User : null
  } catch {
    localStorage.removeItem('aquaculture_token')
    localStorage.removeItem('aquaculture_user')
    return null
  }
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref(localStorage.getItem('aquaculture_token') || '')
  const user = ref<User | null>(storedUser())
  const loading = ref(false)
  const initialized = ref(!token.value)
  const authenticated = computed(() => Boolean(token.value && user.value))

  async function login(username: string, password: string) {
    loading.value = true
    try {
      const result = await authApi.login(username, password)
      token.value = result.token
      user.value = result.user
      localStorage.setItem('aquaculture_token', result.token)
      localStorage.setItem('aquaculture_user', JSON.stringify(result.user))
      initialized.value = true
    } finally {
      loading.value = false
    }
  }

  async function refresh() {
    if (!token.value) { initialized.value = true; return false }
    loading.value = true
    try {
      user.value = await authApi.me()
      localStorage.setItem('aquaculture_user', JSON.stringify(user.value))
      return true
    } catch {
      logout()
      return false
    } finally {
      loading.value = false
      initialized.value = true
    }
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem('aquaculture_token')
    localStorage.removeItem('aquaculture_user')
    initialized.value = true
  }

  function hasRole(...roles: UserRole[]) {
    return Boolean(user.value && roles.includes(user.value.role))
  }

  window.addEventListener('auth-expired', logout)
  return { token, user, loading, initialized, authenticated, login, refresh, logout, hasRole }
})
