import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import type { RequestLog } from '../stats.types'

interface RequestLogsTableProps {
  logs: RequestLog[]
  isLoading: boolean
}

export function RequestLogsTable({ logs, isLoading }: RequestLogsTableProps) {
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString()
  }

  const getStatusBadgeClass = (statusCode: number, error: string) => {
    const isSuccess = statusCode >= 200 && statusCode < 300 && !error
    return isSuccess
      ? 'inline-flex items-center rounded-full bg-green-100 px-2 py-1 text-xs font-medium text-green-700'
      : 'inline-flex items-center rounded-full bg-red-100 px-2 py-1 text-xs font-medium text-red-700'
  }

  const getStatusLabel = (statusCode: number, error: string) => {
    if (error) return `Error: ${error}`
    if (statusCode >= 200 && statusCode < 300) return 'Success'
    return `Error (${statusCode})`
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Recent Request Logs</CardTitle>
      </CardHeader>
      <CardContent>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Account</TableHead>
              <TableHead>Model</TableHead>
              <TableHead>Provider</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Latency</TableHead>
              <TableHead>Time</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {isLoading ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center text-muted-foreground">
                  Loading...
                </TableCell>
              </TableRow>
            ) : logs.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center text-muted-foreground">
                  No request logs found
                </TableCell>
              </TableRow>
            ) : (
              logs.map((log) => (
                <TableRow key={log.id}>
                  <TableCell className="font-mono text-xs">{log.account_id ? log.account_id.substring(0, 8) : '-'}</TableCell>
                  <TableCell className="font-medium">{log.model || '-'}</TableCell>
                  <TableCell>{log.provider_id || '-'}</TableCell>
                  <TableCell>
                    <span className={getStatusBadgeClass(log.status_code, log.error)}>
                      {getStatusLabel(log.status_code, log.error)}
                    </span>
                  </TableCell>
                  <TableCell>{log.latency_ms}ms</TableCell>
                  <TableCell className="text-muted-foreground">
                    {formatDate(log.created_at)}
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </CardContent>
    </Card>
  )
}
