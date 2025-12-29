import { liteApi } from './lite.client'
import type { Account } from '@/features/accounts/accounts.types'

export interface LiteAccountsResponse {
  data: Account[]
  total: number
}

export async function getLiteAccounts(): Promise<LiteAccountsResponse> {
  const response = await liteApi.get<LiteAccountsResponse>('/accounts')
  return response.data
}

export interface OAuthProvider {
  id: string
  name: string
}

export interface OAuthProvidersResponse {
  providers: OAuthProvider[]
}

export async function getLiteOAuthProviders(): Promise<OAuthProvider[]> {
  const response = await liteApi.get<OAuthProvidersResponse>('/oauth/providers')
  return response.data.providers
}

export interface InitOAuthRequest {
  provider: string
  project_id: string
  flow_type: 'auto' | 'manual'
  redirect_uri?: string
}

export interface InitOAuthResponse {
  auth_url: string
  state: string
  flow_type: string
  expires_at: string
}

export async function initLiteOAuth(req: InitOAuthRequest): Promise<InitOAuthResponse> {
  const response = await liteApi.post<InitOAuthResponse>('/oauth/init', req)
  return response.data
}
