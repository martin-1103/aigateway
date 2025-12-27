import { apiClient } from '@/lib/api-client'
import type { Proxy, UpdateProxyRequest } from '../proxies.types'

export async function updateProxy(id: number, data: UpdateProxyRequest): Promise<Proxy> {
  const response = await apiClient.put<Proxy>(`/api/v1/proxies/${id}`, data)
  return response.data
}
