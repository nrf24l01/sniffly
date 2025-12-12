import axios, { type AxiosRequestConfig, type AxiosResponse, type AxiosError, type AxiosRequestHeaders } from 'axios'
import { useAuthStore } from '@/stores/auth'

const api = axios.create({
  baseURL: '/api',
  withCredentials: true
})

// Extract refresh token function to make it reusable
export async function refreshAccessToken() {
  const res = await axios.post(import.meta.env.VITE_BACKEND_URL + '/auth/refresh', {}, {
    withCredentials: true
  })
  return res.data.access_token
}

api.interceptors.request.use(config => {
  const auth = useAuthStore()
  // Resolve token value safely whether the store exposes a plain string or a ref
  const token = (auth as any).accessToken?.value ?? (auth as any).accessToken
  if (token) {
    if (!config.headers) config.headers = {} as AxiosRequestHeaders
    (config.headers as AxiosRequestHeaders)['Authorization'] = 'Bearer ' + token
  }
  return config
})

api.interceptors.response.use(
  (response: AxiosResponse) => response,
  async (error: unknown) => {
    const auth = useAuthStore()

    if (axios.isAxiosError(error)) {
      const status = error.response?.status
      if (status === 401) {
        try {
          const newToken = await refreshAccessToken()
          auth.setToken(newToken)

          const config = error.config as AxiosRequestConfig | undefined
          if (!config) return Promise.reject(error)
          if (!config.headers) config.headers = {} as AxiosRequestHeaders
          (config.headers as AxiosRequestHeaders)['Authorization'] = 'Bearer ' + newToken
          return api(config)
        } catch (refreshError) {
          auth.logout()
          return Promise.reject(refreshError)
        }
      }
    }

    return Promise.reject(error)
  }
)

export default api