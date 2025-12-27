import { apiClient } from '@/lib/api-client'
import type { CreateApiKeyRequest, CreateApiKeyResponse } from '../api-keys.types'

export async function createApiKey(
  data: CreateApiKeyRequest
): Promise<CreateApiKeyResponse> {
  const response = await apiClient.post<CreateApiKeyResponse>(
    '/api/v1/api-keys',
    data
  )
  return response.data
}
