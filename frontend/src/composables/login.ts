import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import api from '@/service/axios'

export function useLogin() {
  const auth = useAuthStore()
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function login(username: string, password: string): Promise<void> {
    loading.value = true
    error.value = null
    try {
      const res = await api.post('/auth/login', { username, password })
      const token: string | undefined = res?.data?.access_token
      if (!token) {
        error.value = 'Login failed. Please try again.'
        return
      }
      auth.setToken(token)
    } catch (err: any) {
      const detail = err?.response?.data?.detail
      error.value = typeof detail === 'string' && detail.length > 0
        ? detail
        : 'Login failed. Please try again.'
    } finally {
      loading.value = false
    }
  }

  return {
    loading,
    error,
    login,
  }
}