import { useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useInitOAuthFlow } from '../hooks'
import type { OAuthProvider } from '../oauth.types'

interface OAuthInitDialogProps {
  provider: OAuthProvider | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

export function OAuthInitDialog({
  provider,
  open,
  onOpenChange,
}: OAuthInitDialogProps) {
  const [accountId, setAccountId] = useState('')
  const { mutate, isPending } = useInitOAuthFlow()

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!provider || !accountId.trim()) return

    mutate({
      provider: provider.name,
      account_id: accountId.trim(),
    })
  }

  const handleClose = () => {
    setAccountId('')
    onOpenChange(false)
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Initialize OAuth Flow</DialogTitle>
          <DialogDescription>
            {provider
              ? `Start OAuth authentication for ${provider.name}. Enter the account ID to associate with this OAuth token.`
              : 'Select a provider to continue.'}
          </DialogDescription>
        </DialogHeader>

        {provider && (
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="provider-name">Provider</Label>
              <Input
                id="provider-name"
                value={provider.name}
                disabled
                aria-readonly="true"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="account-id">Account ID</Label>
              <Input
                id="account-id"
                value={accountId}
                onChange={(e) => setAccountId(e.target.value)}
                placeholder="Enter account ID"
                required
                aria-required="true"
              />
            </div>

            {provider.scopes.length > 0 && (
              <div className="space-y-2">
                <Label>Requested Scopes</Label>
                <div className="flex flex-wrap gap-2">
                  {provider.scopes.map((scope) => (
                    <span
                      key={scope}
                      className="rounded-md bg-secondary px-2 py-1 text-xs"
                    >
                      {scope}
                    </span>
                  ))}
                </div>
              </div>
            )}

            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={handleClose}
                disabled={isPending}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={isPending || !accountId.trim()}>
                {isPending ? 'Redirecting...' : 'Start OAuth Flow'}
              </Button>
            </DialogFooter>
          </form>
        )}
      </DialogContent>
    </Dialog>
  )
}
