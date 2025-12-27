import { Activity, AlertTriangle, TrendingUp, Clock } from 'lucide-react'
import { StatCard } from '@/components/charts'
import type { ProxyStats } from '../stats.types'

interface StatsSummaryCardsProps {
  stats: ProxyStats | undefined
  isLoading: boolean
}

export function StatsSummaryCards({ stats, isLoading }: StatsSummaryCardsProps) {
  const formatNumber = (value: number | undefined) => {
    if (value === undefined) return '--'
    return value.toLocaleString()
  }

  const calculateErrorRate = () => {
    if (!stats || stats.requests_total === 0) return '0%'
    const rate = (stats.errors_total / stats.requests_total) * 100
    return `${rate.toFixed(1)}%`
  }

  if (isLoading) {
    return (
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <StatCard title="Requests Today" value="--" icon={Activity} />
        <StatCard title="Errors Today" value="--" icon={AlertTriangle} />
        <StatCard title="Total Requests" value="--" icon={TrendingUp} />
        <StatCard title="Error Rate" value="--" icon={Clock} />
      </div>
    )
  }

  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
      <StatCard
        title="Requests Today"
        value={formatNumber(stats?.requests_today)}
        icon={Activity}
        description="API requests processed today"
      />
      <StatCard
        title="Errors Today"
        value={formatNumber(stats?.errors_today)}
        icon={AlertTriangle}
        description="Failed requests today"
      />
      <StatCard
        title="Total Requests"
        value={formatNumber(stats?.requests_total)}
        icon={TrendingUp}
        description="All-time request count"
      />
      <StatCard
        title="Error Rate"
        value={calculateErrorRate()}
        icon={Clock}
        description="Overall error percentage"
      />
    </div>
  )
}
