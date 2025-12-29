import { HeaderThemeToggle } from './header-theme-toggle'
import { HeaderUserMenu } from './header-user-menu'

interface HeaderProps {
  username: string
  role: string
  onLogout: () => void
  authMethod?: 'jwt' | 'access_key' | null
}

export function Header({ username, role, onLogout, authMethod }: HeaderProps) {
  return (
    <header className="flex h-14 items-center justify-end gap-2 border-b bg-card px-4">
      {authMethod === 'access_key' && (
        <span className="text-xs text-muted-foreground bg-muted px-2 py-1 rounded">
          Access Key Mode
        </span>
      )}
      <HeaderThemeToggle />
      <HeaderUserMenu username={username} role={role} onLogout={onLogout} />
    </header>
  )
}
