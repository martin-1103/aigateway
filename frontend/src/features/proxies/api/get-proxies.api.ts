import { apiClient } from '@/lib/api-client'
import type { ProxiesResponse } from '../proxies.types'

interface GetProxiesParams {
  limit?: number
  offset?: number
}

export async function getProxies(params?: GetProxiesParams): Promise<ProxiesResponse> {
  const response = await apiClient.get<ProxiesResponse>('/api/v1/proxies', { params })
  return response.data
}
