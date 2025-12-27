export interface Account {
  id: string
  provider: string
  email: string
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface AccountsResponse {
  data: Account[]
  total: number
}

export interface CreateAccountRequest {
  provider: string
  email: string
  credentials: string
  is_active?: boolean
}

export interface UpdateAccountRequest {
  provider?: string
  email?: string
  credentials?: string
  is_active?: boolean
}
