import { Outlet } from 'react-router-dom'
import { useLiteAuth } from './hooks/use-lite-auth'
import { LiteHeader } from './components/lite-header'
import { LiteNav } from './components/lite-nav'
import { LoadingSpinner } from '@/components/feedback'

export function LiteLayout() {
  const { user, isLoading, error, accessKey } = useLiteAuth()

  if (isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background">
        <LoadingSpinner />
      </div>
    )
  }

  if (error || !user) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-destructive">Access Denied</h1>
          <p className="mt-2 text-muted-foreground">
            {error || 'Invalid or missing access key'}
          </p>
          <p className="mt-4 text-sm text-muted-foreground">
            Please use a valid access key URL:
            <br />
            <code className="mt-1 block rounded bg-muted px-2 py-1 text-xs">
              /lite?key=uk_your_access_key
            </code>
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-background">
      <LiteHeader user={user} />
      <LiteNav user={user} />
      <main className="container py-6">
        <Outlet context={{ user, accessKey }} />
      </main>
    </div>
  )
}
