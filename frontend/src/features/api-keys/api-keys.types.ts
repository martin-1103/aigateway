export interface User {
  id: string
  username: string
  role: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface ApiKey {
  id: string
  user_id: string
  key_prefix: string
  label: string
  is_active: boolean
  last_used_at: string | null
  created_at: string
  updated_at: string
  user: User
}

export interface CreateApiKeyRequest {
  label: string
}

export interface CreateApiKeyResponse {
  id: string
  label: string
  key: string
}

export interface ApiKeysResponse {
  data: ApiKey[]
  total: number
}
