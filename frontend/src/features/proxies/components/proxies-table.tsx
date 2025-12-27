import { Server } from 'lucide-react'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { LoadingSpinner } from '@/components/feedback'
import { EmptyState } from '@/components/feedback'
import type { Proxy } from '../proxies.types'
import {
  getStatusBadge,
  getActiveBadge,
  getProtocolBadge,
  renderActions,
} from './proxies-table-columns'

interface ProxiesTableProps {
  proxies: Proxy[]
  isLoading: boolean
  onEdit: (proxy: Proxy) => void
  onDelete: (proxy: Proxy) => void
}

export function ProxiesTable({
  proxies,
  isLoading,
  onEdit,
  onDelete,
}: ProxiesTableProps) {
  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <LoadingSpinner size="lg" />
      </div>
    )
  }

  if (proxies.length === 0) {
    return (
      <EmptyState
        icon={Server}
        title="No proxies found"
        description="Create your first proxy to get started."
      />
    )
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>URL</TableHead>
          <TableHead>Protocol</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>Health</TableHead>
          <TableHead>Accounts</TableHead>
          <TableHead>Success Rate</TableHead>
          <TableHead>Latency</TableHead>
          <TableHead className="w-[50px]">
            <span className="sr-only">Actions</span>
          </TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {proxies.map((proxy) => (
          <TableRow key={proxy.id}>
            <TableCell className="font-mono text-sm">{proxy.url}</TableCell>
            <TableCell>{getProtocolBadge(proxy.protocol)}</TableCell>
            <TableCell>{getActiveBadge(proxy.is_active)}</TableCell>
            <TableCell>{getStatusBadge(proxy.health_status)}</TableCell>
            <TableCell>
              {proxy.current_accounts} / {proxy.max_accounts}
            </TableCell>
            <TableCell>{proxy.success_rate.toFixed(1)}%</TableCell>
            <TableCell>{proxy.avg_latency_ms}ms</TableCell>
            <TableCell>{renderActions(proxy, { onEdit, onDelete })}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
