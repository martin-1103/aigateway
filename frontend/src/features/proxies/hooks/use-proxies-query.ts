import { useQuery } from '@tanstack/react-query'
import { getProxies } from '../api'

interface UseProxiesQueryParams {
  limit?: number
  offset?: number
}

export const PROXIES_QUERY_KEY = ['proxies']

export function useProxiesQuery(params?: UseProxiesQueryParams) {
  return useQuery({
    queryKey: [...PROXIES_QUERY_KEY, params],
    queryFn: () => getProxies(params),
  })
}
