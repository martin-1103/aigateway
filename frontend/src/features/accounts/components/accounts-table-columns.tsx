import type { ColumnDef } from '@tanstack/react-table'
import { Copy, MoreHorizontal, Pencil, Terminal, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import type { Account } from '../accounts.types'

interface ColumnsProps {
  onEdit: (account: Account) => void
  onDelete: (account: Account) => void
}

export function getAccountColumns({ onEdit, onDelete }: ColumnsProps): ColumnDef<Account>[] {
  return [
    {
      accessorKey: 'id',
      header: 'ID',
      cell: ({ row }) => {
        const id = row.original.id
        const shortId = id.length > 8 ? `${id.slice(0, 8)}...` : id

        const handleCopy = () => {
          navigator.clipboard.writeText(id)
          toast.success('ID copied to clipboard')
        }

        return (
          <div className="flex items-center gap-1">
            <code className="text-xs bg-muted px-1.5 py-0.5 rounded">{shortId}</code>
            <Button
              variant="ghost"
              size="sm"
              className="h-6 w-6 p-0"
              onClick={handleCopy}
              title="Copy full ID"
            >
              <Copy className="h-3 w-3" />
            </Button>
          </div>
        )
      },
    },
    {
      accessorKey: 'label',
      header: 'Label',
    },
    {
      accessorKey: 'provider_id',
      header: 'Provider',
      cell: ({ row }) => {
        const account = row.original
        return <span>{account.provider?.name || account.provider_id}</span>
      },
    },
    {
      accessorKey: 'is_active',
      header: 'Status',
      cell: ({ row }) => (
        <span
          className={`inline-flex items-center rounded-full px-2 py-1 text-xs font-medium ${
            row.getValue('is_active')
              ? 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300'
              : 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300'
          }`}
        >
          {row.getValue('is_active') ? 'Active' : 'Inactive'}
        </span>
      ),
    },
    {
      accessorKey: 'created_at',
      header: 'Created',
      cell: ({ row }) => new Date(row.getValue('created_at')).toLocaleDateString(),
    },
    {
      id: 'actions',
      header: () => <span className="sr-only">Actions</span>,
      cell: ({ row }) => {
        const account = row.original

        return (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button
                variant="ghost"
                size="sm"
                className="h-8 w-8 p-0"
                aria-label="Open actions menu"
              >
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem
                onClick={() => {
                  const authStorage = localStorage.getItem('auth-storage')
                  if (!authStorage) {
                    toast.error('Not authenticated')
                    return
                  }
                  const { state } = JSON.parse(authStorage)
                  const authHeader = state?.accessKey
                    ? `-H "X-Access-Key: ${state.accessKey}"`
                    : `-H "Authorization: Bearer ${state.token}"`

                  const baseUrl = window.location.origin.replace(':5173', ':8088')
                  const curl = `curl -X POST "${baseUrl}/v1/chat/completions?account_id=${account.id}" \\
  ${authHeader} \\
  -H "Content-Type: application/json" \\
  -d '{"model": "antigravity:claude-sonnet", "messages": [{"role": "user", "content": "Hello"}]}'`
                  navigator.clipboard.writeText(curl)
                  toast.success('Curl command copied to clipboard')
                }}
              >
                <Terminal className="mr-2 h-4 w-4" />
                Copy Curl
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => onEdit(account)}>
                <Pencil className="mr-2 h-4 w-4" />
                Edit
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={() => onDelete(account)}
                className="text-destructive focus:text-destructive"
              >
                <Trash2 className="mr-2 h-4 w-4" />
                Delete
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        )
      },
    },
  ]
}
