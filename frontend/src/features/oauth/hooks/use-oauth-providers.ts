import { useQuery } from '@tanstack/react-query'
import { getOAuthProviders } from '../api'

export function useOAuthProviders() {
  return useQuery({
    queryKey: ['oauth-providers'],
    queryFn: getOAuthProviders,
  })
}
