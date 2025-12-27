import { HeaderThemeToggle } from './header-theme-toggle'
import { HeaderUserMenu } from './header-user-menu'

interface HeaderProps {
  username: string
  role: string
  onLogout: () => void
}

export function Header({ username, role, onLogout }: HeaderProps) {
  return (
    <header className="flex h-14 items-center justify-end gap-2 border-b bg-card px-4">
      <HeaderThemeToggle />
      <HeaderUserMenu username={username} role={role} onLogout={onLogout} />
    </header>
  )
}
