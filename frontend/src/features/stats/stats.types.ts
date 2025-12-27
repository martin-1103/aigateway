export interface ProxyStats {
  proxy_id: string
  requests_today: number
  errors_today: number
  requests_total: number
  errors_total: number
}

export interface RequestLog {
  id: number
  provider_id: string | null
  account_id: string
  proxy_id: number | null
  model: string
  status_code: number
  latency_ms: number
  retry_count: number
  error: string
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
  [key: string]: string | number
}
