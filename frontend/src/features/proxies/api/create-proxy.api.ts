import { apiClient } from '@/lib/api-client'
import type { Proxy, CreateProxyRequest } from '../proxies.types'

export async function createProxy(data: CreateProxyRequest): Promise<Proxy> {
  const response = await apiClient.post<Proxy>('/api/v1/proxies', data)
  return response.data
}
