import { useMemo, useState, useRef } from 'react'
import {
  flexRender,
  getCoreRowModel,
  useReactTable,
  getPaginationRowModel,
  getSortedRowModel,
} from '@tanstack/react-table'
import type { SortingState } from '@tanstack/react-table'
import { Copy } from 'lucide-react'
import { toast } from 'sonner'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { getAccountColumns } from './accounts-table-columns'
import type { Account } from '../accounts.types'

interface AccountsTableProps {
  data: Account[]
  isLoading?: boolean
  onEdit: (account: Account) => void
  onDelete: (account: Account) => void
}

function generateCurl(accountId: string): string | null {
  const authStorage = localStorage.getItem('auth-storage')
  if (!authStorage) return null

  const { state } = JSON.parse(authStorage)
  const authHeader = state?.accessKey
    ? `X-Access-Key: ${state.accessKey}`
    : `Authorization: Bearer ${state.token}`

  const baseUrl = window.location.origin.replace(':5173', ':8088')
  const body = JSON.stringify({
    model: 'antigravity:claude-sonnet',
    messages: [{ role: 'user', content: 'Hello' }],
  }).replace(/"/g, '\\"')

  return `curl.exe -X POST "${baseUrl}/v1/chat/completions?account_id=${accountId}" -H "${authHeader}" -H "Content-Type: application/json" -d "${body}"`
}

export function AccountsTable({ data, isLoading, onEdit, onDelete }: AccountsTableProps) {
  const [sorting, setSorting] = useState<SortingState>([])
  const [curlCommand, setCurlCommand] = useState<string | null>(null)
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  const handleShowCurl = (account: Account) => {
    const curl = generateCurl(account.id)
    if (!curl) {
      toast.error('Not authenticated')
      return
    }
    setCurlCommand(curl)
  }

  const handleCopy = () => {
    if (textareaRef.current) {
      textareaRef.current.select()
      document.execCommand('copy')
      toast.success('Copied to clipboard')
    }
  }

  const columns = useMemo(
    () => getAccountColumns({ onEdit, onDelete, onShowCurl: handleShowCurl }),
    [onEdit, onDelete]
  )

  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    onSortingChange: setSorting,
    state: { sorting },
  })

  if (isLoading) {
    return (
      <div className="flex h-32 items-center justify-center">
        <p className="text-muted-foreground">Loading accounts...</p>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <TableHead key={header.id}>
                    {header.isPlaceholder
                      ? null
                      : flexRender(header.column.columnDef.header, header.getContext())}
                  </TableHead>
                ))}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow key={row.id}>
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={columns.length} className="h-24 text-center">
                  No accounts found.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      <div className="flex items-center justify-end gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={() => table.previousPage()}
          disabled={!table.getCanPreviousPage()}
        >
          Previous
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={() => table.nextPage()}
          disabled={!table.getCanNextPage()}
        >
          Next
        </Button>
      </div>

      <Dialog open={!!curlCommand} onOpenChange={() => setCurlCommand(null)}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Curl Command</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <textarea
              ref={textareaRef}
              readOnly
              value={curlCommand || ''}
              className="w-full h-32 p-3 font-mono text-sm bg-muted rounded-md border resize-none"
            />
            <Button onClick={handleCopy} className="w-full">
              <Copy className="mr-2 h-4 w-4" />
              Copy to Clipboard
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  )
}
