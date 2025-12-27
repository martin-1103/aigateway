import { useState } from 'react'
import { Plus } from 'lucide-react'
import { PageHeader, PageContent } from '@/components/page'
import { Button } from '@/components/ui/button'
import { ApiKeysTable, ApiKeyCreateDialog } from './components'

export function ApiKeysPage() {
  const [createDialogOpen, setCreateDialogOpen] = useState(false)

  return (
    <>
      <PageHeader
        title="API Keys"
        description="Manage your API keys for accessing the gateway"
        actions={
          <Button onClick={() => setCreateDialogOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Create API Key
          </Button>
        }
      />
      <PageContent>
        <ApiKeysTable />
      </PageContent>
      <ApiKeyCreateDialog
        open={createDialogOpen}
        onClose={() => setCreateDialogOpen(false)}
      />
    </>
  )
}
