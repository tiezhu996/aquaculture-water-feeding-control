import axios from 'axios'

export interface ApiEnvelope<T> {
  data: T
}

export interface ApiErrorBody {
  error?: { code?: string; message?: string; requestId?: string }
}

export const client = axios.create({
  baseURL: '/api',
  timeout: 15_000,
  headers: { 'Content-Type': 'application/json' },
})

client.interceptors.request.use((config) => {
  const token = localStorage.getItem('aquaculture_token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

client.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401 && !String(error.config?.url).includes('/auth/login')) {
      localStorage.removeItem('aquaculture_token')
      localStorage.removeItem('aquaculture_user')
      window.dispatchEvent(new CustomEvent('auth-expired'))
    }
    return Promise.reject(error)
  },
)
