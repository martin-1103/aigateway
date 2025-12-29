import { liteApi } from './lite.client'

export interface LiteUser {
  id: string
  username: string
  role: 'admin' | 'user' | 'provider'
}

export async function getLiteMe(): Promise<LiteUser> {
  const response = await liteApi.get<LiteUser>('/me')
  return response.data
}
