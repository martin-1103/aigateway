import { apiClient } from '@/lib/api-client'
import type { Account, UpdateAccountRequest } from '../accounts.types'

export async function updateAccount(
  id: string,
  data: UpdateAccountRequest
): Promise<Account> {
  const response = await apiClient.put<Account>(`/api/v1/accounts/${id}`, data)
  return response.data
}
