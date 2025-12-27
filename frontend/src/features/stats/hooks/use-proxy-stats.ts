import { useQuery } from '@tanstack/react-query'
import { getProxyStats } from '../api'

export function useProxyStats(proxyId: string) {
  return useQuery({
    queryKey: ['proxy-stats', proxyId],
    queryFn: () => getProxyStats(proxyId),
    enabled: !!proxyId,
  })
}
