import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import type { OAuthProvider } from '../oauth.types'

interface OAuthProvidersListProps {
  providers: OAuthProvider[]
  onSelectProvider: (provider: OAuthProvider) => void
  isLoading?: boolean
}

export function OAuthProvidersList({
  providers,
  onSelectProvider,
  isLoading,
}: OAuthProvidersListProps) {
  if (isLoading) {
    return (
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {[1, 2, 3].map((i) => (
          <Card key={i} className="animate-pulse">
            <CardHeader>
              <div className="h-6 w-32 bg-muted rounded" />
              <div className="h-4 w-48 bg-muted rounded" />
            </CardHeader>
            <CardContent>
              <div className="h-10 w-full bg-muted rounded" />
            </CardContent>
          </Card>
        ))}
      </div>
    )
  }

  if (!Array.isArray(providers) || providers.length === 0) {
    return (
      <Card>
        <CardContent className="py-8 text-center text-muted-foreground">
          No OAuth providers configured
        </CardContent>
      </Card>
    )
  }

  return (
    <div
      className="grid gap-4 md:grid-cols-2 lg:grid-cols-3"
      role="list"
      aria-label="OAuth providers"
    >
      {providers.map((provider) => (
        <Card key={provider.name} role="listitem">
          <CardHeader>
            <CardTitle className="text-lg">{provider.name}</CardTitle>
            <CardDescription>
              Click to start OAuth authentication
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button
              className="w-full"
              onClick={() => onSelectProvider(provider)}
              aria-label={`Initialize OAuth for ${provider.name}`}
            >
              Initialize OAuth
            </Button>
          </CardContent>
        </Card>
      ))}
    </div>
  )
}
