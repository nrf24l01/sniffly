import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'

export function useLogin() {
  const auth = useAuthStore()
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function login(username: string, password: string) {
    loading.value = true
    error.value = null

    try {
      await auth.login(username, password)
    } catch (err: any) {
      // Prefer structured server message, then thrown Error message, then fallback
      error.value = err?.response?.data?.detail ?? err?.response?.data?.message ?? err?.message ?? String(err) ?? 'Login failed'
    } finally {
      loading.value = false
    }
  }

  return { loading, error, login }
}
