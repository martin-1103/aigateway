import { useState } from 'react'
import { Plus } from 'lucide-react'
import { PageHeader, PageContent } from '@/components/page'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { useProxiesQuery } from './hooks'
import {
  ProxiesTable,
  ProxyCreateDialog,
  ProxyEditDialog,
  ProxyDeleteDialog,
} from './components'
import type { Proxy } from './proxies.types'

export function ProxiesPage() {
  const { data, isLoading } = useProxiesQuery()
  const [createOpen, setCreateOpen] = useState(false)
  const [editProxy, setEditProxy] = useState<Proxy | null>(null)
  const [deleteProxy, setDeleteProxy] = useState<Proxy | null>(null)

  const handleEdit = (proxy: Proxy) => setEditProxy(proxy)
  const handleDelete = (proxy: Proxy) => setDeleteProxy(proxy)

  return (
    <>
      <PageHeader
        title="Proxies"
        description="Manage proxy servers for API routing."
        actions={
          <Button onClick={() => setCreateOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Add Proxy
          </Button>
        }
      />
      <PageContent>
        <Card>
          <ProxiesTable
            proxies={data?.data ?? []}
            isLoading={isLoading}
            onEdit={handleEdit}
            onDelete={handleDelete}
          />
        </Card>
      </PageContent>

      <ProxyCreateDialog open={createOpen} onOpenChange={setCreateOpen} />
      <ProxyEditDialog
        proxy={editProxy}
        open={!!editProxy}
        onOpenChange={(open) => !open && setEditProxy(null)}
      />
      <ProxyDeleteDialog
        proxy={deleteProxy}
        open={!!deleteProxy}
        onOpenChange={(open) => !open && setDeleteProxy(null)}
      />
    </>
  )
}
