import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { AreaChart } from '@tremor/react'

interface AreaChartCardProps {
  title: string
  data: Record<string, unknown>[]
  index: string
  categories: string[]
  colors?: string[]
}

export function AreaChartCard({ title, data, index, categories, colors }: AreaChartCardProps) {
  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <AreaChart
          data={data}
          index={index}
          categories={categories}
          colors={colors ?? ['blue', 'emerald']}
          showLegend
          showGridLines={false}
          className="h-72"
        />
      </CardContent>
    </Card>
  )
}
