import { AreaChartCard } from '@/components/charts'
import type { TrendDataPoint } from '../stats.types'

interface RequestsTrendChartProps {
  data: TrendDataPoint[]
}

export function RequestsTrendChart({ data }: RequestsTrendChartProps) {
  return (
    <AreaChartCard
      title="Request Trends (7 Days)"
      data={data}
      index="date"
      categories={['requests', 'errors']}
      colors={['blue', 'red']}
    />
  )
}
