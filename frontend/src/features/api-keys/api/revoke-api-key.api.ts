import { apiClient } from '@/lib/api-client'

export async function revokeApiKey(id: string): Promise<void> {
  await apiClient.delete(`/api/v1/api-keys/${id}`)
}
