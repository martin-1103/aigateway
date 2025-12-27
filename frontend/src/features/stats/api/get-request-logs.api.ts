import { apiClient } from '@/lib/api-client'
import type { RequestLogsResponse } from '../stats.types'

interface GetRequestLogsParams {
  page?: number
  limit?: number
}

export async function getRequestLogs(params?: GetRequestLogsParams): Promise<RequestLogsResponse> {
  const response = await apiClient.get<RequestLogsResponse>('/api/v1/stats/logs', { params })
  return response.data
}
