import { Database, Key, Server, Users } from 'lucide-react'
import { PageHeader, PageContent } from '@/components/page'
import { StatCard } from '@/components/charts'
import { useAuthStore } from '@/features/auth'

export function DashboardPage() {
  const user = useAuthStore((s) => s.user)

  return (
    <>
      <PageHeader
        title="Dashboard"
        description={`Welcome back, ${user?.username}`}
      />
      <PageContent>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          <StatCard title="Users" value="--" icon={Users} />
          <StatCard title="Accounts" value="--" icon={Database} />
          <StatCard title="Proxies" value="--" icon={Server} />
          <StatCard title="API Keys" value="--" icon={Key} />
        </div>
      </PageContent>
    </>
  )
}
