import { NavLink } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { useSidebarStore } from './sidebar.store'
import type { NavItem } from './sidebar.constants'

interface SidebarItemProps {
  item: NavItem
}

export function SidebarItem({ item }: SidebarItemProps) {
  const isCollapsed = useSidebarStore((s) => s.isCollapsed)
  const Icon = item.icon

  return (
    <NavLink
      to={item.href}
      className={({ isActive }) =>
        cn(
          'flex items-center gap-3 rounded-lg px-3 py-2 transition-colors',
          'hover:bg-accent hover:text-accent-foreground',
          isActive ? 'bg-accent text-accent-foreground' : 'text-muted-foreground'
        )
      }
    >
      <Icon className="h-4 w-4 shrink-0" />
      {!isCollapsed && <span>{item.label}</span>}
    </NavLink>
  )
}
