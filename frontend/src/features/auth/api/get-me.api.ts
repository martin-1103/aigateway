import { apiClient } from '@/lib/api-client'
import type { MeResponse } from '../auth.types'

export async function getMe(): Promise<MeResponse> {
  const response = await apiClient.get<MeResponse>('/api/v1/auth/me')
  return response.data
}
