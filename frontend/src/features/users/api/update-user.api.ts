import { apiClient } from '@/lib/api-client'
import type { User, UpdateUserRequest } from '../users.types'

export async function updateUser(id: string, data: UpdateUserRequest): Promise<User> {
  const response = await apiClient.put<User>(`/api/v1/users/${id}`, data)
  return response.data
}
