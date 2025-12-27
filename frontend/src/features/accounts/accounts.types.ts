export interface Provider {
  id: string
  name: string
  base_url: string
  supported_auth_types: string[]
  auth_strategy: string
  supported_models: string
  is_active: boolean
  config: string
  created_at: string
  updated_at: string
}

export interface Account {
  id: string
  provider_id: string
  label: string
  auth_data: string
  metadata: string
  is_active: boolean
  proxy_url: string
  proxy_id?: number
  expires_at?: string
  last_used_at?: string
  usage_count: number
  created_at: string
  updated_at: string
  created_by?: string
  provider?: Provider
}

export interface AccountsResponse {
  data: Account[]
  total: number
  limit: number
  offset: number
}

export interface CreateAccountRequest {
  provider_id: string
  label: string
  auth_data: string
  is_active?: boolean
}

export interface UpdateAccountRequest {
  provider_id?: string
  label?: string
  auth_data?: string
  is_active?: boolean
}
