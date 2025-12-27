import { apiClient } from '@/lib/api-client'
import type { ApiKeysResponse } from '../api-keys.types'

export async function getApiKeys(): Promise<ApiKeysResponse> {
  const response = await apiClient.get<ApiKeysResponse>('/api/v1/api-keys')
  return response.data
}
