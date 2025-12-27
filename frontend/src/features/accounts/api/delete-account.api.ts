import { apiClient } from '@/lib/api-client'

export async function deleteAccount(id: string): Promise<void> {
  await apiClient.delete(`/api/v1/accounts/${id}`)
}
