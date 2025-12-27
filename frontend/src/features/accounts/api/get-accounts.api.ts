import { apiClient } from '@/lib/api-client'
import type { AccountsResponse } from '../accounts.types'

export async function getAccounts(): Promise<AccountsResponse> {
  const response = await apiClient.get<AccountsResponse>('/api/v1/accounts')
  return response.data
}
