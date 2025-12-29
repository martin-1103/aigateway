import { apiClient } from '@/lib/api-client'
import type { ExchangeOAuthRequest, ExchangeOAuthResponse } from '../oauth.types'

export async function exchangeOAuth(
  data: ExchangeOAuthRequest
): Promise<ExchangeOAuthResponse> {
  const response = await apiClient.post<ExchangeOAuthResponse>(
    '/api/v1/oauth/exchange',
    data
  )
  return response.data
}
