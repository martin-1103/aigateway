import { useState, useEffect } from 'react'
import { FolderOpen } from 'lucide-react'
import { getLiteAccounts, getLiteOAuthProviders, initLiteOAuth } from './api/lite-accounts.api'
import type { Account } from '@/features/accounts/accounts.types'
import type { OAuthProvider } from './api/lite-accounts.api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
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
  DialogTrigger,
} from '@/components/ui/dialog'
import { LoadingSpinner, EmptyState } from '@/components/feedback'
import { Badge } from '@/components/ui/badge'

export function LiteAccountsPage() {
  const [accounts, setAccounts] = useState<Account[]>([])
  const [providers, setProviders] = useState<OAuthProvider[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isAddOpen, setIsAddOpen] = useState(false)
  const [isInitiating, setIsInitiating] = useState(false)

  const fetchData = async () => {
    setIsLoading(true)
    try {
      const [accountsRes, providersRes] = await Promise.all([
        getLiteAccounts(),
        getLiteOAuthProviders(),
      ])
      setAccounts(accountsRes.data)
      setProviders(providersRes)
    } catch {
      // Error handled by axios interceptor
    } finally {
      setIsLoading(false)
    }
  }

  useEffect(() => {
    fetchData()
  }, [])

  const handleAddOAuth = async (providerId: string) => {
    setIsInitiating(true)
    try {
      const callbackUrl = `${import.meta.env.VITE_API_URL || 'http://localhost:8088'}/api/v1/lite/oauth/callback`

      const response = await initLiteOAuth({
        provider: providerId,
        project_id: 'default',
        flow_type: 'auto',
        redirect_uri: callbackUrl,
      })

      // Redirect to OAuth provider
      window.location.href = response.auth_url
    } catch {
      setIsInitiating(false)
    }
  }

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">My Accounts</h2>
          <p className="text-muted-foreground">
            Manage your OAuth accounts
          </p>
        </div>
        <Dialog open={isAddOpen} onOpenChange={setIsAddOpen}>
          <DialogTrigger asChild>
            <Button>Add OAuth Account</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Add OAuth Account</DialogTitle>
            </DialogHeader>
            <div className="space-y-4 py-4">
              <p className="text-sm text-muted-foreground">
                Select a provider to authenticate with:
              </p>
              <div className="grid gap-2">
                {providers.map((provider) => (
                  <Button
                    key={provider.id}
                    variant="outline"
                    className="justify-start"
                    onClick={() => handleAddOAuth(provider.id)}
                    disabled={isInitiating}
                  >
                    {isInitiating ? (
                      <LoadingSpinner className="mr-2 h-4 w-4" />
                    ) : null}
                    {provider.name}
                  </Button>
                ))}
              </div>
            </div>
          </DialogContent>
        </Dialog>
      </div>

      {accounts.length === 0 ? (
        <EmptyState
          icon={FolderOpen}
          title="No accounts yet"
          description="Add your first OAuth account to get started"
        />
      ) : (
        <Card>
          <CardHeader>
            <CardTitle>Accounts ({accounts.length})</CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Label</TableHead>
                  <TableHead>Provider</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead>Expires</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {accounts.map((account) => (
                  <TableRow key={account.id}>
                    <TableCell className="font-medium">
                      {account.label}
                    </TableCell>
                    <TableCell>{account.provider_id}</TableCell>
                    <TableCell>
                      <Badge variant={account.is_active ? 'default' : 'secondary'}>
                        {account.is_active ? 'Active' : 'Inactive'}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      {account.expires_at
                        ? new Date(account.expires_at).toLocaleDateString()
                        : '-'}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
