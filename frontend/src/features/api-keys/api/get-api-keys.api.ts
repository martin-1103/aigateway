import { apiClient } from '@/lib/api-client'
import type { ApiKey } from '../api-keys.types'

export async function getApiKeys(): Promise<ApiKey[]> {
  const response = await apiClient.get<ApiKey[]>('/api/v1/api-keys')
  return response.data
}
