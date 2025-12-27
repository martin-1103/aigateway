import { useQuery } from '@tanstack/react-query'
import { getRequestLogs } from '../api'

interface UseRequestLogsOptions {
  page?: number
  limit?: number
}

export function useRequestLogs(options?: UseRequestLogsOptions) {
  return useQuery({
    queryKey: ['request-logs', options?.page, options?.limit],
    queryFn: () => getRequestLogs(options),
  })
}
