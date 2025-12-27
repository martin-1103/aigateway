export interface OAuthProvider {
  id: string
  name: string
  auth_url: string
  scopes: string[]
}

export interface OAuthProvidersResponse {
  providers: OAuthProvider[]
}

export interface InitOAuthRequest {
  provider: string
  project_id: string
  flow_type?: string
}

export interface InitOAuthResponse {
  auth_url: string
}

export interface RefreshTokenRequest {
  provider: string
  account_id: string
}

export interface RefreshTokenResponse {
  success: boolean
  message: string
}
