import { apiClient } from '@/lib/api-client'
import type { RefreshTokenRequest, RefreshTokenResponse } from '../oauth.types'

export async function refreshToken(
  data: RefreshTokenRequest
): Promise<RefreshTokenResponse> {
  const response = await apiClient.post<RefreshTokenResponse>(
    '/api/v1/oauth/refresh',
    data
  )
  return response.data
}
