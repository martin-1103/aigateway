import { apiClient } from '@/lib/api-client'
import type { OAuthProvidersResponse } from '../oauth.types'

export async function getOAuthProviders(): Promise<OAuthProvidersResponse> {
  const response = await apiClient.get<OAuthProvidersResponse>(
    '/api/v1/oauth/providers'
  )
  return response.data
}
