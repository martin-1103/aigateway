export interface ApiKey {
  id: string
  name: string
  key_prefix: string
  created_at: string
  last_used_at?: string
}

export interface CreateApiKeyRequest {
  name: string
}

export interface CreateApiKeyResponse {
  id: string
  name: string
  key: string
}
