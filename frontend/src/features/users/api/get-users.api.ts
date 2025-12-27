import { apiClient } from '@/lib/api-client'
import type { UsersResponse } from '../users.types'

export async function getUsers(): Promise<UsersResponse> {
  const response = await apiClient.get<UsersResponse>('/api/v1/users')
  return response.data
}
