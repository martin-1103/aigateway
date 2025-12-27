import { apiClient } from '@/lib/api-client'
import type { LoginRequest, LoginResponse } from '../auth.types'

export async function login(data: LoginRequest): Promise<LoginResponse> {
  const response = await apiClient.post<LoginResponse>('/api/v1/auth/login', data)
  return response.data
}
