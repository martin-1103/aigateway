import axios from 'axios'

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8088'

export const liteApi = axios.create({
  baseURL: `${API_URL}/api/v1/lite`,
})

liteApi.interceptors.request.use((config) => {
  const params = new URLSearchParams(window.location.search)
  const key = params.get('key')
  if (key) {
    config.headers['X-Access-Key'] = key
  }
  return config
})

export function getAccessKey(): string | null {
  const params = new URLSearchParams(window.location.search)
  return params.get('key')
}
