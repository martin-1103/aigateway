import { NAV_ITEMS } from './sidebar.constants'
import { SidebarItem } from './sidebar-item'

interface SidebarNavProps {
  userRole: 'admin' | 'user' | 'provider'
}

export function SidebarNav({ userRole }: SidebarNavProps) {
  const visibleItems = NAV_ITEMS.filter(
    (item) => item.roles.includes(userRole)
  )

  return (
    <nav className="flex flex-col gap-1 p-2">
      {visibleItems.map((item) => (
        <SidebarItem key={item.href} item={item} />
      ))}
    </nav>
  )
}
