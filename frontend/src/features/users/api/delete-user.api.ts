import { apiClient } from '@/lib/api-client'

export async function deleteUser(id: string): Promise<void> {
  await apiClient.delete(`/api/v1/users/${id}`)
}
