import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { refreshAccessToken } from '@/service/axios'

function isTokenExpired(token: string | null) {
  if (!token) return true
  try {
    const parts = token.split('.')
    if (parts.length < 2) return true
    const payload = JSON.parse(atob(parts[1]!))
    if (!payload.exp) return false
    return Date.now() >= payload.exp * 1000
  } catch (e) {
    return true
  }
}

export const useAuthStore = defineStore('auth', () => {
  const accessToken = ref<string | null>(null)
  const user_id = ref<string | null>(null)
  const username = ref<string | null>(null)

  const isAuthenticated = computed(() =>
    !!accessToken.value && !isTokenExpired(accessToken.value)
  )
  const authHeader = computed(() =>
    accessToken.value ? { Authorization: 'Bearer ' + accessToken.value } : {}
  )

  function setToken(token: string | null) {
    accessToken.value = token
    try {
      if (token) {
        localStorage.setItem('access_token', token)
      } else {
        localStorage.removeItem('access_token')
      }
    } catch (e) {
      // ignore storage errors
    }
    // Extract user_id and username from JWT
    if (token) {
      try {
        const parts = token.split('.')
        if (parts.length < 2) {
          user_id.value = null
          username.value = null
        } else {
          const payload = JSON.parse(atob(parts[1]!))
          user_id.value = payload.user_id || null
          username.value = payload.username || null
        }
      } catch (e) {
        user_id.value = null
        username.value = null
      }
    } else {
      user_id.value = null
      username.value = null
    }
  }

  function logout() {
    accessToken.value = null
    user_id.value = null
    username.value = null
    try {
      localStorage.removeItem('access_token')
    } catch (e) {
      // ignore
    }
  }

  // Add method to refresh token
  async function refreshToken() {
    try {
      const newToken = await refreshAccessToken()
      setToken(newToken)
      return true
    } catch (error) {
      logout()
      return false
    }
  }

  // Restore token from localStorage if present
  try {
    const stored = localStorage.getItem('access_token')
    if (stored) setToken(stored)
  } catch (e) {
    // ignore
  }

  return {
    accessToken,
    user_id,
    username,
    isAuthenticated,
    authHeader,
    setToken,
    logout,
    refreshToken
  }
})