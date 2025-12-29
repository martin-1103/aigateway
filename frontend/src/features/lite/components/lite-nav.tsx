import { NavLink, useSearchParams } from 'react-router-dom'
import { cn } from '@/lib/utils'
import type { LiteUser } from '../api/lite-me.api'

interface LiteNavProps {
  user: LiteUser
}

export function LiteNav({ user }: LiteNavProps) {
  const [searchParams] = useSearchParams()
  const key = searchParams.get('key')

  const tabs = [
    { label: 'Accounts', path: '/lite', roles: ['admin', 'user', 'provider'] },
    { label: 'API Keys', path: '/lite/api-keys', roles: ['admin', 'user'] },
  ]

  const visibleTabs = tabs.filter((t) => t.roles.includes(user.role))

  return (
    <nav className="border-b bg-background">
      <div className="container flex gap-4">
        {visibleTabs.map((tab) => (
          <NavLink
            key={tab.path}
            to={`${tab.path}?key=${key}`}
            end={tab.path === '/lite'}
            className={({ isActive }) =>
              cn(
                'border-b-2 px-1 py-3 text-sm font-medium transition-colors',
                isActive
                  ? 'border-primary text-foreground'
                  : 'border-transparent text-muted-foreground hover:text-foreground'
              )
            }
          >
            {tab.label}
          </NavLink>
        ))}
      </div>
    </nav>
  )
}
