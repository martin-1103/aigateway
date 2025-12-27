import axios from 'axios'
import { logger } from './logger'

const DEBUG_ENABLED = import.meta.env.VITE_DEBUG === 'true'

export const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

apiClient.interceptors.request.use((config) => {
  const authStorage = localStorage.getItem('auth-storage')
  if (authStorage) {
    const { state } = JSON.parse(authStorage)
    if (state?.token) {
      config.headers.Authorization = `Bearer ${state.token}`
    }
  }

  if (DEBUG_ENABLED) {
    logger.api.log('Request', {
      method: config.method?.toUpperCase(),
      url: config.url,
      data: config.data,
      headers: { Authorization: config.headers.Authorization ? '***' : 'none' },
    })
  }

  return config
})

apiClient.interceptors.response.use(
  (response) => {
    if (DEBUG_ENABLED) {
      logger.api.info('Response', {
        status: response.status,
        url: response.config.url,
        data: response.data,
      })
    }
    return response
  },
  (error) => {
    if (DEBUG_ENABLED) {
      logger.api.error('Response Error', {
        status: error.response?.status,
        url: error.response?.config?.url,
        error: error.response?.data,
      })
    }

    if (error.response?.status === 401) {
      localStorage.removeItem('auth-storage')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)
