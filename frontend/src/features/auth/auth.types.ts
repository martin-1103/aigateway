export type Role = 'admin' | 'user' | 'provider'

export interface User {
  id: string
  username: string
  role: Role
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  token: string
  user: User
}

export interface ChangePasswordRequest {
  current_password: string
  new_password: string
}
