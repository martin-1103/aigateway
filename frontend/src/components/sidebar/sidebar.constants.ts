import { BarChart3, Database, GitBranch, Key, Server, Settings, Users } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'

export type Role = 'admin' | 'user' | 'provider'

export interface NavItem {
  label: string
  href: string
  icon: LucideIcon
  roles: Role[]
  allowAccessKey?: boolean
}

export const NAV_ITEMS: NavItem[] = [
  { label: 'Dashboard', href: '/dashboard', icon: BarChart3, roles: ['admin', 'user', 'provider'], allowAccessKey: true },
  { label: 'Users', href: '/users', icon: Users, roles: ['admin'], allowAccessKey: false },
  { label: 'Accounts', href: '/accounts', icon: Database, roles: ['admin', 'user', 'provider'], allowAccessKey: true },
  { label: 'Proxies', href: '/proxies', icon: Server, roles: ['admin'], allowAccessKey: false },
  { label: 'API Keys', href: '/api-keys', icon: Key, roles: ['admin', 'user'], allowAccessKey: false },
  { label: 'Stats', href: '/stats', icon: BarChart3, roles: ['admin', 'user'], allowAccessKey: true },
  { label: 'Model Mappings', href: '/model-mappings', icon: GitBranch, roles: ['admin', 'user'], allowAccessKey: true },
  { label: 'Settings', href: '/settings', icon: Settings, roles: ['admin', 'user', 'provider'], allowAccessKey: false },
]
