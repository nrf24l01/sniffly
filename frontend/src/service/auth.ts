import api from '@/service/axios'

export interface LoginResponse {
  access_token: string
  error?: string
}

export const authService = {
  async login(username: string, password: string): Promise<LoginResponse> {
    try {
      const res = await api.post('/auth/login', { username, password })
      // Successful response (2xx)
      return { access_token: res.data.token }
    } catch (err: any) {
      // If server returned structured field errors, return them as a readable string
      const resp = err?.response
      const data = resp?.data
      if (data && data.errors && typeof data.errors === 'object') {
        const parts: string[] = []
        for (const key in data.errors) {
          const v = data.errors[key]
          if (Array.isArray(v)) parts.push(`${key}: ${v.join(', ')}`)
          else parts.push(`${key}: ${v}`)
        }
        return { access_token: '', error: parts.join('; ') }
      }

      // Fallback to original status-code based messages when no structured errors
      const status = resp?.status
      switch (status) {
        case 400:
          return { access_token: '', error: 'Invalid request. Please check your input.' }
        case 422:
          return { access_token: '', error: 'Invalid request. Please check your input.' }
        case 401:
          return { access_token: '', error: 'Invalid username or password.' }
        case 500:
          return { access_token: '', error: 'Server error. Please try again later.' }
        case 502:
          return { access_token: '', error: 'Server error. Please try again later.' }
        default:
          // If no HTTP response, surface axios error message or generic fallback
          return { access_token: '', error: err?.message ?? 'Login failed. Please try again.' }
      }
    }
  }
}
