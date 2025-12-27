import { useQuery } from '@tanstack/react-query'
import { getAccounts } from '../api'

export const ACCOUNTS_QUERY_KEY = ['accounts']

export function useAccountsQuery() {
  return useQuery({
    queryKey: ACCOUNTS_QUERY_KEY,
    queryFn: getAccounts,
  })
}
