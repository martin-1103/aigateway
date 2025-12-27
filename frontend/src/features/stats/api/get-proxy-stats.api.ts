import { apiClient } from '@/lib/api-client'
import type { ProxyStats } from '../stats.types'

export async function getProxyStats(proxyId: string): Promise<ProxyStats> {
  const response = await apiClient.get<ProxyStats>(`/api/v1/stats/proxies/${proxyId}`)
  return response.data
}
