import { useState, useEffect } from 'react'
import { Copy, RefreshCw, ExternalLink } from 'lucide-react'
import { getMyAccessKey, getMyFullAccessKey, regenerateAccessKey } from './api/access-key.api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { LoadingSpinner } from '@/components/feedback'
import { toast } from 'sonner'

export function SettingsPage() {
  const [maskedKey, setMaskedKey] = useState<string>('')
  const [newKey, setNewKey] = useState<string>('')
  const [isLoading, setIsLoading] = useState(true)
  const [isRegenerating, setIsRegenerating] = useState(false)
  const [showConfirmDialog, setShowConfirmDialog] = useState(false)
  const [showNewKeyDialog, setShowNewKeyDialog] = useState(false)

  useEffect(() => {
    getMyAccessKey()
      .then(setMaskedKey)
      .catch(() => setMaskedKey(''))
      .finally(() => setIsLoading(false))
  }, [])

  const handleRegenerate = async () => {
    setShowConfirmDialog(false)
    setIsRegenerating(true)
    try {
      const key = await regenerateAccessKey()
      setNewKey(key)
      setShowNewKeyDialog(true)
      // Refresh masked key
      getMyAccessKey().then(setMaskedKey)
    } catch (err) {
      toast.error('Failed to regenerate access key')
    } finally {
      setIsRegenerating(false)
    }
  }

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text)
    toast.success('Copied to clipboard')
  }

  const handleOpenWithAccessKey = async () => {
    try {
      const fullKey = await getMyFullAccessKey()
      window.open(`/?key=${fullKey}`, '_blank')
    } catch {
      toast.error('Failed to get access key')
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold tracking-tight">Settings</h2>
        <p className="text-muted-foreground">
          Manage your account settings
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Access Key</CardTitle>
          <CardDescription>
            Use this key to access the panel without logging in.
            Useful for quickly adding OAuth accounts from different browsers.
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          {isLoading ? (
            <LoadingSpinner />
          ) : (
            <>
              <div className="flex items-center gap-2">
                <Input
                  value={maskedKey || 'No key generated'}
                  readOnly
                  className="font-mono"
                />
                <Button
                  variant="outline"
                  size="icon"
                  onClick={() => setShowConfirmDialog(true)}
                  disabled={isRegenerating}
                >
                  {isRegenerating ? (
                    <LoadingSpinner className="h-4 w-4" />
                  ) : (
                    <RefreshCw className="h-4 w-4" />
                  )}
                </Button>
              </div>

              {maskedKey && (
                <div className="rounded-md bg-muted p-4">
                  <p className="text-sm font-medium">Access Key URL:</p>
                  <p className="mt-1 text-xs text-muted-foreground">
                    Bookmark this URL to access your accounts without logging in.
                    Some features are disabled in Access Key mode for security.
                  </p>
                  <div className="mt-2 flex items-center gap-2">
                    <code className="flex-1 rounded bg-background px-2 py-1 text-xs">
                      {window.location.origin}/?key=...
                    </code>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={handleOpenWithAccessKey}
                    >
                      <ExternalLink className="mr-1 h-3 w-3" />
                      Open Panel
                    </Button>
                  </div>
                </div>
              )}
            </>
          )}
        </CardContent>
      </Card>

      {/* Confirm Regenerate Dialog */}
      <Dialog open={showConfirmDialog} onOpenChange={setShowConfirmDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Regenerate Access Key?</DialogTitle>
            <DialogDescription>
              This will invalidate your current access key. Any bookmarked
              URLs will stop working until you update them with the new key.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setShowConfirmDialog(false)}>
              Cancel
            </Button>
            <Button onClick={handleRegenerate}>Regenerate</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* New Key Dialog */}
      <Dialog open={showNewKeyDialog} onOpenChange={setShowNewKeyDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>New Access Key Generated</DialogTitle>
            <DialogDescription>
              Save this key now. You won't be able to see it again!
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="flex items-center gap-2">
              <Input value={newKey} readOnly className="font-mono text-sm" />
              <Button
                variant="outline"
                size="icon"
                onClick={() => copyToClipboard(newKey)}
              >
                <Copy className="h-4 w-4" />
              </Button>
            </div>
            <div className="rounded-md bg-muted p-3">
              <p className="text-sm font-medium">Your Access Key URL:</p>
              <code className="mt-1 block text-xs break-all">
                {window.location.origin}/?key={newKey}
              </code>
              <Button
                variant="link"
                size="sm"
                className="mt-2 h-auto p-0"
                onClick={() => copyToClipboard(`${window.location.origin}/?key=${newKey}`)}
              >
                <Copy className="mr-1 h-3 w-3" />
                Copy Full URL
              </Button>
            </div>
          </div>
          <DialogFooter>
            <Button onClick={() => setShowNewKeyDialog(false)}>Done</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
