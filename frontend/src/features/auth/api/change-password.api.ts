import { apiClient } from '@/lib/api-client'
import type { ChangePasswordRequest } from '../auth.types'

export async function changePassword(data: ChangePasswordRequest): Promise<void> {
  await apiClient.put('/api/v1/auth/password', data)
}
