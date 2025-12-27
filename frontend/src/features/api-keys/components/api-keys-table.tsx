import { useState } from 'react'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Trash2 } from 'lucide-react'
import { useApiKeysQuery } from '../hooks'
import { apiKeysTableColumns } from './api-keys-table-columns'
import { ApiKeyRevokeDialog } from './api-key-revoke-dialog'
import type { ApiKey } from '../api-keys.types'

function formatDate(dateString: string | undefined): string {
  if (!dateString) return 'Never'
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function ApiKeysTable() {
  const { data: apiKeys, isLoading, error } = useApiKeysQuery()
  const [revokeKey, setRevokeKey] = useState<ApiKey | null>(null)

  if (isLoading) {
    return <div className="text-center py-8 text-muted-foreground">Loading...</div>
  }

  if (error) {
    return (
      <div className="text-center py-8 text-destructive">
        Failed to load API keys
      </div>
    )
  }

  if (!apiKeys?.length) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        No API keys found. Create one to get started.
      </div>
    )
  }

  return (
    <>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              {apiKeysTableColumns.map((column) => (
                <TableHead key={column.key} className={column.className}>
                  {column.label}
                </TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {apiKeys.map((apiKey) => (
              <TableRow key={apiKey.id}>
                <TableCell className="font-medium">{apiKey.name}</TableCell>
                <TableCell>
                  <code className="px-2 py-1 bg-muted rounded text-sm">
                    {apiKey.key_prefix}...
                  </code>
                </TableCell>
                <TableCell>{formatDate(apiKey.created_at)}</TableCell>
                <TableCell>{formatDate(apiKey.last_used_at)}</TableCell>
                <TableCell>
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={() => setRevokeKey(apiKey)}
                    aria-label={`Revoke API key ${apiKey.name}`}
                  >
                    <Trash2 className="h-4 w-4 text-destructive" />
                  </Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
      <ApiKeyRevokeDialog
        apiKey={revokeKey}
        onClose={() => setRevokeKey(null)}
      />
    </>
  )
}
