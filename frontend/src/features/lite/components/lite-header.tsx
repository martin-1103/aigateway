import type { LiteUser } from '../api/lite-me.api'

interface LiteHeaderProps {
  user: LiteUser
}

export function LiteHeader({ user }: LiteHeaderProps) {
  return (
    <header className="border-b bg-background">
      <div className="container flex h-14 items-center justify-between">
        <div className="flex items-center gap-2">
          <h1 className="text-lg font-semibold">AIGateway Lite</h1>
        </div>
        <div className="flex items-center gap-4">
          <span className="text-sm text-muted-foreground">
            {user.username}
          </span>
          <span className="rounded-full bg-primary/10 px-2 py-0.5 text-xs font-medium text-primary">
            {user.role}
          </span>
        </div>
      </div>
    </header>
  )
}
