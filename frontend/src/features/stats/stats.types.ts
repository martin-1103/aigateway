export interface ProxyStats {
  proxy_id: string
  requests_today: number
  errors_today: number
  requests_total: number
  errors_total: number
}

export interface RequestLog {
  id: string
  model: string
  provider: string
  status: 'success' | 'error'
  latency_ms: number
  created_at: string
}

export interface RequestLogsResponse {
  logs: RequestLog[]
  total: number
}

export interface TrendDataPoint {
  date: string
  requests: number
  errors: number
}
