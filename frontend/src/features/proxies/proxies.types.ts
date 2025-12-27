export interface Proxy {
  id: number
  url: string
  protocol: 'http' | 'https' | 'socks5'
  is_active: boolean
  health_status: 'healthy' | 'degraded' | 'down'
  max_accounts: number
  current_accounts: number
  last_used_at: string | null
  usage_count: number
  priority: number
  weight: number
  consecutive_failures: number
  max_failures: number
  success_rate: number
  avg_latency_ms: number
  last_checked_at: string | null
  created_at: string
  updated_at: string
}

export interface ProxiesResponse {
  data: Proxy[]
  total: number
  limit: number
  offset: number
}

export interface CreateProxyRequest {
  url: string
  protocol: string
  is_active?: boolean
  max_accounts?: number
  priority?: number
  weight?: number
  max_failures?: number
}

export interface UpdateProxyRequest {
  url?: string
  protocol?: string
  is_active?: boolean
  max_accounts?: number
  priority?: number
  weight?: number
  max_failures?: number
}
