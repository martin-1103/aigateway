import { apiClient } from '@/lib/api-client'
import type { InitOAuthRequest, InitOAuthResponse } from '../oauth.types'

export async function initOAuthFlow(
  data: InitOAuthRequest
): Promise<InitOAuthResponse> {
  const response = await apiClient.post<InitOAuthResponse>(
    '/api/v1/oauth/init',
    data
  )
  return response.data
}
