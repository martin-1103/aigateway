import { PageHeader, PageContent } from '@/components/page'
import { StatsSummaryCards, RequestsTrendChart, RequestLogsTable } from './components'
import { useProxyStats, useRequestLogs } from './hooks'
import type { TrendDataPoint } from './stats.types'

const DEFAULT_PROXY_ID = 'default'

function generateMockTrendData(): TrendDataPoint[] {
  const data: TrendDataPoint[] = []
  const today = new Date()

  for (let i = 6; i >= 0; i--) {
    const date = new Date(today)
    date.setDate(date.getDate() - i)
    data.push({
      date: date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
      requests: Math.floor(Math.random() * 1000) + 100,
      errors: Math.floor(Math.random() * 50),
    })
  }

  return data
}

export function StatsPage() {
  const { data: stats, isLoading: statsLoading } = useProxyStats(DEFAULT_PROXY_ID)
  const { data: logsData, isLoading: logsLoading } = useRequestLogs({ limit: 10 })

  const trendData = generateMockTrendData()

  return (
    <>
      <PageHeader
        title="Statistics"
        description="Monitor API usage and performance metrics"
      />
      <PageContent>
        <div className="space-y-6">
          <StatsSummaryCards stats={stats} isLoading={statsLoading} />
          <RequestsTrendChart data={trendData} />
          <RequestLogsTable logs={logsData?.logs ?? []} isLoading={logsLoading} />
        </div>
      </PageContent>
    </>
  )
}
