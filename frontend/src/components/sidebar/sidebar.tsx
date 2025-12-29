import { ChevronLeft, ChevronRight } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { useSidebarStore } from './sidebar.store'
import { SidebarNav } from './sidebar-nav'

interface SidebarProps {
  userRole: 'admin' | 'user' | 'provider'
  authMethod?: 'jwt' | 'access_key' | null
}

export function Sidebar({ userRole, authMethod }: SidebarProps) {
  const { isCollapsed, toggle } = useSidebarStore()

  return (
    <aside
      className={cn(
        'flex flex-col border-r bg-card transition-all duration-300',
        isCollapsed ? 'w-16' : 'w-64'
      )}
    >
      <div className="flex h-14 items-center justify-between border-b px-4">
        {!isCollapsed && <span className="font-semibold">AIGateway</span>}
        <Button variant="ghost" size="icon" onClick={toggle} className={cn(isCollapsed && 'mx-auto')}>
          {isCollapsed ? <ChevronRight className="h-4 w-4" /> : <ChevronLeft className="h-4 w-4" />}
        </Button>
      </div>
      <SidebarNav userRole={userRole} authMethod={authMethod} />
    </aside>
  )
}
