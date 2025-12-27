import { apiClient } from '@/lib/api-client'
import type { User, CreateUserRequest } from '../users.types'

export async function createUser(data: CreateUserRequest): Promise<User> {
  const response = await apiClient.post<User>('/api/v1/users', data)
  return response.data
}
