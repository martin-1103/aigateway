import { apiClient } from '@/lib/api-client'
import type { User } from '../auth.types'

export async function getMe(): Promise<User> {
  const response = await apiClient.get<User>('/api/v1/auth/me')
  return response.data
}
