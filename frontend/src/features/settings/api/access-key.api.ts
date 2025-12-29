import { apiClient } from '@/lib/api-client'

interface GetMyKeyResponse {
  key: string
}

interface RegenerateKeyResponse {
  key: string
  message: string
}

export async function getMyAccessKey(): Promise<string> {
  const response = await apiClient.get<GetMyKeyResponse>('/api/v1/auth/my-key')
  return response.data.key
}

export async function regenerateAccessKey(): Promise<string> {
  const response = await apiClient.post<RegenerateKeyResponse>('/api/v1/auth/regenerate-key')
  return response.data.key
}
