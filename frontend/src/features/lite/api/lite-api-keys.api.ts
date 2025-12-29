import { liteApi } from './lite.client'

export interface APIKey {
  id: string
  user_id: string
  key_prefix: string
  label: string
  is_active: boolean
  last_used_at?: string
  created_at: string
  updated_at: string
}

export interface APIKeysResponse {
  data: APIKey[]
}

export async function getLiteAPIKeys(): Promise<APIKey[]> {
  const response = await liteApi.get<APIKeysResponse>('/api-keys')
  return response.data.data
}
