import { useState } from 'react'
import { OAuthProvidersList, OAuthInitDialog } from './components'
import { useOAuthProviders } from './hooks'
import type { OAuthProvider } from './oauth.types'

export function OAuthPage() {
  const { data, isLoading, error } = useOAuthProviders()
  const [selectedProvider, setSelectedProvider] = useState<OAuthProvider | null>(
    null
  )
  const [dialogOpen, setDialogOpen] = useState(false)

  const handleSelectProvider = (provider: OAuthProvider) => {
    setSelectedProvider(provider)
    setDialogOpen(true)
  }

  if (error) {
    return (
      <div className="space-y-6">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">OAuth Providers</h1>
          <p className="text-muted-foreground">
            Manage OAuth authentication for AI providers
          </p>
        </div>
        <div className="rounded-md border border-destructive/50 bg-destructive/10 p-4 text-destructive">
          Failed to load OAuth providers. Please try again later.
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">OAuth Providers</h1>
        <p className="text-muted-foreground">
          Manage OAuth authentication for AI providers
        </p>
      </div>

      <OAuthProvidersList
        providers={data?.providers ?? []}
        onSelectProvider={handleSelectProvider}
        isLoading={isLoading}
      />

      <OAuthInitDialog
        provider={selectedProvider}
        open={dialogOpen}
        onOpenChange={setDialogOpen}
      />
    </div>
  )
}
