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
import { useInitOAuthFlow, useExchangeOAuth } from '../hooks'
import type { OAuthProvider } from '../oauth.types'
import { ExternalLink, Loader2 } from 'lucide-react'

interface OAuthInitDialogProps {
  provider: OAuthProvider | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

type Step = 'init' | 'callback'

export function OAuthInitDialog({
  provider,
  open,
  onOpenChange,
}: OAuthInitDialogProps) {
  const [step, setStep] = useState<Step>('init')
  const [projectId, setProjectId] = useState('')
  const [callbackUrl, setCallbackUrl] = useState('')
  const [authUrl, setAuthUrl] = useState('')

  const initMutation = useInitOAuthFlow()
  const exchangeMutation = useExchangeOAuth({
    onSuccess: () => {
      handleClose()
    },
  })

  const handleInitSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!provider || !projectId.trim()) return

    try {
      const response = await initMutation.mutateAsync({
        provider: provider.id,
        project_id: projectId.trim(),
        flow_type: 'manual',
      })
      setAuthUrl(response.auth_url)
      setStep('callback')
      window.open(response.auth_url, '_blank')
    } catch {
      // Error handled by mutation
    }
  }

  const handleExchangeSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!callbackUrl.trim()) return
    exchangeMutation.mutate({ callback_url: callbackUrl.trim() })
  }

  const handleClose = () => {
    setStep('init')
    setProjectId('')
    setCallbackUrl('')
    setAuthUrl('')
    onOpenChange(false)
  }

  const handleBack = () => {
    setStep('init')
    setCallbackUrl('')
  }

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>
            {step === 'init' ? 'Initialize OAuth Flow' : 'Complete OAuth Flow'}
          </DialogTitle>
          <DialogDescription>
            {step === 'init'
              ? provider
                ? `Start OAuth authentication for ${provider.name}.`
                : 'Select a provider to continue.'
              : 'Paste the callback URL from your browser after completing Google login.'}
          </DialogDescription>
        </DialogHeader>

        {provider && step === 'init' && (
          <form onSubmit={handleInitSubmit} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="provider-name">Provider</Label>
              <Input id="provider-name" value={provider.name} disabled />
            </div>

            <div className="space-y-2">
              <Label htmlFor="project-id">Project ID</Label>
              <Input
                id="project-id"
                value={projectId}
                onChange={(e) => setProjectId(e.target.value)}
                placeholder="e.g., my-project, project-123"
                required
              />
            </div>

            <DialogFooter>
              <Button type="button" variant="outline" onClick={handleClose}>
                Cancel
              </Button>
              <Button
                type="submit"
                disabled={initMutation.isPending || !projectId.trim()}
              >
                {initMutation.isPending ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Starting...
                  </>
                ) : (
                  <>
                    <ExternalLink className="mr-2 h-4 w-4" />
                    Start OAuth Flow
                  </>
                )}
              </Button>
            </DialogFooter>
          </form>
        )}

        {provider && step === 'callback' && (
          <form onSubmit={handleExchangeSubmit} className="space-y-4">
            <div className="rounded-md bg-muted p-3 text-sm">
              <p className="font-medium">Instructions:</p>
              <ol className="mt-2 list-decimal pl-4 space-y-1 text-muted-foreground">
                <li>Complete the Google login in the new tab</li>
                <li>After login, you'll be redirected to a URL starting with <code className="text-xs bg-background px-1 rounded">http://localhost:8088/...</code></li>
                <li>Copy the entire URL from your browser's address bar</li>
                <li>Paste it below and click Submit</li>
              </ol>
            </div>

            {authUrl && (
              <div className="space-y-2">
                <Label>Auth URL (already opened)</Label>
                <div className="flex gap-2">
                  <Input value={authUrl} disabled className="text-xs" />
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => window.open(authUrl, '_blank')}
                  >
                    <ExternalLink className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            )}

            <div className="space-y-2">
              <Label htmlFor="callback-url">Callback URL</Label>
              <textarea
                id="callback-url"
                value={callbackUrl}
                onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setCallbackUrl(e.target.value)}
                placeholder="Paste the full callback URL here..."
                rows={3}
                required
                className="flex w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
              />
            </div>

            <DialogFooter>
              <Button type="button" variant="outline" onClick={handleBack}>
                Back
              </Button>
              <Button
                type="submit"
                disabled={exchangeMutation.isPending || !callbackUrl.trim()}
              >
                {exchangeMutation.isPending ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Exchanging...
                  </>
                ) : (
                  'Submit Callback URL'
                )}
              </Button>
            </DialogFooter>
          </form>
        )}
      </DialogContent>
    </Dialog>
  )
}
