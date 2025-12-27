import { MoreHorizontal, Pencil, Trash2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Badge } from '@/components/ui/badge'
import type { Proxy } from '../proxies.types'

interface ColumnActionsProps {
  proxy: Proxy
  onEdit: (proxy: Proxy) => void
  onDelete: (proxy: Proxy) => void
}

function ColumnActions({ proxy, onEdit, onDelete }: ColumnActionsProps) {
  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="sm" aria-label="Open actions menu">
          <MoreHorizontal className="h-4 w-4" />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end">
        <DropdownMenuItem onClick={() => onEdit(proxy)}>
          <Pencil className="mr-2 h-4 w-4" />
          Edit
        </DropdownMenuItem>
        <DropdownMenuItem
          onClick={() => onDelete(proxy)}
          className="text-destructive"
        >
          <Trash2 className="mr-2 h-4 w-4" />
          Delete
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}

export function getStatusBadge(status: string) {
  const variants: Record<string, 'success' | 'warning' | 'destructive'> = {
    healthy: 'success',
    degraded: 'warning',
    down: 'destructive',
  }
  return <Badge variant={variants[status] || 'secondary'}>{status}</Badge>
}

export function getActiveBadge(isActive: boolean) {
  return isActive ? (
    <Badge variant="success">Active</Badge>
  ) : (
    <Badge variant="secondary">Inactive</Badge>
  )
}

export function getProtocolBadge(protocol: string) {
  return <Badge variant="outline">{protocol.toUpperCase()}</Badge>
}

export interface ProxiesTableColumnsProps {
  onEdit: (proxy: Proxy) => void
  onDelete: (proxy: Proxy) => void
}

export function renderActions(
  proxy: Proxy,
  { onEdit, onDelete }: ProxiesTableColumnsProps
) {
  return <ColumnActions proxy={proxy} onEdit={onEdit} onDelete={onDelete} />
}
