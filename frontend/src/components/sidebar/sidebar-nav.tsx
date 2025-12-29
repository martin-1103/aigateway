import { NAV_ITEMS } from './sidebar.constants'
import { SidebarItem } from './sidebar-item'

interface SidebarNavProps {
  userRole: 'admin' | 'user' | 'provider'
  authMethod?: 'jwt' | 'access_key' | null
}

export function SidebarNav({ userRole, authMethod }: SidebarNavProps) {
  const isAccessKey = authMethod === 'access_key'

  const visibleItems = NAV_ITEMS.filter((item) => {
    // Check role permission
    if (!item.roles.includes(userRole)) return false
    // If using access key, only show items that allow it
    if (isAccessKey && item.allowAccessKey === false) return false
    return true
  })

  return (
    <nav className="flex flex-col gap-1 p-2">
      {visibleItems.map((item) => (
        <SidebarItem key={item.href} item={item} />
      ))}
    </nav>
  )
}
