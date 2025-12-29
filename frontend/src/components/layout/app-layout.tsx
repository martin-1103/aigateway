import { Outlet } from 'react-router-dom'
import { Sidebar } from '@/components/sidebar'
import { Header } from '@/components/header'

interface AppLayoutProps {
  username: string
  role: 'admin' | 'user' | 'provider'
  onLogout: () => void
  authMethod?: 'jwt' | 'access_key' | null
}

export function AppLayout({ username, role, onLogout, authMethod }: AppLayoutProps) {
  return (
    <div className="flex h-screen bg-background">
      <Sidebar userRole={role} authMethod={authMethod} />
      <div className="flex flex-1 flex-col overflow-hidden">
        <Header username={username} role={role} onLogout={onLogout} authMethod={authMethod} />
        <main className="flex-1 overflow-auto p-6">
          <Outlet />
        </main>
      </div>
    </div>
  )
}
