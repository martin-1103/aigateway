import { apiClient } from '@/lib/api-client'
import type { Account, CreateAccountRequest } from '../accounts.types'

export async function createAccount(data: CreateAccountRequest): Promise<Account> {
  const response = await apiClient.post<Account>('/api/v1/accounts', data)
  return response.data
}
