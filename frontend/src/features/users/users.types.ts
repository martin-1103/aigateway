export type Role = 'admin' | 'user' | 'provider'

export interface User {
  id: string
  username: string
  role: Role
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface UsersResponse {
  data: User[]
  total: number
}

export interface CreateUserRequest {
  username: string
  password: string
  role: Role
  is_active: boolean
}

export interface UpdateUserRequest {
  username?: string
  password?: string
  role?: Role
  is_active?: boolean
}
