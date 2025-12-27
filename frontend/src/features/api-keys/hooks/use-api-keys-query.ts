import { useQuery } from '@tanstack/react-query'
import { getApiKeys } from '../api'

export const API_KEYS_QUERY_KEY = ['api-keys']

export function useApiKeysQuery() {
  return useQuery({
    queryKey: API_KEYS_QUERY_KEY,
    queryFn: getApiKeys,
  })
}
